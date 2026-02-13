package feed

import (
	"context"
	"testing"
	"time"
)

// TestNewArticleFetcher tests creating a new fetcher
func TestNewArticleFetcher(t *testing.T) {
	fetcher := NewArticleFetcher(nil)

	if fetcher == nil {
		t.Fatal("NewArticleFetcher should not return nil")
	}

	if fetcher.config == nil {
		t.Error("Fetcher config should be initialized")
	}

	if fetcher.client == nil {
		t.Error("Fetcher client should be initialized")
	}
}

// TestDefaultFetcherConfig tests default configuration
func TestDefaultFetcherConfig(t *testing.T) {
	config := DefaultFetcherConfig()

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}

	if config.RetryAttempts != 3 {
		t.Errorf("Expected 3 retry attempts, got %d", config.RetryAttempts)
	}

	if config.UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
}

// TestCustomFetcherConfig tests custom configuration
func TestCustomFetcherConfig(t *testing.T) {
	config := &FetcherConfig{
		Timeout:       15 * time.Second,
		RetryAttempts: 5,
		RetryDelay:    2 * time.Second,
		UserAgent:     "CustomAgent/1.0",
	}

	fetcher := NewArticleFetcher(config)

	if fetcher.config.Timeout != 15*time.Second {
		t.Errorf("Expected custom timeout, got %v", fetcher.config.Timeout)
	}

	if fetcher.config.RetryAttempts != 5 {
		t.Errorf("Expected 5 retry attempts, got %d", fetcher.config.RetryAttempts)
	}
}

// TestFetchFromInactiveSource tests error handling for inactive source
func TestFetchFromInactiveSource(t *testing.T) {
	fetcher := NewArticleFetcher(nil)
	inactiveSource := &NewsSource{
		ID:      "test",
		Name:    "Test Source",
		Active:  false,
		FeedURL: "https://example.com/feed",
	}

	ctx := context.Background()
	_, err := fetcher.FetchFromSource(ctx, inactiveSource)

	if err == nil {
		t.Error("FetchFromSource should return error for inactive source")
	}
}

// TestFetchFromUnsupportedScrapingType tests error for unknown scraping type
func TestFetchFromUnsupportedScrapingType(t *testing.T) {
	fetcher := NewArticleFetcher(nil)
	source := &NewsSource{
		ID:           "test",
		Name:         "Test Source",
		Active:       true,
		FeedURL:      "https://example.com/feed",
		ScrapingType: "unsupported",
	}

	ctx := context.Background()
	_, err := fetcher.FetchFromSource(ctx, source)

	if err == nil {
		t.Error("FetchFromSource should return error for unsupported scraping type")
	}
}

// TestGenerateArticleID tests article ID generation
func TestGenerateArticleID(t *testing.T) {
	url1 := "https://example.com/article-1"
	url2 := "https://example.com/article-1"
	url3 := "https://example.com/article-2"

	id1 := generateArticleID(url1)
	id2 := generateArticleID(url2)
	id3 := generateArticleID(url3)

	if id1 == "" {
		t.Error("Generated ID should not be empty")
	}

	if id1 != id2 {
		t.Error("Same URL should generate same ID")
	}

	if id1 == id3 {
		t.Error("Different URLs should generate different IDs")
	}
}

// TestParsePublishDate tests date parsing
func TestParsePublishDate(t *testing.T) {
	testCases := []struct {
		name      string
		dateStr   string
		shouldErr bool
	}{
		{
			name:      "RFC1123Z format",
			dateStr:   "Wed, 02 Jun 2026 15:30:00 +0000",
			shouldErr: false,
		},
		{
			name:      "RFC3339 format",
			dateStr:   "2026-02-13T10:30:00Z",
			shouldErr: false,
		},
		{
			name:      "Empty date string",
			dateStr:   "",
			shouldErr: false, // Returns current time
		},
		{
			name:      "Invalid date format",
			dateStr:   "not-a-date",
			shouldErr: false, // Returns current time as fallback
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parsePublishDate(tc.dateStr)

			if tc.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.shouldErr && err == nil && result.IsZero() {
				t.Error("Should return non-zero time")
			}
		})
	}
}

