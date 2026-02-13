package article

import (
	"testing"
	"time"
)

// TestArticleDataCreation tests creating valid article data
func TestArticleDataCreation(t *testing.T) {
	article := &ArticleData{
		ID:            "article-001",
		Title:         "New Gaming Regulation in UK",
		Summary:       "The UK gaming authority announced new regulations...",
		OriginalSum:   "Breaking news about gaming regulations",
		URL:           "https://example.com/article-001",
		SourceName:    "iGamingBusiness",
		SourceID:      "igamingbusiness",
		PublishedDate: time.Now().Format(time.RFC3339),
		Categories:    []string{string(CategoryRegulations)},
		Authors:       []string{"John Doe"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if article.ID != "article-001" {
		t.Errorf("Expected ID 'article-001', got '%s'", article.ID)
	}
	if article.Title == "" {
		t.Error("Title should not be empty")
	}
	if article.SourceName != "iGamingBusiness" {
		t.Errorf("Expected SourceName 'iGamingBusiness', got '%s'", article.SourceName)
	}
}

// TestArticleMetadataCreation tests creating article metadata
func TestArticleMetadataCreation(t *testing.T) {
	metadata := &ArticleMetadata{
		ID:            "article-001",
		Title:         "Test Article",
		SourceName:    "iGamingBusiness",
		PublishedDate: time.Now().Format(time.RFC3339),
		URL:           "https://example.com/article-001",
		Categories:    []string{string(CategoryBusiness)},
	}

	if metadata.ID == "" {
		t.Error("Metadata ID should not be empty")
	}
	if metadata.Title == "" {
		t.Error("Metadata Title should not be empty")
	}
	if len(metadata.Categories) == 0 {
		t.Error("Metadata should have categories")
	}
}

// TestRankedArticle tests ranking article with score and position
func TestRankedArticle(t *testing.T) {
	article := &ArticleData{
		ID:    "article-001",
		Title: "Test Article",
	}

	ranked := &RankedArticle{
		Article: *article,
		Score:   0.95,
		Rank:    1,
		Reason:  "trending",
	}

	if ranked.Score < 0 || ranked.Score > 1 {
		t.Errorf("Score should be between 0 and 1, got %f", ranked.Score)
	}
	if ranked.Rank != 1 {
		t.Errorf("Expected rank 1, got %d", ranked.Rank)
	}
	if ranked.Reason != "trending" {
		t.Errorf("Expected reason 'trending', got '%s'", ranked.Reason)
	}
}

// TestDailyDigest tests creating a daily digest
func TestDailyDigest(t *testing.T) {
	articles := make([]RankedArticle, 5)
	for i := 0; i < 5; i++ {
		articles[i] = RankedArticle{
			Article: ArticleData{
				ID:    "article-" + string(rune(i)),
				Title: "Article " + string(rune(i)),
			},
			Score: float64(1 - (i * 0.1)),
			Rank:  i + 1,
		}
	}

	digest := &DailyDigest{
		Date:     "2026-02-13",
		Articles: articles,
		Summary:  "Daily summary of top iGaming articles",
		Created:  time.Now(),
	}

	if digest.Date == "" {
		t.Error("Date should not be empty")
	}
	if len(digest.Articles) != 5 {
		t.Errorf("Expected 5 articles, got %d", len(digest.Articles))
	}
	if digest.Articles[0].Rank != 1 {
		t.Errorf("First article should be rank 1, got %d", digest.Articles[0].Rank)
	}
}

// TestArticleCategories tests category enum
func TestArticleCategories(t *testing.T) {
	categories := []ArticleCategory{
		CategoryRegulations,
		CategoryBusiness,
		CategoryTechnology,
		CategorySportsBetting,
		CategoryMergerAcquisition,
		CategoryInternational,
		CategoryPayments,
		CategoryResponsibleGaming,
	}

	if len(categories) != 8 {
		t.Errorf("Expected 8 categories, got %d", len(categories))
	}

	for _, cat := range categories {
		if cat == "" {
			t.Error("Category should not be empty")
		}
	}
}

// TestErrorCodes tests error code enum
func TestErrorCodes(t *testing.T) {
	errors := []ErrorCode{
		ErrorCodeInvalidID,
		ErrorCodeNotFound,
		ErrorCodeFetchError,
		ErrorCodeRateLimited,
		ErrorCodeInternalError,
		ErrorCodeValidationError,
		ErrorCodeSummarizationFailed,
	}

	if len(errors) != 7 {
		t.Errorf("Expected 7 error codes, got %d", len(errors))
	}

	for _, errCode := range errors {
		if errCode == "" {
			t.Error("Error code should not be empty")
		}
	}
}

// TestApiError tests error response structure
func TestApiError(t *testing.T) {
	apiErr := &ApiError{
		Code:    ErrorCodeNotFound,
		Message: "Article not found",
		Details: map[string]interface{}{
			"articleId": "article-999",
		},
	}

	if apiErr.Code != ErrorCodeNotFound {
		t.Errorf("Expected error code %v, got %v", ErrorCodeNotFound, apiErr.Code)
	}
	if apiErr.Message == "" {
		t.Error("Error message should not be empty")
	}
	if len(apiErr.Details) != 1 {
		t.Errorf("Expected 1 detail, got %d", len(apiErr.Details))
	}
}

// TestArticleFilter tests filtering options
func TestArticleFilter(t *testing.T) {
	filter := &ArticleFilter{
		SourceNames: []string{"iGamingBusiness", "Gambling Insider"},
		Categories:  []string{string(CategoryBusiness)},
		DateFrom:    time.Now().AddDate(0, 0, -7),
		DateTo:      time.Now(),
		Search:      "regulations",
		Limit:       10,
		Offset:      0,
	}

	if len(filter.SourceNames) != 2 {
		t.Errorf("Expected 2 sources in filter, got %d", len(filter.SourceNames))
	}
	if filter.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", filter.Limit)
	}
	if filter.Offset != 0 {
		t.Errorf("Expected offset 0, got %d", filter.Offset)
	}
}

// TestRankingCriteria tests ranking weight configuration
func TestRankingCriteria(t *testing.T) {
	criteria := NewRankingCriteria()

	if criteria == nil {
		t.Fatal("NewRankingCriteria should not return nil")
	}

	// Check default weights
	if criteria.RecencyWeight != 0.40 {
		t.Errorf("Expected recency weight 0.40, got %f", criteria.RecencyWeight)
	}
	if criteria.SourceWeight != 0.30 {
		t.Errorf("Expected source weight 0.30, got %f", criteria.SourceWeight)
	}
	if criteria.EngagementWeight != 0.20 {
		t.Errorf("Expected engagement weight 0.20, got %f", criteria.EngagementWeight)
	}
	if criteria.CategoryWeight != 0.10 {
		t.Errorf("Expected category weight 0.10, got %f", criteria.CategoryWeight)
	}

	// Verify weights sum to 1.0
	totalWeight := criteria.RecencyWeight + criteria.SourceWeight +
		criteria.EngagementWeight + criteria.CategoryWeight
	expectedTotal := 1.0
	if totalWeight != expectedTotal {
		t.Errorf("Weights should sum to 1.0, got %f", totalWeight)
	}
}

// TestRankingCriteriaCustomization tests modifying ranking criteria
func TestRankingCriteriaCustomization(t *testing.T) {
	criteria := NewRankingCriteria()

	// Modify criteria
	criteria.RecencyWeight = 0.50
	criteria.SourceWeight = 0.25
	criteria.EngagementWeight = 0.15
	criteria.CategoryWeight = 0.10

	if criteria.RecencyWeight != 0.50 {
		t.Errorf("Expected modified recency weight 0.50, got %f", criteria.RecencyWeight)
	}

	totalWeight := criteria.RecencyWeight + criteria.SourceWeight +
		criteria.EngagementWeight + criteria.CategoryWeight
	expectedTotal := 1.0
	if totalWeight != expectedTotal {
		t.Errorf("Modified weights should sum to 1.0, got %f", totalWeight)
	}
}

// TestArticleDataWithMetadata tests article with complex metadata
func TestArticleDataWithMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"views":     1500,
		"shares":    45,
		"comments":  12,
		"engagement_score": 0.78,
		"sentiment": "positive",
	}

	article := &ArticleData{
		ID:       "article-001",
		Title:    "Test Article",
		Metadata: metadata,
	}

	if article.Metadata == nil {
		t.Error("Metadata should not be nil")
	}
	if len(article.Metadata) != 5 {
		t.Errorf("Expected 5 metadata fields, got %d", len(article.Metadata))
	}

	views, ok := article.Metadata["views"].(float64)
	if !ok {
		t.Error("Failed to retrieve views from metadata")
	}
	if views != 1500 {
		t.Errorf("Expected views 1500, got %v", views)
	}
}

