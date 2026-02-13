package summary

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"main/lib/logger"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

const (
	// LLM Configuration
	llmTimeout       = 90 * time.Second
	maxRetries       = 3
	openAIModel      = "gpt-4.1"
	openAITimeout    = llmTimeout

	// OpenAI API Configuration
	openAITemperature     = 0.6
	openAIMaxOutputTokens = 4096
	openAITopP           = 0.95
	openAIStore          = true

	// BM25 Configuration
	bm25K1        = 1.2
	bm25B         = 0.75
	bm25MinScore  = 0.1
	bm25ShortTitleMinScore = 0.05
	bm25MediumTitleMinScore = 0.08
	bm25CommonTermThreshold = 0.8
	bm25ShortTitleCommonTermThreshold = 0.9
	bm25FallbackSimilarityThreshold = 0.3

	// Validation Limits
	maxHeadlineLength    = 200
	maxSummaryWords     = 1000
	minSummaryWords     = 50
	warningSummaryWords = 800

	// Text Processing
	maxTitleLengthForShort = 3
	maxTitleLengthForMedium = 5

	// API Endpoints
	openAPIURL = "https://api.openai.com/v1/responses"
)

// Core data types
type Paper struct {
	Title    string
	URL      string
	Abstract string
	PubDate  time.Time
}

// HuggingFace API response structures - the actual format is an array of items, not nested in "papers"
type DailyPapersResponse []DailyPaperItem

type DailyPaperItem struct {
	Paper       DailyPaper `json:"paper"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	PublishedAt time.Time  `json:"publishedAt"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
}

type DailyPaper struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Authors     []Author  `json:"authors"`
	PublishedAt time.Time `json:"publishedAt"`
}

type PaperDetailsResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Authors     []Author  `json:"authors"`
	PublishedAt time.Time `json:"publishedAt"`
	URL         string    `json:"url,omitempty"`
}

type Author struct {
	ID     string `json:"_id"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`
}

type GetSummaryRawResult struct {
	Data    []byte
	Source  string
	BlobURL *string // Optional: URL for client to fetch directly
}

// Validation types
type ValidationSeverity int

const (
	SeverityWarning ValidationSeverity = iota
	SeverityError
)

type ValidationError struct {
	Field    string
	Message  string
	Details  string
	Severity ValidationSeverity
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation %s in %s: %s", e.severityString(), e.Field, e.Message)
}

func (e ValidationError) severityString() string {
	if e.Severity == SeverityWarning {
		return "warning"
	}
	return "error"
}

// OpenAI API structures
type OpenAIRequest struct {
	Model           string          `json:"model"`
	Input           []OpenAIMessage `json:"input"`
	Text            OpenAIText      `json:"text"`
	Reasoning       map[string]any  `json:"reasoning"` // Empty object for now
	Tools           []any           `json:"tools"`     // Empty array for now
	Temperature     float64         `json:"temperature"`
	MaxOutputTokens int             `json:"max_output_tokens"`
	TopP            float64         `json:"top_p"`
	Store           bool            `json:"store"`
}

type OpenAIMessage struct {
	Role    string               `json:"role"`
	Content []OpenAIContentBlock `json:"content"`
}

type OpenAIContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type OpenAIText struct {
	Format OpenAIFormat `json:"format"`
}

type OpenAIFormat struct {
	Type string `json:"type"`
}

// OpenAIResponse represents the top-level response object from /v1/responses
type OpenAIResponse struct {
	ID     string                `json:"id"`
	Object string                `json:"object"`
	Model  string                `json:"model"`
	Output []OpenAIOutputMessage `json:"output"` // Added Output field
	// Other top-level fields like status, usage, etc., can be added if needed
}

// OpenAIOutputMessage represents the message object within the 'output' array
type OpenAIOutputMessage struct {
	ID      string                  `json:"id"`
	Type    string                  `json:"type"`
	Role    string                  `json:"role"`    // Role is here
	Content []OpenAIResponseContent `json:"content"` // Content is here
}

type OpenAIResponseContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
	// Annotations field omitted as it's not needed for extraction
}

// summarizeWithLLM summarizes the markdown content using the OpenAI API
func summarizeWithLLM(ctx context.Context, markdownContent string, feedURLs map[string]string) (string, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := summarizeWithLLMAttempt(ctx, markdownContent, feedURLs, attempt)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Check if this is a retryable validation error (duplications or link issues)
		if isRetryableValidationError(err) && attempt < maxRetries {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				slog.Warn("LLM summary validation failed due to duplications, retrying",
					"attempt", attempt,
					"error", err)
			} else {
				slog.Warn("LLM summary validation failed due to link formatting, retrying",
					"attempt", attempt,
					"error", err)
			}
			continue
		}

		slog.Warn("LLM validation failed", "attempt", attempt, "error", err)
	}
	return "", lastErr
}

// isRetryableValidationError checks if the error should trigger a retry
func isRetryableValidationError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a validation error with retryable fields
	var validationErr ValidationError
	if errors.As(err, &validationErr) {
		// Retry for duplications and link formatting errors
		return validationErr.Field == "duplicates" || validationErr.Field == "links"
	}

	// Check error message for retryable keywords
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate") ||
		   strings.Contains(errMsg, "duplication") ||
		   strings.Contains(errMsg, "link") ||
		   strings.Contains(errMsg, "url")
}

