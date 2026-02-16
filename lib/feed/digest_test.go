package feed

import (
	"main/lib/article"
	"testing"
	"time"
)

func TestNewDigestBuilder(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)
	if builder == nil {
		t.Error("NewDigestBuilder() returned nil")
	}

	if builder.cache != cache {
		t.Error("Cache not properly set")
	}
	if builder.ranker != ranker {
		t.Error("Ranker not properly set")
	}
	if builder.summarizer != summarizer {
		t.Error("Summarizer not properly set")
	}
}

func TestBuildDailyDigestValidDate(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

	// Add test articles to cache
	articles := []article.ArticleData{
		{
			ID:            "test-1",
			Title:         "Article 1",
			URL:           "https://example.com/1",
			SourceID:      "igamingbusiness",
			PublishedDate: time.Now().Format(time.RFC3339),
			Summary:       "Summary 1",
		},
		{
			ID:            "test-2",
			Title:         "Article 2",
			URL:           "https://example.com/2",
			SourceID:      "sporttech",
			PublishedDate: time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
			Summary:       "Summary 2",
		},
	}

	cache.SetBatch(articles)

	today := time.Now().Format("2006-01-02")
	digest, err := builder.BuildDailyDigest(today)

	if err != nil {
		t.Fatalf("BuildDailyDigest() error = %v", err)
	}

	if digest == nil {
		t.Fatal("BuildDailyDigest() returned nil")
	}

	if digest.Date != today {
		t.Errorf("Digest date = %s, expected %s", digest.Date, today)
	}

	if len(digest.Articles) == 0 {
		t.Error("Digest should have articles")
	}

	if digest.Summary == "" {
		t.Error("Digest summary should not be empty")
	}
}

func TestBuildDailyDigestInvalidDate(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

	_, err := builder.BuildDailyDigest("not-a-date")
	if err == nil {
		t.Error("BuildDailyDigest() should error with invalid date")
	}
}

func TestBuildTodayDigest(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

	// Add test articles
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

	cache.SetBatch(articles)

	digest, err := builder.BuildTodayDigest()
	if err != nil {
		t.Fatalf("BuildTodayDigest() error = %v", err)
	}

	if digest == nil {
		t.Fatal("BuildTodayDigest() returned nil")
	}

	today := time.Now().Format("2006-01-02")
	if digest.Date != today {
		t.Errorf("Digest date = %s, expected %s", digest.Date, today)
	}
}

func TestBuildDigestFromArticlesTopN(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

	// Create 10 test articles
	articles := make([]article.ArticleData, 10)
	for i := 0; i < 10; i++ {
		articles[i] = article.ArticleData{
			ID:            string(rune(i)),
			Title:         "Article " + string(rune(i+'0')),
			URL:           "https://example.com/" + string(rune(i+'0')),
			SourceID:      "igamingbusiness",
			PublishedDate: time.Now().AddDate(0, 0, -i).Format(time.RFC3339),
			Summary:       "Summary " + string(rune(i+'0')),
		}
	}

	opts := &DigestOptions{
		TopN:           5,
		MinScore:       0.0,
		IncludeReasons: true,
	}

	digest, err := builder.BuildDigestFromArticles(articles, opts, time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("BuildDigestFromArticles() error = %v", err)
	}

	if len(digest.Articles) != 5 {
		t.Errorf("Digest should have 5 articles, got %d", len(digest.Articles))
	}

	// Verify ranks are 1-5
	for i, ranked := range digest.Articles {
		expectedRank := i + 1
		if ranked.Rank != expectedRank {
			t.Errorf("Article %d has rank %d, expected %d", i, ranked.Rank, expectedRank)
		}
	}
}

