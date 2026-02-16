package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/lib/article"
	"net/http"
	"time"
)

// SummarizerConfig contains configuration for the article summarizer
type SummarizerConfig struct {
	APIKey      string        // Claude API key from environment
	Model       string        // Claude model (default: "claude-3-5-sonnet-20241022")
	MaxTokens   int           // Maximum tokens for summary (~150 for 2-3 sentences)
	Temperature float64       // Temperature for generation (0.7 = balanced)
	TimeoutSec  int           // API timeout in seconds
}

// ArticleSummarizer generates summaries for articles using Claude API
type ArticleSummarizer struct {
	config *SummarizerConfig
	client *http.Client
}

// claudeMessage represents a message in the Claude API request
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeRequest represents the Claude API request body
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

// claudeResponse represents the Claude API response body
type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewArticleSummarizer creates a new article summarizer
func NewArticleSummarizer(config *SummarizerConfig) (*ArticleSummarizer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid summarizer config: %w", err)
	}

	return &ArticleSummarizer{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.TimeoutSec) * time.Second,
		},
	}, nil
}

// Validate checks if the configuration is valid
func (sc *SummarizerConfig) Validate() error {
	if sc.APIKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	if sc.Model == "" {
		sc.Model = "claude-3-5-sonnet-20241022"
	}
	if sc.MaxTokens <= 0 {
		sc.MaxTokens = 150
	}
	if sc.Temperature == 0 {
		sc.Temperature = 0.7
	}
	if sc.TimeoutSec <= 0 {
		sc.TimeoutSec = 30
	}
	if sc.Temperature < 0 || sc.Temperature > 1 {
		return fmt.Errorf("temperature must be between 0 and 1")
	}
	return nil
}

// SummarizeArticle generates a summary for a single article
func (as *ArticleSummarizer) SummarizeArticle(ctx context.Context, art *article.ArticleData) (string, error) {
	if art == nil {
		return "", fmt.Errorf("article cannot be nil")
	}
	if art.Title == "" || art.URL == "" {
		return "", fmt.Errorf("article must have title and URL")
	}

	// Build the prompt with article context
	prompt := fmt.Sprintf(
		"Summarize this iGaming news article in 2-3 sentences for a news digest. Focus on key insights and impact.\n\n"+
			"Title: %s\n"+
			"Source: %s\n"+
			"Summary: %s\n\n"+
			"Provide a concise, professional summary:",
		art.Title, art.SourceName, art.OriginalSum,
	)

	// Create Claude API request
	req := claudeRequest{
		Model:       as.config.Model,
		MaxTokens:   as.config.MaxTokens,
		Temperature: as.config.Temperature,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to Claude API
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", as.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := as.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to parse Claude response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("Claude API returned empty content")
	}

	summary := claudeResp.Content[0].Text

	// Store metadata about the summary
	if art.Metadata == nil {
		art.Metadata = make(map[string]interface{})
	}
	art.Metadata["summarizer_version"] = "1.0"
	art.Metadata["model_used"] = as.config.Model
	art.Metadata["tokens_used"] = claudeResp.Usage.OutputTokens
	art.Metadata["summarized_at"] = time.Now().Format(time.RFC3339)

	art.UpdatedAt = time.Now()

	return summary, nil
}

// SummarizeBatch summarizes multiple articles sequentially with rate limiting
func (as *ArticleSummarizer) SummarizeBatch(ctx context.Context, articles []article.ArticleData) error {
	if len(articles) == 0 {
		return nil
	}

	// Rate limiting: 1 article per second to respect API limits
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := range articles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Continue to next iteration
		}

		summary, err := as.SummarizeArticle(ctx, &articles[i])
		if err != nil {
			// Log failure but continue (graceful degradation)
			fmt.Printf("Failed to summarize article %s: %v\n", articles[i].ID, err)
			// Set empty summary as fallback
			articles[i].Summary = ""
			continue
		}

		articles[i].Summary = summary
	}

	return nil
}