// summarizeWithLLMAttempt performs a single LLM summarization attempt
func summarizeWithLLMAttempt(ctx context.Context, markdownContent string, feedURLs map[string]string, attempt int) (string, error) {
	apiURL := openAPIURL
	apiKey := os.Getenv("OPENAI_API_KEY")

	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Construct the exact prompt as requested
	basePrompt := `Create a brief morning briefing on these AI research papers, written in a conversational style for busy professionals. Focus on what's new and what it means for businesses and society.
Format the output in markdown:
## Morning Headline
(1 sentence, 15 words or less)
## What's New
(2–3 sentences total, written like you're explaining it to a friend over coffee. Use multiple paragraphs.)

 - Cover all papers in a natural, flowing narrative
 - Group related papers together
 - Include key metrics and outcomes
 - Keep the tone light and engaging

Important: When referring to a paper, write if possible a short title inside square brackets like [Paper Title] and DO NOT include URLs anywhere in the output. Links will be added automatically. Every paper reference must be a complete markdown link with both text and URL.
Keep it under 200 words. Start with the most impressive or important paper. Focus on outcomes and implications, not technical details. Do not write a word count.
Do not enclose in a markdown code block, just return the markdown.`

	// Add stronger emphasis for retry attempts
	if attempt > 1 {
		basePrompt += `

CRITICAL: Do NOT repeat the same paper title or short title anywhere in the summary. Each paper should only be mentioned once. If you mention a paper, do not reference it again later in the text.
CRITICAL: Every paper reference must be a complete markdown link with both text and URL like [Paper Title](URL). Do NOT create links without URLs. This is extremely important - previous attempts had formatting errors and were rejected.`
	} else {
		basePrompt += `

CRITICAL: Do NOT repeat the same paper title or short title anywhere in the summary. Each paper should only be mentioned once. If you mention a paper, do not reference it again later in the text.
CRITICAL: Every paper reference must be a complete markdown link with both text and URL like [Paper Title](URL). Do NOT create links without URLs.`
	}

	basePrompt += `
Below are the paper abstracts and information in markdown format:`

	promptText := basePrompt + markdownContent

	// Construct the OpenAI request body
	request := OpenAIRequest{
		Model: openAIModel,
		Input: []OpenAIMessage{
			{
				Role: "user",
				Content: []OpenAIContentBlock{
					{
						Type: "input_text",
						Text: promptText,
					},
				},
			},
		},
		Text: OpenAIText{
			Format: OpenAIFormat{
				Type: "text",
			},
		},
		Reasoning:       make(map[string]any), // Empty object
		Tools:           make([]any, 0),       // Empty array
		Temperature:     openAITemperature,
		MaxOutputTokens: openAIMaxOutputTokens,
		TopP:            openAITopP,
		Store:           openAIStore,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create OpenAI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey) // Use OpenAI key

	// Create an HTTP client with the LLM timeout
	client := &http.Client{
		Timeout: openAITimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", fmt.Errorf("timeout calling OpenAI API: %w", err)
		}
		return "", fmt.Errorf("failed to send request to OpenAI API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		slog.Error("Failed to read OpenAI response body", "error", readErr)
		// Return specific error about reading the body, but include original status code if not OK
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP error %d from OpenAI API and failed to read body: %w", resp.StatusCode, readErr)
		}
		return "", fmt.Errorf("failed to read OpenAI response body: %w", readErr)
	}

	// Log the raw response body for debugging
	// slog.Info("Raw OpenAI API Response Body", "status_code", resp.StatusCode, "body", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		// We already logged the body, just return the error
		return "", fmt.Errorf("HTTP error %d from OpenAI API: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the single OpenAI response object from the read bytes
	var openAIResp OpenAIResponse                                  // Decode into the struct, not a slice
	if err := json.Unmarshal(bodyBytes, &openAIResp); err != nil { // Use json.Unmarshal with the byte slice
		// Log the body again specifically on decode error
		slog.Error("Failed to decode OpenAI response JSON", "error", err, "raw_body", string(bodyBytes))
		return "", fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	// Extract the text content from the nested structure
	if len(openAIResp.Output) == 0 || openAIResp.Output[0].Role != "assistant" || len(openAIResp.Output[0].Content) == 0 || openAIResp.Output[0].Content[0].Type != "output_text" {
		// Log the parsed struct for better debugging if validation fails
		slog.Warn("OpenAI response structure unexpected or empty after parsing", "parsedResponse", openAIResp)
		return "", fmt.Errorf("invalid or empty response structure from OpenAI API")
	}

	// Extract the markdown text directly from the nested path
	markdownSummary := openAIResp.Output[0].Content[0].Text

	// Sanitize any raw URLs and programmatically inject links from the feed
	sanitized := sanitizeSummaryMarkdown(markdownSummary)
	// Apply a conservative headline length clamp to avoid occasional LLM overflow
	sanitized = enforceHeadlineLength(sanitized, maxHeadlineLength)
	linkedMarkdown := replacePlaceholdersWithLinks(sanitized, feedURLs)

	// Validate the linked summary content
	if err := validateSummaryContent(linkedMarkdown, feedURLs); err != nil {
		slog.Error("LLM summary validation failed",
			"error", err,
			"summary", linkedMarkdown)
		return "", fmt.Errorf("LLM summary validation failed: %w", err)
	}

	slog.Info("Successfully validated LLM summary",
		"summary_length", len(linkedMarkdown),
		"link_count", len(regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).FindAllString(linkedMarkdown, -1)))

	return linkedMarkdown, nil
}

