package feed

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"main/lib/article"
	"main/lib/logger"
	"net/http"
	"strings"
	"time"
)

// FetcherConfig holds configuration for article fetching
type FetcherConfig struct {
	Timeout      time.Duration
	RetryAttempts int
	RetryDelay   time.Duration
	UserAgent    string
}

// DefaultFetcherConfig returns default configuration
func DefaultFetcherConfig() *FetcherConfig {
	return &FetcherConfig{
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
		UserAgent:     "iGaming-TLDR/1.0 (+https://gaming-tldr.example.com)",
	}
}

// ArticleFetcher fetches articles from news sources
type ArticleFetcher struct {
	config *FetcherConfig
	client *http.Client
}

// NewArticleFetcher creates a new article fetcher
func NewArticleFetcher(config *FetcherConfig) *ArticleFetcher {
	if config == nil {
		config = DefaultFetcherConfig()
	}

	return &ArticleFetcher{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// FetchFromSource fetches and parses articles from a single source
func (af *ArticleFetcher) FetchFromSource(ctx context.Context, source *NewsSource) ([]article.ArticleData, error) {
	if !source.Active {
		return nil, fmt.Errorf("source %s is not active", source.Name)
	}

	switch source.ScrapingType {
	case "rss":
		return af.fetchFromRSS(ctx, source)
	case "scrape":
		// TODO: Implement web scraping in Phase 2
		return nil, fmt.Errorf("web scraping not yet implemented")
	case "api":
		// TODO: Implement API scraping in Phase 2
		return nil, fmt.Errorf("API scraping not yet implemented")
	default:
		return nil, fmt.Errorf("unknown scraping type: %s", source.ScrapingType)
	}
}

// fetchFromRSS fetches articles from an RSS feed
func (af *ArticleFetcher) fetchFromRSS(ctx context.Context, source *NewsSource) ([]article.ArticleData, error) {
	logger.Info("Fetching from RSS feed", map[string]interface{}{
		"source": source.Name,
		"url":    source.FeedURL,
	})

	feedData, err := af.fetchRSSFeed(ctx, source.FeedURL)
	if err != nil {
		return nil, err
	}

	articles := af.parseRSSFeed(feedData, source)
	logger.Info("Successfully parsed RSS feed", map[string]interface{}{
		"source":        source.Name,
		"articleCount":  len(articles),
	})

	return articles, nil
}

// fetchRSSFeed fetches RSS feed with retry logic
func (af *ArticleFetcher) fetchRSSFeed(ctx context.Context, feedURL string) (*RssFeed, error) {
	var lastErr error

	for attempt := 0; attempt < af.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(af.config.RetryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		feed, err := af.fetchRSSFeedAttempt(ctx, feedURL)
		if err == nil {
			return feed, nil
		}

		lastErr = err
		logger.Warn("RSS fetch attempt failed, retrying", map[string]interface{}{
			"url":     feedURL,
			"attempt": attempt + 1,
			"error":   err.Error(),
		})
	}

	return nil, fmt.Errorf("failed to fetch RSS feed after %d attempts: %w", af.config.RetryAttempts, lastErr)
}

// fetchRSSFeedAttempt attempts to fetch RSS feed once
func (af *ArticleFetcher) fetchRSSFeedAttempt(ctx context.Context, feedURL string) (*RssFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to avoid blocking
	req.Header.Set("User-Agent", af.config.UserAgent)

	resp, err := af.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse RSS feed
	var rssData struct {
		Channel struct {
			Title         string `xml:"title"`
			Description   string `xml:"description"`
			Link          string `xml:"link"`
			LastBuildDate string `xml:"lastBuildDate"`
			Items         []struct {
				Title       string `xml:"title"`
				Link        string `xml:"link"`
				Description string `xml:"description"`
				PubDate     string `xml:"pubDate"`
				GUID        string `xml:"guid"`
				// Additional fields for news content
				Content     string `xml:"content"`
				Image       string `xml:"image"`
				Author      string `xml:"author"`
				Categories  []string `xml:"category"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	if err := xml.Unmarshal(body, &rssData); err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	// Convert to our feed format
	feed := &RssFeed{
		Title:         rssData.Channel.Title,
		Description:   rssData.Channel.Description,
		Link:          rssData.Channel.Link,
		LastBuildDate: rssData.Channel.LastBuildDate,
		Items:         make([]FeedItem, 0, len(rssData.Channel.Items)),
	}

	for _, item := range rssData.Channel.Items {
		feed.Items = append(feed.Items, FeedItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			PubDate:     item.PubDate,
			GUID:        GUIDString(item.GUID),
		})
	}

	return feed, nil
}

// parseRSSFeed converts RSS feed items to article data
func (af *ArticleFetcher) parseRSSFeed(feed *RssFeed, source *NewsSource) []article.ArticleData {
	articles := make([]article.ArticleData, 0, len(feed.Items))

	for _, item := range feed.Items {
		parsed := af.parseRSSItem(item, source)
		if parsed != nil {
			articles = append(articles, *parsed)
		}
	}

	return articles
}

// parseRSSItem converts a single RSS item to article data
func (af *ArticleFetcher) parseRSSItem(item FeedItem, source *NewsSource) *article.ArticleData {
	if item.Title == "" || item.Link == "" {
		logger.Warn("Skipping incomplete RSS item", map[string]interface{}{
			"title": item.Title,
			"link":  item.Link,
		})
		return nil
	}

	// Generate ID from URL hash
	id := generateArticleID(item.Link)

	// Parse publication date
	pubDate, err := parsePublishDate(item.PubDate)
	if err != nil {
		logger.Debug("Failed to parse publish date", map[string]interface{}{
			"date":  item.PubDate,
			"error": err.Error(),
		})
		pubDate = time.Now() // Fallback to now
	}

	article := &article.ArticleData{
		ID:            id,
		Title:         item.Title,
		OriginalSum:   stripHTML(item.Description),
		URL:           item.Link,
		SourceName:    source.Name,
		SourceID:      source.ID,
		PublishedDate: pubDate.Format(time.RFC3339),
		Categories:    []string{source.Category},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return &article
}

// FetchFromSources fetches articles from multiple sources
func (af *ArticleFetcher) FetchFromSources(ctx context.Context, sources []*NewsSource) ([]article.ArticleData, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources provided")
	}

	allArticles := make([]article.ArticleData, 0)
	errCount := 0

	for _, source := range sources {
		articles, err := af.FetchFromSource(ctx, source)
		if err != nil {
			logger.Error("Failed to fetch from source", err, map[string]interface{}{
				"source": source.Name,
			})
			errCount++
			continue
		}

		allArticles = append(allArticles, articles...)
	}

	if errCount == len(sources) {
		return nil, fmt.Errorf("failed to fetch from all %d sources", len(sources))
	}

	logger.Info("Fetch complete from multiple sources", map[string]interface{}{
		"totalSources":   len(sources),
		"failedSources":  errCount,
		"totalArticles":  len(allArticles),
	})

	return allArticles, nil
}

// Helper function to generate article ID from URL
func generateArticleID(url string) string {
	// Simple hash based on URL
	hash := 0
	for _, char := range url {
		hash = ((hash << 5) - hash) + int(char)
	}
	return fmt.Sprintf("%x", uint32(hash))
}

// Helper function to parse common date formats
func parsePublishDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}

	// Try common RSS date formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		time.RFC822,
		time.RFC822Z,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Now(), fmt.Errorf("unable to parse date: %s", dateStr)
}

// Helper function to strip HTML tags from text
func stripHTML(html string) string {
	// Simple HTML tag removal
	result := html
	start := 0
	for {
		start = strings.Index(result[start:], "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}

	// Decode common HTML entities
	result = strings.ReplaceAll(result, "&amp;", "&")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&quot;", "\"")
	result = strings.ReplaceAll(result, "&#39;", "'")

	// Clean up whitespace
	result = strings.TrimSpace(result)

	return result
}
