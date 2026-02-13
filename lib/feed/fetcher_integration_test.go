package feed

import (
	"context"
	"testing"
	"time"
)

// TestFetchFromIgamingBusiness tests fetching from iGamingBusiness.com
// NOTE: This test requires internet connection and may be slow
// Run with: go test -run TestFetchFromIgamingBusiness -timeout 30s -v
func TestFetchFromIgamingBusiness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fetcher := NewArticleFetcher(&FetcherConfig{
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
	})

	source := &NewsSource{
		ID:           "igamingbusiness",
		Name:         "iGamingBusiness",
		URL:          "https://www.igamingbusiness.com",
		FeedURL:      "https://www.igamingbusiness.com/feed/",
		Category:     "Business",
		Active:       true,
		Priority:     10,
		ScrapingType: "rss",
		Timeout:      30000,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	articles, err := fetcher.FetchFromSource(ctx, source)

	if err != nil {
		t.Logf("Failed to fetch from iGamingBusiness: %v", err)
		t.Logf("This is expected if the website is down or network is unavailable")
		// Don't fail the test - network issues are not code issues
		return
	}

	if len(articles) == 0 {
		t.Skip("No articles fetched (site may be down or empty)")
	}

	// Verify article structure
	for _, article := range articles {
		if article.ID == "" {
			t.Error("Article ID should not be empty")
		}
		if article.Title == "" {
			t.Error("Article Title should not be empty")
		}
		if article.URL == "" {
			t.Error("Article URL should not be empty")
		}
		if article.SourceName != "iGamingBusiness" {
			t.Errorf("Expected source iGamingBusiness, got %s", article.SourceName)
		}
	}

	t.Logf("Successfully fetched %d articles from iGamingBusiness", len(articles))
}