// sanitizeSummaryMarkdown removes raw URLs that the LLM may include
func sanitizeSummaryMarkdown(input string) string {
	// Remove URLs in parentheses
	reParen := regexp.MustCompile(`\s+\((https?://[^)]+)\)`)
	cleaned := reParen.ReplaceAllString(input, "")

	// Remove angle-bracket autolinks
	reAngle := regexp.MustCompile(`<https?://[^>]+>`)
	cleaned = reAngle.ReplaceAllString(cleaned, "")

	// Remove bare URLs
	reBare := regexp.MustCompile(`\s+https?://\S+`)
	cleaned = reBare.ReplaceAllString(cleaned, "")

	// Collapse repeated spaces
	cleaned = regexp.MustCompile(`[ \t]{2,}`).ReplaceAllString(cleaned, " ")
	return cleaned
}

// enforceHeadlineLength truncates the content under "## Morning Headline" to maxHeadlineLength,
// attempting to cut on sentence or word boundaries and appending an ellipsis if truncated.
func enforceHeadlineLength(markdown string, maxChars int) string {
	// 1) Find the Morning Headline heading line
	reHeadline := regexp.MustCompile(`(?m)^##\s*Morning Headline\s*$`)
	headlineIdx := reHeadline.FindStringIndex(markdown)
	if headlineIdx == nil {
		return markdown
	}

	// Content starts after the end of the headline line
	contentStart := headlineIdx[1]
	// Skip following newlines
	for contentStart < len(markdown) && (markdown[contentStart] == '\n' || markdown[contentStart] == '\r') {
		contentStart++
	}

	// 2) Find the next H2 after Morning Headline to determine section end
	reNextH2 := regexp.MustCompile(`(?m)^##\s+`)
	nextIdx := reNextH2.FindStringIndex(markdown[contentStart:])
	contentEnd := len(markdown)
	if nextIdx != nil {
		contentEnd = contentStart + nextIdx[0]
	}

	if contentStart >= contentEnd {
		return markdown
	}

	sectionContent := markdown[contentStart:contentEnd]

	// Consider only the first paragraph for the headline
	paraEnd := strings.Index(sectionContent, "\n\n")
	var paragraph string
	if paraEnd == -1 {
		paragraph = strings.TrimSpace(sectionContent)
	} else {
		paragraph = strings.TrimSpace(sectionContent[:paraEnd])
	}

	// If already within limit, leave as-is but ensure only first paragraph remains in section
	runeParagraph := []rune(paragraph)
	if len(runeParagraph) <= maxChars {
		rebuilt := markdown[:contentStart] + paragraph + "\n\n" + markdown[contentEnd:]
		return rebuilt
	}

	// Truncate at sentence boundary if possible, else at last space
	cutoff := maxChars
	if cutoff > len(runeParagraph) {
		cutoff = len(runeParagraph)
	}
	candidate := string(runeParagraph[:cutoff])
	tail := candidate
	if len(candidate) > 40 {
		tail = candidate[len(candidate)-40:]
	}
	lastPunct := -1
	for i := len(tail) - 1; i >= 0; i-- {
		switch tail[i] {
		case '.', '!', '?':
			lastPunct = len(candidate) - (len(tail) - i)
			i = -1
		}
	}
	if lastPunct != -1 && lastPunct > maxChars/2 {
		candidate = candidate[:lastPunct+1]
	} else {
		lastSpace := strings.LastIndex(candidate, " ")
		if lastSpace > maxChars/2 {
			candidate = candidate[:lastSpace]
		}
		candidate = strings.TrimRight(candidate, " ") + "…"
	}

	rebuilt := markdown[:contentStart] + strings.TrimSpace(candidate) + "\n\n" + markdown[contentEnd:]
	return rebuilt
}

// deriveArxivIDFromURL tries to extract an arXiv-style ID from known sources
// like Hugging Face papers pages (e.g., https://huggingface.co/papers/2508.03694)
// or arXiv links (e.g., https://arxiv.org/abs/2508.03694).
func deriveArxivIDFromURL(url string) string {
	normalizedURL := normalizeURLForPathExtraction(url)
	// Take last path segment
	lastSlash := strings.LastIndex(normalizedURL, "/")
	if lastSlash == -1 || lastSlash+1 >= len(normalizedURL) {
		return ""
	}
	segment := normalizedURL[lastSlash+1:]
	// Basic sanity: allow formats like 2508.03694 or 2508.0369x (rare extensions)
	re := regexp.MustCompile(`^[0-9]{4}\.[0-9]{4,5}[a-zA-Z0-9-]*$`)
	if re.MatchString(segment) {
		return segment
	}
	return ""
}

// toTLDRLink rewrites a paper URL to the unified TLDR route if an arXiv ID is found.
func toTLDRLink(url string) string {
	if id := deriveArxivIDFromURL(url); id != "" {
		return "https://tldr.takara.ai/p/" + id
	}
	return url
}