// TestArticleDataWithMultipleAuthors tests article with multiple authors
func TestArticleDataWithMultipleAuthors(t *testing.T) {
	article := &ArticleData{
		ID:      "article-001",
		Title:   "Collaborative Article",
		Authors: []string{"John Doe", "Jane Smith", "Bob Johnson"},
	}

	if len(article.Authors) != 3 {
		t.Errorf("Expected 3 authors, got %d", len(article.Authors))
	}

	expectedAuthors := []string{"John Doe", "Jane Smith", "Bob Johnson"}
	for i, author := range article.Authors {
		if author != expectedAuthors[i] {
			t.Errorf("Expected author '%s', got '%s'", expectedAuthors[i], author)
		}
	}
}

// TestArticleDataWithMultipleCategories tests article with multiple categories
func TestArticleDataWithMultipleCategories(t *testing.T) {
	article := &ArticleData{
		ID:         "article-001",
		Title:      "Multi-category Article",
		Categories: []string{string(CategoryBusiness), string(CategoryRegulations)},
	}

	if len(article.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(article.Categories))
	}

	if article.Categories[0] != string(CategoryBusiness) {
		t.Errorf("Expected first category Business, got %s", article.Categories[0])
	}
	if article.Categories[1] != string(CategoryRegulations) {
		t.Errorf("Expected second category Regulations, got %s", article.Categories[1])
	}
}

// TestRankedArticleScoreValidation tests rank score bounds
func TestRankedArticleScoreValidation(t *testing.T) {
	testCases := []struct {
		name        string
		score       float64
		shouldBeOk  bool
	}{
		{"Perfect score", 1.0, true},
		{"Zero score", 0.0, true},
		{"Mid range", 0.5, true},
		{"High score", 0.95, true},
		{"Low score", 0.05, true},
		{"Negative score", -0.1, false},
		{"Over one", 1.1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ranked := &RankedArticle{
				Score: tc.score,
			}

			isValid := ranked.Score >= 0 && ranked.Score <= 1
			if isValid != tc.shouldBeOk {
				t.Errorf("Score %f validity mismatch: expected %v, got %v",
					tc.score, tc.shouldBeOk, isValid)
			}
		})
	}
}

// TestDailyDigestDateFormat tests date format handling
func TestDailyDigestDateFormat(t *testing.T) {
	validDates := []string{
		"2026-02-13",
		"2025-12-31",
		"2026-01-01",
	}

	for _, dateStr := range validDates {
		digest := &DailyDigest{
			Date: dateStr,
		}

		if digest.Date == "" {
			t.Errorf("Date should not be empty for '%s'", dateStr)
		}
	}
}
