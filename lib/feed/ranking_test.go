package feed

import (
	"main/lib/article"
	"testing"
	"time"
)

func TestNewRankingEngine(t *testing.T) {
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()

	ranker := NewRankingEngine(criteria, sourceMgr)
	if ranker == nil {
		t.Error("NewRankingEngine() returned nil")
	}

	if ranker.criteria != criteria {
		t.Error("Criteria not properly set")
	}
}

func TestNewRankingEngineNilCriteria(t *testing.T) {
	sourceMgr := NewSourceManager()
	ranker := NewRankingEngine(nil, sourceMgr)

	if ranker == nil {
		t.Error("NewRankingEngine() should create default criteria")
	}

	if ranker.criteria == nil {
		t.Error("Criteria should not be nil")
	}
}

func TestCalculateRecencyScore(t *testing.T) {
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)

	tests := []struct {
		name        string
		publishDate string
		minScore    float64
		maxScore    float64
	}{
		{
			name:        "today's article",
			publishDate: time.Now().Format(time.RFC3339),
			minScore:    0.95, // Should be very high
			maxScore:    1.0,
		},
		{
			name:        "article from 7 days ago",
			publishDate: time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
			minScore:    0.0001,
			maxScore:    0.01, // Exponential decay: exp(-7) ≈ 0.0009
		},
		{
			name:        "empty date",
			publishDate: "",
			minScore:    0.4,
			maxScore:    0.6, // Neutral score
		},
		{
			name:        "invalid date",
			publishDate: "not-a-date",
			minScore:    0.4,
			maxScore:    0.6, // Neutral score
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := ranker.calculateRecencyScore(tt.publishDate)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateRecencyScore() = %f, expected between %f and %f", score, tt.minScore, tt.maxScore)
			}

			// Ensure score is in [0, 1]
			if score < 0 || score > 1 {
				t.Errorf("Score out of bounds: %f", score)
			}
		})
	}
}

func TestCalculateSourceScore(t *testing.T) {
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)

	tests := []struct {
		name      string
		sourceID  string
		minScore  float64
		maxScore  float64
	}{
		{
			name:     "igamingbusiness (priority 10)",
			sourceID: "igamingbusiness",
			minScore: 0.99,
			maxScore: 1.0,
		},
		{
			name:     "sportech (priority 7)",
			sourceID: "sporttech",
			minScore: 0.65,
			maxScore: 0.75,
		},
		{
			name:     "unknown source",
			sourceID: "unknown",
			minScore: 0.4,
			maxScore: 0.6, // Neutral score
		},
		{
			name:     "empty source ID",
			sourceID: "",
			minScore: 0.4,
			maxScore: 0.6, // Neutral score
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := ranker.calculateSourceScore(tt.sourceID)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateSourceScore() = %f, expected between %f and %f", score, tt.minScore, tt.maxScore)
			}

			// Ensure score is in [0, 1]
			if score < 0 || score > 1 {
				t.Errorf("Score out of bounds: %f", score)
			}
		})
	}
}

func TestCalculateEngagementScore(t *testing.T) {
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)

	tests := []struct {
		name      string
		metadata  map[string]interface{}
		minScore  float64
		maxScore  float64
	}{
		{
			name:     "with views",
			metadata: map[string]interface{}{"views": 500.0},
			minScore: 0.4,
			maxScore: 0.6,
		},
		{
			name:     "with high engagement",
			metadata: map[string]interface{}{"engagement_score": 950.0},
			minScore: 0.9,
			maxScore: 1.0,
		},
		{
			name:     "no metadata",
			metadata: nil,
			minScore: 0.4,
			maxScore: 0.6, // Neutral score
		},
		{
			name:     "empty metadata",
			metadata: map[string]interface{}{},
			minScore: 0.4,
			maxScore: 0.6, // Neutral score
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			art := &article.ArticleData{
				ID:       "test",
				Title:    "Test",
				URL:      "https://example.com",
				Metadata: tt.metadata,
			}

			score := ranker.calculateEngagementScore(art)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateEngagementScore() = %f, expected between %f and %f", score, tt.minScore, tt.maxScore)
			}

			// Ensure score is in [0, 1]
			if score < 0 || score > 1 {
				t.Errorf("Score out of bounds: %f", score)
			}
		})
	}
}