// replacePlaceholdersWithLinks replaces [Title] placeholders with markdown links
func replacePlaceholdersWithLinks(summaryMarkdown string, links map[string]string) string {
	// Pre-build BM25 index to reuse for multiple placeholder lookups
	bm25 := NewBM25(links)

	var builder strings.Builder
	currentIndex := 0
	for currentIndex < len(summaryMarkdown) {
		if summaryMarkdown[currentIndex] == '[' {
			// Find the closing bracket, but be defensive about bounds
			closeIdx := strings.IndexByte(summaryMarkdown[currentIndex+1:], ']')
			if closeIdx == -1 {
				// No closing bracket found, just write the rest and break
				builder.WriteString(summaryMarkdown[currentIndex:])
				break
			}
			closeIdx += currentIndex + 1

			// Skip if already a link [title](url)
			if closeIdx+1 < len(summaryMarkdown) && summaryMarkdown[closeIdx+1] == '(' {
				endParen := strings.IndexByte(summaryMarkdown[closeIdx+2:], ')')
				if endParen != -1 {
					builder.WriteString(summaryMarkdown[currentIndex : closeIdx+2+endParen+1])
					currentIndex = closeIdx + 2 + endParen + 1
					continue
				}
			}

			title := strings.TrimSpace(summaryMarkdown[currentIndex+1 : closeIdx])
			// Skip empty titles or titles that are just punctuation
			if title == "" || regexp.MustCompile(`^[[:punct:]]+$`).MatchString(title) {
				builder.WriteString(summaryMarkdown[currentIndex : closeIdx+1])
				currentIndex = closeIdx + 1
				continue
			}

			if url := findMatchingURLWithBM25(title, bm25); url != "" {
				builder.WriteString(fmt.Sprintf("[%s](%s)", title, toTLDRLink(url)))
				currentIndex = closeIdx + 1
				continue
			} else {
				// Log when we can't find a match for debugging
				slog.Debug("Could not find matching URL for title", "title", title)
			}
		}
		builder.WriteByte(summaryMarkdown[currentIndex])
		currentIndex++
	}
	return builder.String()
}

// BM25Config holds BM25 scoring parameters
type BM25Config struct {
	K1                          float64
	B                           float64
	MinScore                    float64
	ShortTitleMinScore          float64
	MediumTitleMinScore         float64
	CommonTermThreshold         float64
	ShortTitleCommonTermThreshold float64
	FallbackSimilarityThreshold float64
}

// DefaultBM25Config returns the default BM25 configuration
func DefaultBM25Config() BM25Config {
	return BM25Config{
		K1:                          bm25K1,
		B:                           bm25B,
		MinScore:                    bm25MinScore,
		ShortTitleMinScore:          bm25ShortTitleMinScore,
		MediumTitleMinScore:         bm25MediumTitleMinScore,
		CommonTermThreshold:         bm25CommonTermThreshold,
		ShortTitleCommonTermThreshold: bm25ShortTitleCommonTermThreshold,
		FallbackSimilarityThreshold: bm25FallbackSimilarityThreshold,
	}
}

// BM25 holds the corpus and scoring parameters for BM25 ranking
type BM25 struct {
	Docs      [][]string        // Tokenized documents
	DocFreq   map[string]int    // Term frequency across all documents
	AvgDocLen float64           // Average document length
	Config    BM25Config        // BM25 scoring configuration
	Titles    map[string]string // docID -> original title
	URLs      map[string]string // docID -> URL
}