func TestBuildDigestFromArticlesMinScore(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

	articles := []article.ArticleData{
		{
			ID:            "test-1",
			Title:         "Article 1",
			URL:           "https://example.com/1",
			PublishedDate: time.Now().Format(time.RFC3339),
			Summary:       "Summary 1",
		},
		{
			ID:            "test-2",
			Title:         "Article 2",
			URL:           "https://example.com/2",
			PublishedDate: time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
			Summary:       "Summary 2",
		},
	}

	opts := &DigestOptions{
		TopN:           10,
		MinScore:       0.5,
		IncludeReasons: true,
	}

	digest, err := builder.BuildDigestFromArticles(articles, opts, time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("BuildDigestFromArticles() error = %v", err)
	}

	// Should only include high-scoring articles
	if len(digest.Articles) > 2 {
		t.Errorf("Digest should have at most 2 articles, got %d", len(digest.Articles))
	}
}

func TestBuildDigestEmptyArticles(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

	opts := &DigestOptions{
		TopN:           5,
		MinScore:       0.0,
		IncludeReasons: true,
	}

	digest, err := builder.BuildDigestFromArticles([]article.ArticleData{}, opts, time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("BuildDigestFromArticles() error = %v", err)
	}

	if len(digest.Articles) != 0 {
		t.Errorf("Digest should have 0 articles, got %d", len(digest.Articles))
	}
}

func TestBuildDigestNilOptions(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

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

	// Pass nil options - should use defaults
	digest, err := builder.BuildDigestFromArticles(articles, nil, time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("BuildDigestFromArticles() error = %v", err)
	}

	if digest == nil {
		t.Fatal("Digest should not be nil")
	}

	if len(digest.Articles) == 0 {
		t.Error("Digest should use default TopN=5")
	}
}

func TestFallbackDigestSummary(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)
	builder := NewDigestBuilder(cache, ranker, nil)

	articles := []article.RankedArticle{
		{
			Article: article.ArticleData{
				Title: "Article 1",
			},
			Rank: 1,
		},
		{
			Article: article.ArticleData{
				Title: "Article 2",
			},
			Rank: 2,
		},
	}

	summary := builder.fallbackDigestSummary(articles)

	if summary == "" {
		t.Error("fallbackDigestSummary() returned empty string")
	}

	if len(summary) < 10 {
		t.Error("fallbackDigestSummary() returned too short summary")
	}
}

func TestFallbackDigestSummaryEmpty(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)
	builder := NewDigestBuilder(cache, ranker, nil)

	summary := builder.fallbackDigestSummary([]article.RankedArticle{})

	if summary == "" {
		t.Error("fallbackDigestSummary() should return default message for empty articles")
	}
}

func TestDigestCreatedTimestamp(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

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

	cache.SetBatch(articles)

	digest, _ := builder.BuildTodayDigest()

	if digest.Created.IsZero() {
		t.Error("Digest Created timestamp should be set")
	}

	if digest.Created.After(time.Now().Add(time.Second)) {
		t.Error("Digest Created timestamp should not be in future")
	}
}

func TestBuildDigestFromArticlesDefaultTopN(t *testing.T) {
	cache := NewArticleCache(time.Hour, 100)
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)
	summarizer, _ := NewArticleSummarizer(&SummarizerConfig{
		APIKey: "test",
	})

	builder := NewDigestBuilder(cache, ranker, summarizer)

	// Create 10 articles
	articles := make([]article.ArticleData, 10)
	for i := 0; i < 10; i++ {
		articles[i] = article.ArticleData{
			ID:            string(rune(i)),
			Title:         "Article " + string(rune(i+'0')),
			URL:           "https://example.com/" + string(rune(i+'0')),
			SourceID:      "igamingbusiness",
			PublishedDate: time.Now().AddDate(0, 0, -i).Format(time.RFC3339),
			Summary:       "Summary " + string(rune(i+'0')),
		}
	}

	// Build with default options (nil)
	digest, err := builder.BuildDigestFromArticles(articles, nil, time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatalf("BuildDigestFromArticles() error = %v", err)
	}

	// Default TopN should be 5
	if len(digest.Articles) != 5 {
		t.Errorf("Default TopN should be 5, got %d articles", len(digest.Articles))
	}
}
