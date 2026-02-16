package feed

import (
	"fmt"
	"main/lib/article"
	"testing"
	"time"
)

// TestNewArticleCache tests cache initialization
func TestNewArticleCache(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	if cache == nil {
		t.Fatal("NewArticleCache should not return nil")
	}

	if cache.Size() != 0 {
		t.Errorf("New cache should be empty, got %d", cache.Size())
	}

	if cache.GetTTL() != 5*time.Minute {
		t.Errorf("Expected TTL 5m, got %v", cache.GetTTL())
	}
}

// TestCacheSetAndGet tests adding and retrieving articles
func TestCacheSetAndGet(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	article := article.ArticleData{
		ID:    "test-1",
		Title: "Test Article",
	}

	cache.Set(article)

	retrieved, found := cache.Get("test-1")
	if !found {
		t.Fatal("Article should be found in cache")
	}

	if retrieved.ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got '%s'", retrieved.ID)
	}

	if retrieved.Title != "Test Article" {
		t.Errorf("Expected title, got '%s'", retrieved.Title)
	}
}

// TestCacheGetNonexistent tests retrieving non-existent article
func TestCacheGetNonexistent(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	_, found := cache.Get("nonexistent")
	if found {
		t.Error("Non-existent article should not be found")
	}
}

// TestCacheExpiration tests cache entry expiration
func TestCacheExpiration(t *testing.T) {
	cache := NewArticleCache(100*time.Millisecond, 100)

	article := article.ArticleData{
		ID:    "test-expire",
		Title: "Test",
	}

	cache.Set(article)

	// Should be found immediately
	_, found := cache.Get("test-expire")
	if !found {
		t.Fatal("Article should be found immediately after add")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("test-expire")
	if found {
		t.Error("Expired article should not be found")
	}
}

// TestCacheSetBatch tests batch insertion
func TestCacheSetBatch(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	articles := []article.ArticleData{
		{ID: "batch-1", Title: "Article 1"},
		{ID: "batch-2", Title: "Article 2"},
		{ID: "batch-3", Title: "Article 3"},
	}

	cache.SetBatch(articles)

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}

	// Verify all articles are in cache
	for _, art := range articles {
		retrieved, found := cache.Get(art.ID)
		if !found {
			t.Errorf("Article %s should be in cache", art.ID)
		}

		if retrieved.Title != art.Title {
			t.Errorf("Expected title '%s', got '%s'", art.Title, retrieved.Title)
		}
	}
}

// TestCacheGetAll tests retrieving all articles
func TestCacheGetAll(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	articles := []article.ArticleData{
		{ID: "1", Title: "Article 1"},
		{ID: "2", Title: "Article 2"},
		{ID: "3", Title: "Article 3"},
	}

	cache.SetBatch(articles)

	all := cache.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 articles, got %d", len(all))
	}
}

// TestCacheRemove tests removing articles
func TestCacheRemove(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	article := article.ArticleData{
		ID:    "remove-test",
		Title: "Test",
	}

	cache.Set(article)

	if cache.Size() != 1 {
		t.Errorf("Cache should have 1 item before remove")
	}

	removed := cache.Remove("remove-test")
	if !removed {
		t.Error("Remove should return true for existing article")
	}

	if cache.Size() != 0 {
		t.Errorf("Cache should be empty after remove")
	}

	// Try removing non-existent
	removed = cache.Remove("nonexistent")
	if removed {
		t.Error("Remove should return false for non-existent article")
	}
}

// TestCacheClear tests clearing entire cache
func TestCacheClear(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	articles := []article.ArticleData{
		{ID: "1", Title: "Article 1"},
		{ID: "2", Title: "Article 2"},
		{ID: "3", Title: "Article 3"},
	}

	cache.SetBatch(articles)

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3")
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Cache should be empty after clear")
	}
}