func TestCalculateCategoryScore(t *testing.T) {
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)

	art := &article.ArticleData{
		ID:         "test",
		Title:      "Test",
		URL:        "https://example.com",
		Categories: []string{"Regulations"},
	}

	score := ranker.calculateCategoryScore(art)

	if score < 0 || score > 1 {
		t.Errorf("calculateCategoryScore() out of bounds: %f", score)
	}
}

func TestAssignReason(t *testing.T) {
	tests := []struct {
		name          string
		breakdown     *ScoreBreakdown
		expectedReason string
	}{
		{
			name: "trending article",
			breakdown: &ScoreBreakdown{
				RecencyScore: 0.95,
				SourceScore: 0.5,
				EngagementScore: 0.5,
				CategoryScore: 0.5,
			},
			expectedReason: "trending",
		},
		{
			name: "authoritative source",
			breakdown: &ScoreBreakdown{
				RecencyScore: 0.5,
				SourceScore: 0.95,
				EngagementScore: 0.5,
				CategoryScore: 0.5,
			},
			expectedReason: "authoritative",
		},
		{
			name: "high engagement",
			breakdown: &ScoreBreakdown{
				RecencyScore: 0.5,
				SourceScore: 0.5,
				EngagementScore: 0.95,
				CategoryScore: 0.5,
			},
			expectedReason: "high-engagement",
		},
		{
			name: "diverse category",
			breakdown: &ScoreBreakdown{
				RecencyScore: 0.5,
				SourceScore: 0.5,
				EngagementScore: 0.5,
				CategoryScore: 0.7,
			},
			expectedReason: "diverse",
		},
		{
			name: "multiple reasons",
			breakdown: &ScoreBreakdown{
				RecencyScore: 0.95,
				SourceScore: 0.95,
				EngagementScore: 0.5,
				CategoryScore: 0.5,
			},
			expectedReason: "trending, authoritative",
		},
		{
			name: "no special reasons",
			breakdown: &ScoreBreakdown{
				RecencyScore: 0.3,
				SourceScore: 0.3,
				EngagementScore: 0.3,
				CategoryScore: 0.3,
			},
			expectedReason: "featured",
		},
	}

	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := ranker.assignReason(tt.breakdown)
			if reason == "" {
				t.Error("assignReason() returned empty string")
			}
			if reason != tt.expectedReason {
				t.Errorf("assignReason() = %s, expected %s", reason, tt.expectedReason)
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)

	art := &article.ArticleData{
		ID:            "test-1",
		Title:         "Test Article",
		URL:           "https://example.com",
		SourceName:    "iGamingBusiness",
		SourceID:      "igamingbusiness",
		PublishedDate: time.Now().Format(time.RFC3339),
		Categories:    []string{"Regulations"},
	}

	breakdown, err := ranker.CalculateScore(art)
	if err != nil {
		t.Fatalf("CalculateScore() error = %v", err)
	}

	if breakdown == nil {
		t.Fatal("CalculateScore() returned nil breakdown")
	}

	// Verify all scores are in [0, 1]
	if breakdown.RecencyScore < 0 || breakdown.RecencyScore > 1 {
		t.Errorf("RecencyScore out of bounds: %f", breakdown.RecencyScore)
	}
	if breakdown.SourceScore < 0 || breakdown.SourceScore > 1 {
		t.Errorf("SourceScore out of bounds: %f", breakdown.SourceScore)
	}
	if breakdown.EngagementScore < 0 || breakdown.EngagementScore > 1 {
		t.Errorf("EngagementScore out of bounds: %f", breakdown.EngagementScore)
	}
	if breakdown.CategoryScore < 0 || breakdown.CategoryScore > 1 {
		t.Errorf("CategoryScore out of bounds: %f", breakdown.CategoryScore)
	}
	if breakdown.FinalScore < 0 || breakdown.FinalScore > 1 {
		t.Errorf("FinalScore out of bounds: %f", breakdown.FinalScore)
	}

	if breakdown.Reason == "" {
		t.Error("Reason should not be empty")
	}
}

func TestCalculateScoreNilArticle(t *testing.T) {
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)

	_, err := ranker.CalculateScore(nil)
	if err == nil {
		t.Error("CalculateScore() should error with nil article")
	}
}

