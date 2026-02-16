package handler

import (
	"context"
	"main/lib/article"
	"main/lib/feed"
	"main/lib/logger"
	"main/lib/middleware"
	"net/http"
	"os"
	"time"
)

// Handler generates and returns daily digests with top 5 ranked articles
func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)
	logger.LogRequestStart(r)

	// Get date from query parameter, default to today
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		logger.Error("Invalid date format", err, map[string]interface{}{
			"date": dateStr,
		})
		middleware.WriteJSONError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// Initialize cache manager
	cacheManager := feed.GetGlobalCacheManager(24*time.Hour, 5000)

	// Get Claude API key from environment
	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")
	if claudeAPIKey == "" {
		// If no API key, return digest with articles but no AI summaries
		logger.Warn("CLAUDE_API_KEY not set, digest will use fallback summaries", ctx)
	}

	// Initialize components only if API key is available
	var summarizer *feed.ArticleSummarizer
	var digestBuilder *feed.DigestBuilder
	var ranker *feed.RankingEngine

	if claudeAPIKey != "" {
		summarizerConfig := &feed.SummarizerConfig{
			APIKey:      claudeAPIKey,
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   150,
			Temperature: 0.7,
			TimeoutSec:  30,
		}

		var err error
		summarizer, err = feed.NewArticleSummarizer(summarizerConfig)
		if err != nil {
			logger.Warn("Failed to initialize summarizer", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Initialize ranking engine
	criteria := article.NewRankingCriteria()
	sourceMgr := feed.NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker = feed.NewRankingEngine(criteria, sourceMgr)

	// Initialize digest builder
	cache := feed.NewArticleCache(24*time.Hour, 5000)
	if summarizer != nil {
		digestBuilder = feed.NewDigestBuilder(cache, ranker, summarizer)
	} else {
		digestBuilder = feed.NewDigestBuilder(cache, ranker, nil)
	}

	// Get articles - for now, create sample articles
	// In a real implementation, this would fetch from the fetcher
	articles := getSampleArticles()

	// Add articles to cache
	cache.SetBatch(articles)

	// Enhance articles with summaries if summarizer is available
	if summarizer != nil {
		enhanceCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := summarizer.SummarizeBatch(enhanceCtx, articles); err != nil {
			logger.Warn("Failed to summarize articles", map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Update cache with enhanced articles
		cache.SetBatch(articles)
	}

	// Build digest
	digest, err := digestBuilder.BuildDailyDigest(dateStr)
	if err != nil {
		logger.Error("Failed to build digest", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to generate digest")
		return
	}

	// Return digest as JSON
	middleware.WriteJSONSuccess(w, http.StatusOK, digest)
}

// getSampleArticles returns sample articles for demonstration
// In production, this would fetch real articles from the feed
func getSampleArticles() []article.ArticleData {
	return []article.ArticleData{
		{
			ID:            "article-001",
			Title:         "New iGaming Regulations Announced in UK",
			OriginalSum:   "UK gambling authority unveils stricter regulations for online gaming operators",
			URL:           "https://example.com/uk-regulations",
			SourceName:    "iGamingBusiness",
			SourceID:      "igamingbusiness",
			PublishedDate: time.Now().Format(time.RFC3339),
			Categories:    []string{"Regulations", "United Kingdom"},
			Metadata: map[string]interface{}{
				"views": 1500.0,
				"shares": 45.0,
			},
		},
		{
			ID:            "article-002",
			Title:         "Sports Betting Market Growth Exceeds Expectations",
			OriginalSum:   "Latest data shows sports betting segment growing faster than predicted",
			URL:           "https://example.com/sports-betting-growth",
			SourceName:    "Gambling Insider",
			SourceID:      "gamblinginsider",
			PublishedDate: time.Now().Add(-6 * time.Hour).Format(time.RFC3339),
			Categories:    []string{"Business", "Sports Betting"},
			Metadata: map[string]interface{}{
				"views": 2100.0,
				"shares": 67.0,
			},
		},
		{
			ID:            "article-003",
			Title:         "AI-Powered Responsible Gaming Tools Launch",
			OriginalSum:   "Technology companies deploy machine learning for player protection",
			URL:           "https://example.com/ai-responsible-gaming",
			SourceName:    "eGaming Review",
			SourceID:      "eganingreview",
			PublishedDate: time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			Categories:    []string{"Technology", "Responsible Gaming"},
			Metadata: map[string]interface{}{
				"views": 890.0,
				"shares": 34.0,
			},
		},
		{
			ID:            "article-004",
			Title:         "Merger: Two Major Gaming Operators Combine",
			OriginalSum:   "Strategic acquisition creates largest unified gaming platform",
			URL:           "https://example.com/gaming-merger",
			SourceName:    "Sportech",
			SourceID:      "sporttech",
			PublishedDate: time.Now().Add(-18 * time.Hour).Format(time.RFC3339),
			Categories:    []string{"M&A", "Business"},
			Metadata: map[string]interface{}{
				"views": 3200.0,
				"shares": 120.0,
			},
		},
		{
			ID:            "article-005",
			Title:         "Payment Processing: New Standards Established",
			OriginalSum:   "Industry sets new security and speed benchmarks for transactions",
			URL:           "https://example.com/payment-standards",
			SourceName:    "Betting Industry",
			SourceID:      "betindustry",
			PublishedDate: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			Categories:    []string{"Payments", "Technology"},
			Metadata: map[string]interface{}{
				"views": 750.0,
				"shares": 28.0,
			},
		},
		{
			ID:            "article-006",
			Title:         "Asian Markets Show Strong Gaming Adoption",
			OriginalSum:   "Emerging markets in Asia driving significant industry growth",
			URL:           "https://example.com/asia-gaming",
			SourceName:    "iGamingBusiness",
			SourceID:      "igamingbusiness",
			PublishedDate: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			Categories:    []string{"International", "Business"},
			Metadata: map[string]interface{}{
				"views": 1800.0,
				"shares": 52.0,
			},
		},
		{
			ID:            "article-007",
			Title:         "Compliance Tools Market Expands Rapidly",
			OriginalSum:   "Regulatory technology vendors report record growth in Q1",
			URL:           "https://example.com/compliance-market",
			SourceName:    "Gambling Insider",
			SourceID:      "gamblinginsider",
			PublishedDate: time.Now().Add(-8 * time.Hour).Format(time.RFC3339),
			Categories:    []string{"Technology", "Regulations"},
			Metadata: map[string]interface{}{
				"views": 1200.0,
				"shares": 41.0,
			},
		},
		{
			ID:            "article-008",
			Title:         "Player Protection: New Research Findings",
			OriginalSum:   "Study reveals impact of responsible gaming interventions",
			URL:           "https://example.com/player-protection",
			SourceName:    "eGaming Review",
			SourceID:      "eganingreview",
			PublishedDate: time.Now().Add(-20 * time.Hour).Format(time.RFC3339),
			Categories:    []string{"Responsible Gaming", "Business"},
			Metadata: map[string]interface{}{
				"views": 950.0,
				"shares": 37.0,
			},
		},
	}
}