// TestCacheClearExpired tests clearing expired entries
func TestCacheClearExpired(t *testing.T) {
	cache := NewArticleCache(100*time.Millisecond, 100)

	// Add articles
	for i := 0; i < 5; i++ {
		cache.Set(article.ArticleData{ID: string(rune(i))})
	}

	if cache.Size() != 5 {
		t.Errorf("Expected cache size 5")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Clear expired
	cleared := cache.ClearExpired()

	if cleared != 5 {
		t.Errorf("Expected to clear 5 expired items, cleared %d", cleared)
	}

	if cache.Size() != 0 {
		t.Errorf("Cache should be empty after clearing expired")
	}
}

// TestCacheGetBySource tests filtering by source
func TestCacheGetBySource(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	articles := []article.ArticleData{
		{ID: "1", Title: "Article 1", SourceName: "Source A"},
		{ID: "2", Title: "Article 2", SourceName: "Source B"},
		{ID: "3", Title: "Article 3", SourceName: "Source A"},
	}

	cache.SetBatch(articles)

	sourceA := cache.GetBySource("Source A")
	if len(sourceA) != 2 {
		t.Errorf("Expected 2 articles from Source A, got %d", len(sourceA))
	}

	sourceB := cache.GetBySource("Source B")
	if len(sourceB) != 1 {
		t.Errorf("Expected 1 article from Source B, got %d", len(sourceB))
	}
}

// TestCacheGetByCategory tests filtering by category
func TestCacheGetByCategory(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	articles := []article.ArticleData{
		{ID: "1", Title: "Article 1", Categories: []string{"Business", "Tech"}},
		{ID: "2", Title: "Article 2", Categories: []string{"Sports"}},
		{ID: "3", Title: "Article 3", Categories: []string{"Business"}},
	}

	cache.SetBatch(articles)

	business := cache.GetByCategory("Business")
	if len(business) != 2 {
		t.Errorf("Expected 2 business articles, got %d", len(business))
	}

	sports := cache.GetByCategory("Sports")
	if len(sports) != 1 {
		t.Errorf("Expected 1 sports article, got %d", len(sports))
	}
}

// TestCacheMaxSizeEviction tests cache size limit and eviction
func TestCacheMaxSizeEviction(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 3)

	// Add 3 articles (at max size)
	for i := 0; i < 3; i++ {
		cache.Set(article.ArticleData{ID: string(rune(i))})
	}

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}

	// Add 4th article (should evict oldest)
	cache.Set(article.ArticleData{ID: "3"})

	// Should still be at max size
	if cache.Size() != 3 {
		t.Errorf("Cache should not exceed max size, got %d", cache.Size())
	}
}

// TestCacheSetTTL tests updating TTL
func TestCacheSetTTL(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 100)

	if cache.GetTTL() != 5*time.Minute {
		t.Errorf("Expected TTL 5m")
	}

	cache.SetTTL(10 * time.Minute)

	if cache.GetTTL() != 10*time.Minute {
		t.Errorf("Expected TTL 10m after update")
	}
}

// TestCacheGetSize tests getting size information
func TestCacheGetSize(t *testing.T) {
	cache := NewArticleCache(100*time.Millisecond, 100)

	// Add some articles
	cache.SetBatch([]article.ArticleData{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	})

	valid, expired := cache.GetSize()

	if valid != 3 {
		t.Errorf("Expected 3 valid entries, got %d", valid)
	}

	if expired != 0 {
		t.Errorf("Expected 0 expired entries, got %d", expired)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	valid, expired = cache.GetSize()

	if valid != 0 {
		t.Errorf("Expected 0 valid after expiration, got %d", valid)
	}

	if expired != 3 {
		t.Errorf("Expected 3 expired, got %d", expired)
	}
}

// TestCacheGetStats tests cache statistics
func TestCacheGetStats(t *testing.T) {
	cache := NewArticleCache(5*time.Minute, 10)

	cache.SetBatch([]article.ArticleData{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	})

	stats := cache.GetStats()

	if stats.TotalEntries != 3 {
		t.Errorf("Expected 3 total entries, got %d", stats.TotalEntries)
	}

	if stats.ValidEntries != 3 {
		t.Errorf("Expected 3 valid entries, got %d", stats.ValidEntries)
	}

	if stats.MaxSize != 10 {
		t.Errorf("Expected max size 10, got %d", stats.MaxSize)
	}

	if stats.UtilizationPct != 30.0 {
		t.Errorf("Expected 30%% utilization, got %.1f%%", stats.UtilizationPct)
	}
}

// TestNewCacheManager tests cache manager creation
func TestNewCacheManager(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)

	if manager == nil {
		t.Fatal("NewCacheManager should not return nil")
	}

	if manager.articleCache == nil {
		t.Error("Article cache should be initialized")
	}

	if manager.sourceMetadata == nil {
		t.Error("Source metadata should be initialized")
	}
}

// TestCacheManagerArticleCaching tests caching through manager
func TestCacheManagerArticleCaching(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)

	articles := []article.ArticleData{
		{ID: "1", Title: "Article 1"},
		{ID: "2", Title: "Article 2"},
	}

	err := manager.CacheArticles(articles, "source-1")
	if err != nil {
		t.Fatalf("CacheArticles failed: %v", err)
	}

	cached := manager.GetCachedArticles()
	if len(cached) != 2 {
		t.Errorf("Expected 2 cached articles, got %d", len(cached))
	}
}

