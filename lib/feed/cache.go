package feed

import (
	"context"
	"fmt"
	"main/lib/article"
	"sync"
	"time"
)

// CacheEntry represents a cached article with metadata
type CacheEntry struct {
	Article   article.ArticleData
	Timestamp time.Time
	ExpiresAt time.Time
}

// ArticleCache provides in-memory caching for articles
type ArticleCache struct {
	mu       sync.RWMutex
	articles map[string]*CacheEntry
	ttl      time.Duration
	maxSize  int
}

// NewArticleCache creates a new article cache
func NewArticleCache(ttl time.Duration, maxSize int) *ArticleCache {
	return &ArticleCache{
		articles: make(map[string]*CacheEntry),
		ttl:      ttl,
		maxSize:  maxSize,
	}
}

// Set adds or updates an article in the cache
func (ac *ArticleCache) Set(article article.ArticleData) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Check cache size limit
	if len(ac.articles) >= ac.maxSize && ac.articles[article.ID] == nil {
		// Remove oldest entry to make room
		ac.evictOldest()
	}

	ac.articles[article.ID] = &CacheEntry{
		Article:   article,
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(ac.ttl),
	}

	return nil
}

// Get retrieves an article from the cache
func (ac *ArticleCache) Get(id string) (*article.ArticleData, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	entry, exists := ac.articles[id]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return &entry.Article, true
}

// SetBatch adds multiple articles to the cache
func (ac *ArticleCache) SetBatch(articles []article.ArticleData) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	for _, art := range articles {
		ac.articles[art.ID] = &CacheEntry{
			Article:   art,
			Timestamp: time.Now(),
			ExpiresAt: time.Now().Add(ac.ttl),
		}
	}

	return nil
}

// GetAll retrieves all non-expired articles from the cache
func (ac *ArticleCache) GetAll() []article.ArticleData {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	var articles []article.ArticleData
	now := time.Now()

	for _, entry := range ac.articles {
		if now.Before(entry.ExpiresAt) {
			articles = append(articles, entry.Article)
		}
	}

	return articles
}

// GetBySource retrieves articles from a specific source
func (ac *ArticleCache) GetBySource(sourceName string) []article.ArticleData {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	var articles []article.ArticleData
	now := time.Now()

	for _, entry := range ac.articles {
		if entry.Article.SourceName == sourceName && now.Before(entry.ExpiresAt) {
			articles = append(articles, entry.Article)
		}
	}

	return articles
}

// GetByCategory retrieves articles in a specific category
func (ac *ArticleCache) GetByCategory(category string) []article.ArticleData {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	var articles []article.ArticleData
	now := time.Now()

	for _, entry := range ac.articles {
		// Check if article has this category
		hasCategory := false
		for _, cat := range entry.Article.Categories {
			if cat == category {
				hasCategory = true
				break
			}
		}

		if hasCategory && now.Before(entry.ExpiresAt) {
			articles = append(articles, entry.Article)
		}
	}

	return articles
}

// Remove deletes an article from the cache
func (ac *ArticleCache) Remove(id string) bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if _, exists := ac.articles[id]; exists {
		delete(ac.articles, id)
		return true
	}

	return false
}

// Clear removes all articles from the cache
func (ac *ArticleCache) Clear() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.articles = make(map[string]*CacheEntry)
}

// ClearExpired removes all expired articles from the cache
func (ac *ArticleCache) ClearExpired() int {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	now := time.Now()
	count := 0

	for id, entry := range ac.articles {
		if now.After(entry.ExpiresAt) {
			delete(ac.articles, id)
			count++
		}
	}

	return count
}

// Size returns the number of articles in the cache
func (ac *ArticleCache) Size() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	return len(ac.articles)
}

// GetSize returns the number of articles and expired count
func (ac *ArticleCache) GetSize() (valid int, expired int) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	now := time.Now()

	for _, entry := range ac.articles {
		if now.Before(entry.ExpiresAt) {
			valid++
		} else {
			expired++
		}
	}

	return
}

// evictOldest removes the oldest entry from the cache
func (ac *ArticleCache) evictOldest() {
	var oldestID string
	var oldestTime time.Time

	for id, entry := range ac.articles {
		if oldestTime.IsZero() || entry.Timestamp.Before(oldestTime) {
			oldestID = id
			oldestTime = entry.Timestamp
		}
	}

	if oldestID != "" {
		delete(ac.articles, oldestID)
	}
}

// SetTTL updates the TTL for new entries
func (ac *ArticleCache) SetTTL(ttl time.Duration) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.ttl = ttl
}

// GetTTL returns the current TTL
func (ac *ArticleCache) GetTTL() time.Duration {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	return ac.ttl
}

// Stats returns cache statistics
type CacheStats struct {
	TotalEntries  int
	ValidEntries  int
	ExpiredCount  int
	OldestEntry   time.Time
	NewestEntry   time.Time
	AverageTTL    time.Duration
	MaxSize       int
	CurrentSize   int
	UtilizationPct float64
}