func TestRankArticles(t *testing.T) {
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)

	articles := []article.ArticleData{
		{
			ID:            "test-1",
			Title:         "Article 1",
			URL:           "https://example.com/1",
			SourceID:      "igamingbusiness",
			PublishedDate: time.Now().Format(time.RFC3339),
		},
		{
			ID:            "test-2",
			Title:         "Article 2",
			URL:           "https://example.com/2",
			SourceID:      "sporttech",
			PublishedDate: time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
		},
		{
			ID:            "test-3",
			Title:         "Article 3",
			URL:           "https://example.com/3",
			SourceID:      "eganingreview",
			PublishedDate: time.Now().AddDate(0, 0, -2).Format(time.RFC3339),
		},
	}

	ranked, err := ranker.RankArticles(articles)
	if err != nil {
		t.Fatalf("RankArticles() error = %v", err)
	}

	if len(ranked) != len(articles) {
		t.Errorf("RankArticles() returned %d articles, expected %d", len(ranked), len(articles))
	}

	// Verify rankings are sequential
	for i, r := range ranked {
		expectedRank := i + 1
		if r.Rank != expectedRank {
			t.Errorf("Article %d has rank %d, expected %d", i, r.Rank, expectedRank)
		}
	}

	// Verify scores are in descending order
	for i := 0; i < len(ranked)-1; i++ {
		if ranked[i].Score < ranked[i+1].Score {
			t.Errorf("Articles not sorted by score: %f > %f", ranked[i].Score, ranked[i+1].Score)
		}
	}
}

func TestRankArticlesEmpty(t *testing.T) {
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)

	ranked, err := ranker.RankArticles([]article.ArticleData{})
	if err != nil {
		t.Fatalf("RankArticles() error = %v", err)
	}

	if len(ranked) != 0 {
		t.Errorf("RankArticles() should return empty slice, got %d", len(ranked))
	}
}

func TestGetTopN(t *testing.T) {
	criteria := article.NewRankingCriteria()
	sourceMgr := NewSourceManager()
	sourceMgr.LoadDefaultSources()
	ranker := NewRankingEngine(criteria, sourceMgr)

	articles := []article.ArticleData{
		{
			ID:            "test-1",
			Title:         "Article 1",
			URL:           "https://example.com/1",
			SourceID:      "igamingbusiness",
			PublishedDate: time.Now().Format(time.RFC3339),
		},
		{
			ID:            "test-2",
			Title:         "Article 2",
			URL:           "https://example.com/2",
			SourceID:      "sporttech",
			PublishedDate: time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
		},
		{
			ID:            "test-3",
			Title:         "Article 3",
			URL:           "https://example.com/3",
			SourceID:      "eganingreview",
			PublishedDate: time.Now().AddDate(0, 0, -2).Format(time.RFC3339),
		},
		{
			ID:            "test-4",
			Title:         "Article 4",
			URL:           "https://example.com/4",
			SourceID:      "betindustry",
			PublishedDate: time.Now().AddDate(0, 0, -3).Format(time.RFC3339),
		},
	}

	ranked, err := ranker.GetTopN(articles, 2)
	if err != nil {
		t.Fatalf("GetTopN() error = %v", err)
	}

	if len(ranked) != 2 {
		t.Errorf("GetTopN(2) returned %d articles, expected 2", len(ranked))
	}

	if ranked[0].Rank != 1 || ranked[1].Rank != 2 {
		t.Errorf("Top 2 articles don't have ranks 1 and 2")
	}
}

func TestGetTopNMoreThanAvailable(t *testing.T) {
	criteria := article.NewRankingCriteria()
	ranker := NewRankingEngine(criteria, nil)

	articles := []article.ArticleData{
		{
			ID:            "test-1",
			Title:         "Article 1",
			URL:           "https://example.com/1",
			PublishedDate: time.Now().Format(time.RFC3339),
		},
	}

	ranked, err := ranker.GetTopN(articles, 10)
	if err != nil {
		t.Fatalf("GetTopN() error = %v", err)
	}

	if len(ranked) != 1 {
		t.Errorf("GetTopN(10) on 1 article returned %d, expected 1", len(ranked))
	}
}

func TestWeightSum(t *testing.T) {
	// Verify weights sum to 1.0
	criteria := article.NewRankingCriteria()

	weightSum := criteria.RecencyWeight +
		criteria.SourceWeight +
		criteria.EngagementWeight +
		criteria.CategoryWeight

	expectedSum := 1.0
	tolerance := 0.0001

	if weightSum < expectedSum-tolerance || weightSum > expectedSum+tolerance {
		t.Errorf("Weights don't sum to 1.0: %f", weightSum)
	}
}