// NewBM25 creates a new BM25 instance from paper titles with proper normalization
func NewBM25(links map[string]string) *BM25 {
	logger.Debug("Building BM25 index", map[string]interface{}{
		"num_titles": len(links),
	})

	// Pre-allocate slices and maps for better performance
	docs := make([][]string, 0, len(links))
	titles := make(map[string]string)
	urls := make(map[string]string)

	// Process documents in parallel for better performance with large feeds
	type docResult struct {
		docID   int
		title   string
		url     string
		tokens  []string
		tokenLen int
		df      map[string]int
	}

	results := make(chan docResult, len(links))
	var wg sync.WaitGroup

	// Process each document in parallel
	docID := 0
	for title, url := range links {
		wg.Add(1)
		go func(id int, t, u string) {
			defer wg.Done()

			// Use the same normalization for documents as queries
			tokens := normalizeText(t)

			// Build document frequency for this document only
			localDF := make(map[string]int)
			seen := make(map[string]bool)
			for _, token := range tokens {
				if !seen[token] {
					localDF[token] = 1
					seen[token] = true
				}
			}

			results <- docResult{
				docID:    id,
				title:    t,
				url:      u,
				tokens:   tokens,
				tokenLen: len(tokens),
				df:       localDF,
			}
		}(docID, title, url)
		docID++
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and merge document frequencies
	var totalLen float64
	globalDF := make(map[string]int)

	for result := range results {
		titles[strconv.Itoa(result.docID)] = result.title
		urls[strconv.Itoa(result.docID)] = result.url
		docs = append(docs, result.tokens)
		totalLen += float64(result.tokenLen)

		// Merge document frequencies
		for token, count := range result.df {
			globalDF[token] += count
		}
	}

	avgDocLen := totalLen / float64(len(docs))
	if len(docs) == 0 {
		avgDocLen = 0
	}

	logger.Debug("BM25 index built", map[string]interface{}{
		"num_docs":     len(docs),
		"avg_doc_len":  avgDocLen,
		"total_tokens": int(totalLen),
		"vocab_size":   len(globalDF),
	})

	return &BM25{
		Docs:      docs,
		DocFreq:   globalDF,
		AvgDocLen: avgDocLen,
		Config:    DefaultBM25Config(),
		Titles:    titles,
		URLs:      urls,
	}
}

// normalizeText applies comprehensive text normalization for BM25
func normalizeText(text string) []string {
	// 1. Lowercase the text
	text = strings.ToLower(text)

	// 2. Remove punctuation while preserving meaningful characters
	// Keep alphanumeric, spaces, and some special chars like hyphens, apostrophes
	re := regexp.MustCompile(`[^\w\s\-']`)
	text = re.ReplaceAllString(text, " ")

	// 3. Normalize whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// 4. Tokenize by whitespace
	tokens := strings.Fields(text)

	// 5. Remove stopwords (optional but recommended for better discriminative power)
	tokens = removeStopwords(tokens)

	return tokens
}

// tokenize splits text into lowercase tokens (legacy function, use normalizeText instead)
func tokenize(text string) []string {
	return normalizeText(text)
}

// removeStopwords removes common English stopwords
func removeStopwords(tokens []string) []string {
	stopwords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true,
		"by": true, "for": true, "from": true, "has": true, "he": true, "in": true, "is": true,
		"it": true, "its": true, "of": true, "on": true, "that": true, "the": true, "to": true,
		"was": true, "will": true, "with": true, "would": true, "could": true, "should": true,
		// Academic-specific stopwords that might not be discriminative in paper titles
		"paper": true, "study": true, "research": true, "analysis": true, "method": true,
		"approach": true, "system": true, "model": true, "algorithm": true, "framework": true,
	}

	var filtered []string
	for _, token := range tokens {
		if len(token) > 1 && !stopwords[token] { // Keep tokens longer than 1 char
			filtered = append(filtered, token)
		}
	}

	return filtered
}

// IDF calculates inverse document frequency for a term
func (bm *BM25) IDF(term string) float64 {
	N := float64(len(bm.Docs))
	n := float64(bm.DocFreq[term])
	if n == 0 {
		return 0
	}
	return math.Log((N - n + 0.5) / (n + 0.5))
}

// Score calculates BM25 score for a document against a query
func (bm *BM25) Score(doc []string, query []string) float64 {
	score := 0.0
	termFreq := make(map[string]int)
	for _, token := range doc {
		termFreq[token]++
	}
	docLen := float64(len(doc))

	for _, qTerm := range query {
		f := float64(termFreq[qTerm])
		idf := bm.IDF(qTerm)
		score += idf * ((f * (bm.Config.K1 + 1)) / (f + bm.Config.K1*(1-bm.Config.B+bm.Config.B*docLen/bm.AvgDocLen)))
	}

	return score
}

// Rank returns BM25 scores for all documents against a query
func (bm *BM25) Rank(query string) []float64 {
	qTerms := tokenize(query)
	scores := make([]float64, len(bm.Docs))
	for i, doc := range bm.Docs {
		scores[i] = bm.Score(doc, qTerms)
	}
	return scores
}

// RankTokens returns BM25 scores for all documents against pre-normalized query tokens
func (bm *BM25) RankTokens(queryTokens []string) []float64 {
	scores := make([]float64, len(bm.Docs))
	for i, doc := range bm.Docs {
		scores[i] = bm.Score(doc, queryTokens)
	}
	return scores
}

// findMatchingURL uses BM25 to find the best matching URL for a title
func findMatchingURL(title string, links map[string]string) string {
	bm25 := NewBM25(links)
	return findMatchingURLWithBM25(title, bm25)
}