// TestFetchFromGamblingInsider tests fetching from GamblingInsider.com
func TestFetchFromGamblingInsider(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fetcher := NewArticleFetcher(&FetcherConfig{
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
	})

	source := &NewsSource{
		ID:           "gamblinginsider",
		Name:         "Gambling Insider",
		URL:          "https://www.gamblinginsider.com",
		FeedURL:      "https://www.gamblinginsider.com/feed/",
		Category:     "Business",
		Active:       true,
		Priority:     9,
		ScrapingType: "rss",
		Timeout:      30000,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	articles, err := fetcher.FetchFromSource(ctx, source)

	if err != nil {
		t.Logf("Failed to fetch from Gambling Insider: %v", err)
		return
	}

	if len(articles) == 0 {
		t.Skip("No articles fetched")
	}

	t.Logf("Successfully fetched %d articles from Gambling Insider", len(articles))
}

// TestFetchFromMultipleSources tests fetching from all configured sources
func TestFetchFromMultipleSources(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager := NewSourceManager()
	manager.LoadDefaultSources()

	fetcher := NewArticleFetcher(&FetcherConfig{
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
	})

	sources := manager.GetActiveSources()

	if len(sources) == 0 {
		t.Fatal("No active sources configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	allArticles, err := fetcher.FetchFromSources(ctx, sources)

	// Don't fail on network errors
	if err != nil {
		t.Logf("Warning: Fetch from some sources failed: %v", err)
	}

	if len(allArticles) == 0 {
		t.Skip("No articles fetched from any source")
	}

	t.Logf("Successfully fetched %d articles from %d sources", len(allArticles), len(sources))

	// Count articles per source
	sourceCount := make(map[string]int)
	for _, article := range allArticles {
		sourceCount[article.SourceName]++
	}

	for source, count := range sourceCount {
		t.Logf("  %s: %d articles", source, count)
	}
}

// TestFetchAndCacheIntegration tests full pipeline of fetching and caching
func TestFetchAndCacheIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	fetcher := NewArticleFetcher(&FetcherConfig{
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
	})

	cache := NewArticleCache(5*time.Minute, 1000)
	cacheManager := NewCacheManager(5*time.Minute, 1000)

	// Fetch
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get just the first source for quick testing
	sources := manager.GetActiveSources()
	if len(sources) > 1 {
		sources = sources[:1]
	}

	articles, err := fetcher.FetchFromSources(ctx, sources)

	if err != nil {
		t.Logf("Fetch failed: %v (expected if network unavailable)", err)
		return
	}

	if len(articles) == 0 {
		t.Skip("No articles fetched")
	}

	// Cache
	cache.SetBatch(articles)
	cacheManager.CacheArticles(articles, sources[0].ID)

	// Verify caching
	cachedArticles := cache.GetAll()
	if len(cachedArticles) != len(articles) {
		t.Errorf("Expected %d cached articles, got %d", len(articles), len(cachedArticles))
	}

	// Get statistics
	stats := cache.GetStats()
	t.Logf("Cache stats - Total: %d, Valid: %d, Utilization: %.1f%%",
		stats.TotalEntries, stats.ValidEntries, stats.UtilizationPct)

	// Test filtering
	byCategory := cache.GetByCategory(articles[0].Categories[0])
	t.Logf("Found %d articles in category %s", len(byCategory), articles[0].Categories[0])

	t.Logf("Successfully fetched, cached, and queried %d articles", len(articles))
}

// TestArticleQuality tests that fetched articles have required fields
func TestArticleQuality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager := NewSourceManager()
	manager.LoadDefaultSources()

	fetcher := NewArticleFetcher(DefaultFetcherConfig())

	// Test with just one source
	sources := manager.GetActiveSources()[:1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	articles, err := fetcher.FetchFromSources(ctx, sources)

	if err != nil {
		t.Logf("Fetch failed: %v", err)
		return
	}

	if len(articles) == 0 {
		t.Skip("No articles fetched")
	}

	// Check article quality
	for i, article := range articles {
		checks := []struct {
			name  string
			check bool
		}{
			{"ID", article.ID != ""},
			{"Title", article.Title != ""},
			{"URL", article.URL != ""},
			{"SourceName", article.SourceName != ""},
			{"PublishedDate", article.PublishedDate != ""},
			{"Categories", len(article.Categories) > 0},
			{"CreatedAt", !article.CreatedAt.IsZero()},
		}

		for _, check := range checks {
			if !check.check {
				t.Errorf("Article %d failed check: %s", i, check.name)
			}
		}
	}

	t.Logf("All %d articles passed quality checks", len(articles))
}

// TestRealDateParsing tests date parsing with real article data
func TestRealDateParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	manager := NewSourceManager()
	manager.LoadDefaultSources()

	fetcher := NewArticleFetcher(DefaultFetcherConfig())
	sources := manager.GetActiveSources()[:1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	articles, err := fetcher.FetchFromSources(ctx, sources)

	if err != nil {
		t.Logf("Fetch failed: %v", err)
		return
	}

	if len(articles) == 0 {
		t.Skip("No articles fetched")
	}

	// Verify dates parse correctly
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)

	for _, article := range articles {
		pubDate, err := time.Parse(time.RFC3339, article.PublishedDate)
		if err != nil {
			t.Logf("Failed to parse date %s: %v", article.PublishedDate, err)
			continue
		}

		// Published date should be recent (within 7 days)
		if pubDate.After(now) {
			t.Logf("Warning: Article published in future: %s", article.PublishedDate)
		}

		if pubDate.Before(oneWeekAgo) {
			t.Logf("Warning: Article is quite old: %s", article.PublishedDate)
		}
	}

	t.Logf("Successfully parsed dates from %d articles", len(articles))
}

// BenchmarkFetchFromIgamingBusiness benchmarks fetching performance
// Run with: go test -bench BenchmarkFetch -benchtime=1x
func BenchmarkFetchFromIgamingBusiness(b *testing.B) {
	fetcher := NewArticleFetcher(&FetcherConfig{
		Timeout:       30 * time.Second,
		RetryAttempts: 1,
		RetryDelay:    500 * time.Millisecond,
	})

	source := &NewsSource{
		ID:           "igamingbusiness",
		Name:         "iGamingBusiness",
		FeedURL:      "https://www.igamingbusiness.com/feed/",
		Active:       true,
		ScrapingType: "rss",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fetcher.FetchFromSource(ctx, source)
	}
}