// TestStripHTML tests HTML tag removal
func TestStripHTML(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "<p>Hello World</p>",
			expected: "Hello World",
		},
		{
			input:    "<div>Test <span>content</span></div>",
			expected: "Test content",
		},
		{
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			input:    "<p>Multiple   spaces</p>",
			expected: "Multiple   spaces",
		},
		{
			input:    "Text with &amp; entities &lt;tag&gt;",
			expected: "Text with & entities <tag>",
		},
	}

	for _, tc := range testCases {
		result := stripHTML(tc.input)
		if result != tc.expected {
			t.Errorf("Expected '%s', got '%s'", tc.expected, result)
		}
	}
}

// TestStripHTMLEntityDecoding tests HTML entity decoding
func TestStripHTMLEntityDecoding(t *testing.T) {
	input := "Title with &amp; ampersand &quot;quotes&quot; and &#39;apostrophe&#39;"
	result := stripHTML(input)

	if !contains(result, "&") {
		t.Error("Should decode &amp; to &")
	}

	if !contains(result, "\"") {
		t.Error("Should decode &quot; to \"")
	}

	if !contains(result, "'") {
		t.Error("Should decode &#39; to '")
	}
}

// TestFetchFromSourcesWithEmptyList tests error for empty source list
func TestFetchFromSourcesWithEmptyList(t *testing.T) {
	fetcher := NewArticleFetcher(nil)
	ctx := context.Background()

	_, err := fetcher.FetchFromSources(ctx, []*NewsSource{})

	if err == nil {
		t.Error("FetchFromSources should return error for empty source list")
	}
}

// TestCacheEntry tests cache entry structure
func TestCacheEntry(t *testing.T) {
	now := time.Now()
	entry := &CacheEntry{
		Timestamp: now,
		ExpiresAt: now.Add(5 * time.Minute),
	}

	if entry.Timestamp != now {
		t.Error("Timestamp should be set correctly")
	}

	if entry.ExpiresAt.Before(now) {
		t.Error("ExpiresAt should be in future")
	}
}

// TestParseRSSItem tests RSS item parsing
func TestParseRSSItem(t *testing.T) {
	fetcher := NewArticleFetcher(nil)
	source := &NewsSource{
		ID:       "test-source",
		Name:     "Test Source",
		Category: "Business",
	}

	item := FeedItem{
		Title:       "Test Article",
		Link:        "https://example.com/article",
		Description: "<p>Test description</p>",
		PubDate:     "Wed, 02 Jun 2026 15:30:00 +0000",
		GUID:        "guid-123",
	}

	article := fetcher.parseRSSItem(item, source)

	if article == nil {
		t.Fatal("parseRSSItem should not return nil")
	}

	if article.Title != "Test Article" {
		t.Errorf("Expected title 'Test Article', got '%s'", article.Title)
	}

	if article.URL != "https://example.com/article" {
		t.Errorf("Expected URL, got '%s'", article.URL)
	}

	if article.SourceName != "Test Source" {
		t.Errorf("Expected source 'Test Source', got '%s'", article.SourceName)
	}

	if len(article.Categories) == 0 {
		t.Error("Article should have at least one category")
	}
}

// TestParseRSSItemWithMissingFields tests handling of incomplete items
func TestParseRSSItemWithMissingFields(t *testing.T) {
	fetcher := NewArticleFetcher(nil)
	source := &NewsSource{
		ID:   "test",
		Name: "Test",
	}

	// Item missing title
	item1 := FeedItem{
		Link: "https://example.com/article",
	}

	result1 := fetcher.parseRSSItem(item1, source)
	if result1 != nil {
		t.Error("Should return nil for item missing title")
	}

	// Item missing link
	item2 := FeedItem{
		Title: "Test Article",
	}

	result2 := fetcher.parseRSSItem(item2, source)
	if result2 != nil {
		t.Error("Should return nil for item missing link")
	}
}

// TestParseRSSFeed tests parsing multiple RSS items
func TestParseRSSFeed(t *testing.T) {
	fetcher := NewArticleFetcher(nil)
	source := &NewsSource{
		ID:       "test",
		Name:     "Test Source",
		Category: "Business",
	}

	feed := &RssFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				Title: "Article 1",
				Link:  "https://example.com/1",
			},
			{
				Title: "Article 2",
				Link:  "https://example.com/2",
			},
			{
				Title: "Article 3",
				Link:  "https://example.com/3",
			},
		},
	}

	articles := fetcher.parseRSSFeed(feed, source)

	if len(articles) != 3 {
		t.Errorf("Expected 3 articles, got %d", len(articles))
	}

	for _, art := range articles {
		if art.SourceName != "Test Source" {
			t.Errorf("All articles should have correct source")
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