// TestCacheManagerEmptyArticles tests error for empty articles
func TestCacheManagerEmptyArticles(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)

	err := manager.CacheArticles([]article.ArticleData{}, "source-1")
	if err == nil {
		t.Error("CacheArticles should error for empty slice")
	}
}

// TestCacheManagerSourceMetadata tests source metadata tracking
func TestCacheManagerSourceMetadata(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)

	articles := []article.ArticleData{{ID: "1"}}
	manager.CacheArticles(articles, "source-1")

	metadata := manager.GetSourceMetadata("source-1")

	if metadata == nil {
		t.Fatal("Source metadata should be created")
	}

	if metadata.SourceID != "source-1" {
		t.Errorf("Expected source ID 'source-1', got '%s'", metadata.SourceID)
	}

	if metadata.FetchCount != 1 {
		t.Errorf("Expected fetch count 1, got %d", metadata.FetchCount)
	}
}

// TestCacheManagerRecordError tests error recording
func TestCacheManagerRecordError(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)

	testErr := fmt.Errorf("test error")
	manager.RecordSourceError("source-1", testErr)

	metadata := manager.GetSourceMetadata("source-1")
	if metadata == nil {
		t.Fatal("Source metadata should exist")
	}

	if metadata.ErrorCount != 1 {
		t.Errorf("Expected error count 1, got %d", metadata.ErrorCount)
	}

	if metadata.LastError == nil {
		t.Error("Last error should be recorded")
	}
}

// TestGetGlobalCacheManager tests singleton pattern
func TestGetGlobalCacheManager(t *testing.T) {
	manager1 := GetGlobalCacheManager(5*time.Minute, 100)
	manager2 := GetGlobalCacheManager(5*time.Minute, 100)

	if manager1 != manager2 {
		t.Error("Global cache manager should be singleton")
	}
}

// TestCacheManagerSetSummarizer tests summarizer registration
func TestCacheManagerSetSummarizer(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	manager.SetSummarizer(summarizer)

	// Verify summarizer is set
	manager.mu.RLock()
	if manager.summarizer == nil {
		t.Error("Summarizer should be set")
	}
	manager.mu.RUnlock()
}

// TestCacheManagerSetRankingEngine tests ranking engine registration
func TestCacheManagerSetRankingEngine(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)

	manager.SetRankingEngine(ranker)

	// Verify ranker is set
	manager.mu.RLock()
	if manager.rankingEngine == nil {
		t.Error("Ranking engine should be set")
	}
	manager.mu.RUnlock()
}

// TestCacheManagerSetDigestBuilder tests digest builder registration
func TestCacheManagerSetDigestBuilder(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})
	builder := NewDigestBuilder(cache, ranker, summarizer)

	manager.SetDigestBuilder(builder)

	// Verify builder is set
	manager.mu.RLock()
	if manager.digestBuilder == nil {
		t.Error("Digest builder should be set")
	}
	manager.mu.RUnlock()
}

// TestCacheManagerGetDailyDigest tests digest retrieval
func TestCacheManagerGetDailyDigest(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})
	builder := NewDigestBuilder(cache, ranker, summarizer)

	manager.SetDigestBuilder(builder)

	// Add articles to cache
	articles := []article.ArticleData{
		{
			ID:            "test-1",
			Title:         "Article 1",
			URL:           "https://example.com/1",
			SourceID:      "igamingbusiness",
			PublishedDate: time.Now().Format(time.RFC3339),
			Summary:       "Summary 1",
		},
	}
	manager.articleCache.SetBatch(articles)

	// Get digest
	today := time.Now().Format("2006-01-02")
	digest, err := manager.GetDailyDigest(today)

	if err != nil {
		t.Fatalf("GetDailyDigest() error = %v", err)
	}

	if digest == nil {
		t.Fatal("Digest should not be nil")
	}

	if digest.Date != today {
		t.Errorf("Digest date = %s, expected %s", digest.Date, today)
	}
}

// TestCacheManagerGetDailyDigestNoBuilder tests error when builder not set
func TestCacheManagerGetDailyDigestNoBuilder(t *testing.T) {
	manager := NewCacheManager(5*time.Minute, 100)

	_, err := manager.GetDailyDigest(time.Now().Format("2006-01-02"))
	if err == nil {
		t.Error("GetDailyDigest() should error when builder not configured")
	}
}