// findMatchingURLWithBM25 uses a pre-built BM25 instance to find the best matching URL for a title
func findMatchingURLWithBM25(title string, bm25 *BM25) string {
	if len(bm25.Titles) == 0 {
		logger.Debug("No links provided for title matching", map[string]interface{}{
			"title": title,
		})
		return ""
	}

	logger.Debug("Starting BM25 title matching", map[string]interface{}{
		"title":     title,
		"num_docs":  len(bm25.Titles),
	})

	// Normalize query once and reuse
	queryTerms := normalizeText(title)
	titleLength := len(queryTerms)

	// Log query normalization details (only for the query, not corpus)
	logger.Debug("Query normalization", map[string]interface{}{
		"originalTitle":    title,
		"normalizedTokens": queryTerms,
		"tokenCount":       titleLength,
	})

	// Use pre-normalized tokens to avoid double normalization
	scores := bm25.RankTokens(queryTerms)

	var candidates []map[string]interface{}

	// Find best scoring document with detailed logging

	logger.Debug("Query analysis", map[string]interface{}{
		"title": title,
		"queryTerms": queryTerms,
		"titleLength": titleLength,
		"avgDocLen": bm25.AvgDocLen,
		"totalDocs": len(bm25.Docs),
		"vocabSize": len(bm25.DocFreq),
		"bm25Params": map[string]float64{
			"k1": bm25.Config.K1,
			"b":  bm25.Config.B,
		},
	})

	// Process candidates in parallel for better performance with large document sets
	type candidateResult struct {
		index int
		candidate map[string]interface{}
	}

	candidateChan := make(chan candidateResult, len(scores))
	var wg sync.WaitGroup

	for i, score := range scores {
		wg.Add(1)
		go func(idx int, documentScore float64) {
			defer wg.Done()

			docID := strconv.Itoa(idx)
			docTitle := bm25.Titles[docID]
			docTokens := bm25.Docs[idx]

			// Calculate detailed matching info
			termMatches := make(map[string]float64)
			for _, qTerm := range queryTerms {
				idf := bm25.IDF(qTerm)
				tf := 0
				for _, token := range docTokens {
					if token == qTerm {
						tf++
					}
				}
				if tf > 0 {
					termMatches[qTerm] = float64(tf) * idf
				}
			}

			candidate := map[string]interface{}{
				"docID":      docID,
				"title":      docTitle,
				"url":        bm25.URLs[docID],
				"score":      documentScore,
				"termMatches": termMatches,
				"docLength":  len(docTokens),
			}

			candidateChan <- candidateResult{index: idx, candidate: candidate}
		}(i, score)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(candidateChan)
	}()

	// Collect results in original order
	candidates = make([]map[string]interface{}, len(scores))
	for result := range candidateChan {
		candidates[result.index] = result.candidate
	}

	// Filter candidates and find best match
	var finalBestScore float64 = -1
	var finalBestDocID string

	for _, candidate := range candidates {
		docID := candidate["docID"].(string)
		docTitle := candidate["title"].(string)
		score := candidate["score"].(float64)

		// Skip very common terms (appears in >80% of docs) for longer titles only
		shouldSkip := true
		for _, qTerm := range queryTerms {
			// For very short titles (< 3 tokens), be less strict about common terms
			threshold := int(float64(len(bm25.Docs)) * bm25.Config.CommonTermThreshold)
			if titleLength < maxTitleLengthForShort {
				threshold = int(float64(len(bm25.Docs)) * bm25.Config.ShortTitleCommonTermThreshold)
			}
			if bm25.DocFreq[qTerm] <= threshold {
				shouldSkip = false
				break
			}
		}

		if shouldSkip {
			logger.Debug("Skipping document due to common terms", map[string]interface{}{
				"docID": docID,
				"title": docTitle,
				"commonTerms": func() []string {
					var common []string
					for _, qTerm := range queryTerms {
						if bm25.DocFreq[qTerm] > len(bm25.Docs)*8/10 {
							common = append(common, qTerm)
						}
					}
					return common
				}(),
			})
			continue
		}

		// Use different thresholds based on title length
		threshold := bm25.Config.MinScore
		if titleLength < maxTitleLengthForShort {
			threshold = bm25.Config.ShortTitleMinScore
		} else if titleLength < maxTitleLengthForMedium {
			threshold = bm25.Config.MediumTitleMinScore
		}

		logger.Debug("Evaluating candidate", map[string]interface{}{
			"docID":     docID,
			"title":     docTitle,
			"score":     score,
			"threshold": threshold,
		})

		if score > finalBestScore && score > threshold {
			finalBestScore = score
			finalBestDocID = docID
		}
	}

	// Special fallback for very short titles (acronyms) - if no good BM25 match,
	// try simple substring matching as a last resort
	if finalBestDocID == "" && titleLength <= 2 && titleLength > 0 {
		logger.Debug("Trying fallback substring matching for short title", map[string]interface{}{
			"title": title,
			"titleLength": titleLength,
		})

		titleLower := strings.ToLower(title)
		for i := range scores {
			docID := strconv.Itoa(i)
			docTitle := strings.ToLower(bm25.Titles[docID])

			// Try substring matching for very short titles
			if strings.Contains(docTitle, titleLower) || strings.Contains(titleLower, docTitle) {
				// Calculate a simple similarity score
				similarity := 0.0
				if strings.Contains(docTitle, titleLower) {
					similarity = float64(len(titleLower)) / float64(len(docTitle))
				} else {
					similarity = float64(len(docTitle)) / float64(len(titleLower))
				}

				if similarity > bm25.Config.FallbackSimilarityThreshold {
					finalBestScore = similarity
					finalBestDocID = docID
					logger.Debug("Fallback substring match for short title", map[string]interface{}{
						"searchTitle":  title,
						"matchedTitle": bm25.Titles[docID],
						"similarity":   similarity,
					})
					break
				}
			}
		}
	}

	// Always log all candidates for debugging (moved before match check)
	logger.Debug("All BM25 candidates", map[string]interface{}{
		"searchTitle": title,
		"candidates":  candidates,
		"bestScore":   finalBestScore,
		"bestDocID":   finalBestDocID,
	})

	if finalBestDocID != "" {
		matchedTitle := bm25.Titles[finalBestDocID]
		bestURL := bm25.URLs[finalBestDocID]

		// Calculate what threshold was used
		usedThreshold := bm25.Config.MinScore
		if titleLength < maxTitleLengthForShort {
			usedThreshold = bm25.Config.ShortTitleMinScore
		} else if titleLength < maxTitleLengthForMedium {
			usedThreshold = bm25.Config.MediumTitleMinScore
		}

		logger.Info("BM25 title match found", map[string]interface{}{
			"searchTitle":   title,
			"matchedTitle":  matchedTitle,
			"bestScore":     finalBestScore,
			"usedThreshold": usedThreshold,
			"titleLength":   titleLength,
			"numCandidates": len(candidates),
		})

		return bestURL
	}

	// No match found - calculate what threshold was used for logging
	usedThreshold := bm25.Config.MinScore
	if titleLength < maxTitleLengthForShort {
		usedThreshold = bm25.Config.ShortTitleMinScore
	} else if titleLength < maxTitleLengthForMedium {
		usedThreshold = bm25.Config.MediumTitleMinScore
	}

	logger.Warn("No BM25 title match found", map[string]interface{}{
		"searchTitle":   title,
		"bestScore":     finalBestScore,
		"usedThreshold": usedThreshold,
		"titleLength":   titleLength,
		"numCandidates": len(candidates),
		"topCandidates": getTopCandidates(candidates, 5), // Show top 5 for analysis
	})

	return ""
}

