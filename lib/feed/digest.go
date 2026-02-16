package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/lib/article"
	"net/http"
	"sort"
	"strings"
	"time"
)

// DigestBuilder creates daily digests with top-ranked articles
type DigestBuilder struct {
	cache      *ArticleCache
	ranker     *RankingEngine
	summarizer *ArticleSummarizer
}

// DigestOptions configures digest creation
type DigestOptions struct {
	TopN           int     // Default: 5
	MinScore       float64 // Default: 0.0
	IncludeReasons bool    // Default: true
}

// NewDigestBuilder creates a new digest builder
func NewDigestBuilder(cache *ArticleCache, ranker *RankingEngine, summarizer *ArticleSummarizer) *DigestBuilder {
	return &DigestBuilder{
		cache:      cache,
		ranker:     ranker,
		summarizer: summarizer,
	}
}

// BuildDailyDigest creates a digest for a specific date
func (db *DigestBuilder) BuildDailyDigest(date string) (*article.DailyDigest, error) {
	// Validate date format (YYYY-MM-DD)
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: expected YYYY-MM-DD, got %s", date)
	}

	articles := db.cache.GetAll()
	opts := &DigestOptions{
		TopN:           5,
		MinScore:       0.0,
		IncludeReasons: true,
	}

	return db.BuildDigestFromArticles(articles, opts, date)
}

// BuildTodayDigest creates a digest for today
func (db *DigestBuilder) BuildTodayDigest() (*article.DailyDigest, error) {
	today := time.Now().Format("2006-01-02")
	return db.BuildDailyDigest(today)
}

// BuildDigestFromArticles creates a digest from a specific set of articles
func (db *DigestBuilder) BuildDigestFromArticles(articles []article.ArticleData, opts *DigestOptions, dateStr string) (*article.DailyDigest, error) {
	if opts == nil {
		opts = &DigestOptions{
			TopN:           5,
			MinScore:       0.0,
			IncludeReasons: true,
		}
	}

	// Ensure TopN is reasonable
	if opts.TopN <= 0 {
		opts.TopN = 5
	}

	// Rank articles
	rankedArticles, err := db.ranker.RankArticles(articles)
	if err != nil {
		return nil, fmt.Errorf("failed to rank articles: %w", err)
	}

	// Filter by minimum score and take top N
	var selectedArticles []article.RankedArticle
	for _, ranked := range rankedArticles {
		if ranked.Score >= opts.MinScore && len(selectedArticles) < opts.TopN {
			selectedArticles = append(selectedArticles, ranked)
		}
	}

	// Create digest
	digest := &article.DailyDigest{
		Date:     dateStr,
		Articles: selectedArticles,
		Created:  time.Now(),
	}

	// Generate digest summary and headline from Claude API if summarizer available
	if db.summarizer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		headline, err := db.generateDigestHeadline(ctx, selectedArticles)
		if err != nil {
			// Log error but don't fail (graceful degradation)
			fmt.Printf("Failed to generate digest headline: %v\n", err)
			digest.Headline = db.fallbackDigestHeadline(selectedArticles)
		} else {
			digest.Headline = headline
		}

		summary, err := db.generateDigestSummary(ctx, selectedArticles)
		if err != nil {
			// Log error but don't fail (graceful degradation)
			fmt.Printf("Failed to generate digest summary: %v\n", err)
			digest.Summary = db.fallbackDigestSummary(selectedArticles)
		} else {
			digest.Summary = summary
		}
	} else {
		digest.Headline = db.fallbackDigestHeadline(selectedArticles)
		digest.Summary = db.fallbackDigestSummary(selectedArticles)
	}

	return digest, nil
}

// generateDigestSummary calls Claude API to generate an executive summary
func (db *DigestBuilder) generateDigestSummary(ctx context.Context, articles []article.RankedArticle) (string, error) {
	if len(articles) == 0 {
		return "", fmt.Errorf("no articles to summarize")
	}

	// Build context from article titles and summaries
	var articleContext strings.Builder
	for i, ranked := range articles {
		articleContext.WriteString(fmt.Sprintf("%d. %s\n", i+1, ranked.Article.Title))
		if ranked.Article.Summary != "" {
			articleContext.WriteString(fmt.Sprintf("   %s\n", ranked.Article.Summary))
		}
	}

	prompt := fmt.Sprintf(
		"Write a 1-paragraph executive summary (~3-4 sentences) of these top iGaming news articles. Focus on key trends and industry impact.\n\n%s",
		articleContext.String(),
	)

	// Create Claude API request
	req := claudeRequest{
		Model:       db.summarizer.config.Model,
		MaxTokens:   200,
		Temperature: 0.7,
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

	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", db.summarizer.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

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

	return claudeResp.Content[0].Text, nil
}

// generateDigestHeadline calls Claude API to generate a one-sentence headline
func (db *DigestBuilder) generateDigestHeadline(ctx context.Context, articles []article.RankedArticle) (string, error) {
	if len(articles) == 0 {
		return "", fmt.Errorf("no articles to create headline from")
	}

	// Build context from article titles
	var articleContext strings.Builder
	for i, ranked := range articles {
		articleContext.WriteString(fmt.Sprintf("%d. %s\n", i+1, ranked.Article.Title))
	}

	prompt := fmt.Sprintf(
		"Write a single, compelling headline (one sentence, max 15 words) that captures the main theme of today's iGaming news. Be specific and newsworthy.\n\nTop stories:\n%s",
		articleContext.String(),
	)

	// Create Claude API request
	req := claudeRequest{
		Model:       db.summarizer.config.Model,
		MaxTokens:   50,
		Temperature: 0.7,
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

	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", db.summarizer.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

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

	return claudeResp.Content[0].Text, nil
}

// fallbackDigestHeadline creates a simple headline from top article
func (db *DigestBuilder) fallbackDigestHeadline(articles []article.RankedArticle) string {
	if len(articles) == 0 {
		return "No major stories today"
	}

	// Use the top-ranked article title as the headline
	return fmt.Sprintf("Top story: %s", articles[0].Article.Title)
}

// fallbackDigestSummary creates a summary from article titles and scores
func (db *DigestBuilder) fallbackDigestSummary(articles []article.RankedArticle) string {
	if len(articles) == 0 {
		return "No articles available for today's digest."
	}

	// Sort by rank
	sorted := make([]article.RankedArticle, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rank < sorted[j].Rank
	})

	// Build summary from titles
	var titles []string
	for _, ranked := range sorted {
		titles = append(titles, ranked.Article.Title)
	}

	return fmt.Sprintf(
		"Today's top %d iGaming news stories: %s",
		len(articles),
		strings.Join(titles, "; "),
	)
}