// GetStats returns cache statistics
func (ac *ArticleCache) GetStats() CacheStats {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	stats := CacheStats{
		TotalEntries: len(ac.articles),
		MaxSize:      ac.maxSize,
	}

	if ac.maxSize > 0 {
		stats.UtilizationPct = float64(len(ac.articles)) / float64(ac.maxSize) * 100
	}

	if len(ac.articles) == 0 {
		return stats
	}

	now := time.Now()
	var totalTTL time.Duration
	var validCount int

	for _, entry := range ac.articles {
		if now.Before(entry.ExpiresAt) {
			validCount++
		} else {
			stats.ExpiredCount++
		}

		if stats.OldestEntry.IsZero() || entry.Timestamp.Before(stats.OldestEntry) {
			stats.OldestEntry = entry.Timestamp
		}

		if entry.Timestamp.After(stats.NewestEntry) {
			stats.NewestEntry = entry.Timestamp
		}

		remaining := entry.ExpiresAt.Sub(now)
		if remaining > 0 {
			totalTTL += remaining
		}
	}

	stats.ValidEntries = validCount
	stats.CurrentSize = len(ac.articles)

	if validCount > 0 {
		stats.AverageTTL = totalTTL / time.Duration(validCount)
	}

	return stats
}

// SourceCache manages per-source caching metadata
type SourceCache struct {
	SourceID      string
	LastFetchTime time.Time
	FetchCount    int
	ErrorCount    int
	LastError     error
	CacheHits     int
	CacheMisses   int
}

// CacheManager manages article caching across sources
type CacheManager struct {
	mu             sync.RWMutex
	articleCache   *ArticleCache
	sourceMetadata map[string]*SourceCache
	summarizer     *ArticleSummarizer
	rankingEngine  *RankingEngine
	digestBuilder  *DigestBuilder
}

// NewCacheManager creates a new cache manager
func NewCacheManager(ttl time.Duration, maxSize int) *CacheManager {
	return &CacheManager{
		articleCache:   NewArticleCache(ttl, maxSize),
		sourceMetadata: make(map[string]*SourceCache),
	}
}

// CacheArticles caches articles from a source
func (cm *CacheManager) CacheArticles(articles []article.ArticleData, sourceID string) error {
	if len(articles) == 0 {
		return fmt.Errorf("no articles to cache")
	}

	// Update source metadata
	cm.mu.Lock()
	if _, exists := cm.sourceMetadata[sourceID]; !exists {
		cm.sourceMetadata[sourceID] = &SourceCache{
			SourceID: sourceID,
		}
	}
	sourceMeta := cm.sourceMetadata[sourceID]
	sourceMeta.LastFetchTime = time.Now()
	sourceMeta.FetchCount++
	cm.mu.Unlock()

	// Cache articles
	return cm.articleCache.SetBatch(articles)
}

// GetCachedArticles retrieves cached articles
func (cm *CacheManager) GetCachedArticles() []article.ArticleData {
	return cm.articleCache.GetAll()
}

// GetSourceMetadata retrieves metadata for a source
func (cm *CacheManager) GetSourceMetadata(sourceID string) *SourceCache {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.sourceMetadata[sourceID]
}

// RecordSourceError records an error for a source
func (cm *CacheManager) RecordSourceError(sourceID string, err error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.sourceMetadata[sourceID]; !exists {
		cm.sourceMetadata[sourceID] = &SourceCache{
			SourceID: sourceID,
		}
	}

	cm.sourceMetadata[sourceID].ErrorCount++
	cm.sourceMetadata[sourceID].LastError = err
}

// GetCacheManager returns global cache manager (singleton pattern)
var globalCacheManager *CacheManager
var cacheMutex sync.Mutex

// GetGlobalCacheManager returns the global cache manager instance
func GetGlobalCacheManager(ttl time.Duration, maxSize int) *CacheManager {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if globalCacheManager == nil {
		globalCacheManager = NewCacheManager(ttl, maxSize)
	}

	return globalCacheManager
}

// SetSummarizer registers a summarizer for article enhancement
func (cm *CacheManager) SetSummarizer(s *ArticleSummarizer) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.summarizer = s
}

// SetRankingEngine registers a ranking engine
func (cm *CacheManager) SetRankingEngine(r *RankingEngine) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.rankingEngine = r
}

// SetDigestBuilder registers a digest builder
func (cm *CacheManager) SetDigestBuilder(d *DigestBuilder) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.digestBuilder = d
}

// EnhanceArticles summarizes and ranks articles in batch
func (cm *CacheManager) EnhanceArticles(ctx context.Context, articles []article.ArticleData) error {
	cm.mu.RLock()
	summarizer := cm.summarizer
	cm.mu.RUnlock()

	if summarizer == nil {
		return fmt.Errorf("summarizer not configured")
	}

	// Summarize articles
	return summarizer.SummarizeBatch(ctx, articles)
}

// GetDailyDigest builds and returns a daily digest
func (cm *CacheManager) GetDailyDigest(date string) (*article.DailyDigest, error) {
	cm.mu.RLock()
	digestBuilder := cm.digestBuilder
	cm.mu.RUnlock()

	if digestBuilder == nil {
		return nil, fmt.Errorf("digest builder not configured")
	}

	return digestBuilder.BuildDailyDigest(date)
}