// getTopCandidates returns the top N candidates sorted by score for analysis
func getTopCandidates(candidates []map[string]interface{}, n int) []map[string]interface{} {
	if len(candidates) <= n {
		return candidates
	}

	// Sort candidates by score (descending)
	sorted := make([]map[string]interface{}, len(candidates))
	copy(sorted, candidates)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			scoreI := sorted[i]["score"].(float64)
			scoreJ := sorted[j]["score"].(float64)
			if scoreJ > scoreI {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[:n]
}

// createValidationError creates a standardized ValidationError
func createValidationError(field, message, details string, severity ValidationSeverity) ValidationError {
	return ValidationError{
		Field:    field,
		Message:  message,
		Details:  details,
		Severity: severity,
	}
}

// normalizeURLForComparison cleans and normalizes a URL for consistent comparison
func normalizeURLForComparison(url string) string {
	// Strip query parameters and fragments
	if idx := strings.IndexByte(url, '?'); idx != -1 {
		url = url[:idx]
	}
	if idx := strings.IndexByte(url, '#'); idx != -1 {
		url = url[:idx]
	}
	return strings.ToLower(strings.TrimSuffix(url, "/"))
}

// normalizeURLForPathExtraction cleans URL but preserves path structure for extraction
func normalizeURLForPathExtraction(url string) string {
	// Strip query parameters and fragments
	if idx := strings.IndexByte(url, '?'); idx != -1 {
		url = url[:idx]
	}
	if idx := strings.IndexByte(url, '#'); idx != -1 {
		url = url[:idx]
	}
	return strings.TrimSuffix(url, "/")
}

// validateSummaryContent performs all validations on the summary
func validateSummaryContent(markdown string, feedURLs map[string]string) error {
	// Run validations in parallel using goroutines
	errChan := make(chan error, 4)

	go func() {
		if err := validateMarkdownStructure(markdown); err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	go func() {
		if err := validateMarkdownLinks(markdown, feedURLs); err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	go func() {
		if err := validateSummaryLength(markdown); err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	go func() {
		if err := validateNoDuplicateTitles(markdown, feedURLs); err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	// Collect errors
	var errors []error
	for i := 0; i < 4; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	return nil
}

// validateNoDuplicateTitles checks that no paper title is mentioned more than once
func validateNoDuplicateTitles(markdown string, feedURLs map[string]string) error {
	// Extract all paper titles from markdown links [Title](URL)
	titleRegex := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	matches := titleRegex.FindAllStringSubmatch(markdown, -1)

	if len(matches) == 0 {
		return createValidationError("titles", "no paper titles found in summary", markdown, SeverityError)
	}

	// Track seen titles
	seenTitles := make(map[string]int)
	var duplicates []string

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		title := strings.TrimSpace(match[1])

		// Skip empty titles or titles that are just punctuation
		if title == "" || regexp.MustCompile(`^[[:punct:]]+$`).MatchString(title) {
			continue
		}

		seenTitles[title]++
		if seenTitles[title] > 1 {
			duplicates = append(duplicates, title)
		}
	}

	if len(duplicates) > 0 {
		return createValidationError(
			"duplicates",
			fmt.Sprintf("duplicate paper titles found: %v", duplicates),
			fmt.Sprintf("Titles that appear multiple times: %v", duplicates),
			SeverityError,
		)
	}

	return nil
}

// validateMarkdownStructure checks if the markdown has required sections
func validateMarkdownStructure(markdown string) error {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(markdown))

	foundSections := make(map[string]bool)
	var headlineText string
	var collectingHeadlineContent bool
	var currentSectionTextBuilder strings.Builder

	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if h, ok := node.(*ast.Heading); ok && h.Level == 2 {
			if entering {
				if collectingHeadlineContent {
					headlineText = strings.TrimSpace(currentSectionTextBuilder.String())
					collectingHeadlineContent = false
				}

				var currentHeadingTitle string
				for _, child := range h.Children {
					if t, ok := child.(*ast.Text); ok {
						currentHeadingTitle += string(t.Literal)
					}
				}
				currentHeadingTitle = strings.TrimSpace(currentHeadingTitle)
				foundSections[currentHeadingTitle] = true

				if currentHeadingTitle == "Morning Headline" {
					collectingHeadlineContent = true
					currentSectionTextBuilder.Reset()
				}
			}
			return ast.GoToNext
		}

		if collectingHeadlineContent && entering {
			if para, ok := node.(*ast.Paragraph); ok {
				var paragraphContent strings.Builder
				for _, child := range para.Children {
					if t, ok := child.(*ast.Text); ok {
						paragraphContent.WriteString(string(t.Literal))
					} else if l, ok := child.(*ast.Link); ok {
						for _, linkChild := range l.Children {
							if lt, ok := linkChild.(*ast.Text); ok {
								paragraphContent.WriteString(string(lt.Literal))
							}
						}
					}
				}
				if currentSectionTextBuilder.Len() > 0 {
					currentSectionTextBuilder.WriteString(" ")
				}
				currentSectionTextBuilder.WriteString(paragraphContent.String())
			}
		}
		return ast.GoToNext
	})

	if collectingHeadlineContent {
		headlineText = strings.TrimSpace(currentSectionTextBuilder.String())
	}

	requiredSections := []string{"Morning Headline", "What's New"}
	for _, section := range requiredSections {
		if !foundSections[section] {
			return createValidationError(
				"structure",
				fmt.Sprintf("missing required section: %s", section),
				markdown,
				SeverityError,
			)
		}
	}

	if headlineText == "" {
		return createValidationError("headline", "headline content is empty", markdown, SeverityError)
	}

	headlineText = regexp.MustCompile(`\s+`).ReplaceAllString(headlineText, " ")
	if len(headlineText) > maxHeadlineLength {
		return createValidationError(
			"headline",
			fmt.Sprintf("headline too long: %d characters (limit %d)", len(headlineText), maxHeadlineLength),
			headlineText,
			SeverityError,
		)
	}

	return nil
}

// validateMarkdownLinks checks if all markdown links are properly formatted and have URLs
func validateMarkdownLinks(markdown string, feedURLs map[string]string) error {
	normalizedFeedURLs := make(map[string]string)
	for url := range feedURLs {
		normalizedFeedURLs[normalizeURLForComparison(url)] = url
	}

	// Check for links that are missing URLs completely [text] instead of [text](url)
	// Find all [text] patterns and filter out those followed by (
	allBracketRegex := regexp.MustCompile(`\[([^\]]+)\]`)
	allBracketMatches := allBracketRegex.FindAllStringSubmatch(markdown, -1)

	var missingURLMatches [][]string
	for _, match := range allBracketMatches {
		if len(match) >= 2 {
			// Check if this bracket is followed by ( - if not, it's missing URL
			bracketEnd := strings.Index(markdown, match[0])
			if bracketEnd >= 0 {
				// Look ahead to see if next character is (
				nextChar := ""
				if bracketEnd+len(match[0]) < len(markdown) {
					nextChar = string(markdown[bracketEnd+len(match[0])])
				}
				if nextChar != "(" {
					missingURLMatches = append(missingURLMatches, match)
				}
			}
		}
	}

	// Check for properly formatted links [text](url)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	matches := linkRegex.FindAllStringSubmatch(markdown, -1)

	if len(matches) == 0 && len(missingURLMatches) == 0 {
		return createValidationError("links", "no markdown links found in summary", markdown, SeverityError)
	}

	var errors []string

	// Check for links missing URLs completely
	for _, match := range missingURLMatches {
		if len(match) >= 2 {
			title := strings.TrimSpace(match[1])
			if title != "" {
				errors = append(errors, fmt.Sprintf("link missing URL: [%s]", title))
			}
		}
	}

	// Check for properly formatted links
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		title := match[1]
		url := match[2]

		// Check for empty URLs
		if strings.TrimSpace(url) == "" {
			errors = append(errors, fmt.Sprintf("empty URL for link: [%s]", title))
			continue
		}

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			errors = append(errors, fmt.Sprintf("invalid protocol: %s", url))
			continue
		}

		normalizedURL := normalizeURLForComparison(url)
		if _, exists := normalizedFeedURLs[normalizedURL]; !exists {
			continue // Allow links not in feed for now
		}
	}

	if len(errors) > 0 {
		return createValidationError(
			"links",
			fmt.Sprintf("found %d link formatting errors", len(errors)),
			fmt.Sprintf("link errors: %v", errors),
			SeverityError,
		)
	}

	return nil
}

// validateSummaryLength checks if the summary is within reasonable bounds
func validateSummaryLength(markdown string) error {
	plainText := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(markdown, "$1")
	words := strings.Fields(plainText)

	if len(words) > maxSummaryWords {
		return createValidationError(
			"length",
			fmt.Sprintf("summary too long: %d words", len(words)),
			fmt.Sprintf("max allowed: %d, current: %d", maxSummaryWords, len(words)),
			SeverityError,
		)
	}

	if len(words) < minSummaryWords {
		return createValidationError(
			"length",
			fmt.Sprintf("summary too short: %d words", len(words)),
			fmt.Sprintf("min expected: %d, current: %d", minSummaryWords, len(words)),
			SeverityError,
		)
	}

	if len(words) > warningSummaryWords {
		slog.Warn("Summary approaching length limit", "word_count", len(words), "max_allowed", maxSummaryWords)
	}

	return nil
}
