package article

import "time"

// Article represents a curated iGaming news article
type ArticleData struct {
	ID            string                 `json:"id"`            // Unique identifier (hash of URL or custom)
	Title         string                 `json:"title"`         // Article headline
	Summary       string                 `json:"summary"`       // AI-generated summary
	OriginalSum   string                 `json:"originalSummary"` // Article excerpt/description
	FullContent   string                 `json:"fullContent,omitempty"` // Scraped article body (optional)
	URL           string                 `json:"url"`           // Source article URL
	SourceName    string                 `json:"sourceName"`    // e.g., "iGamingBusiness", "Sportech"
	SourceID      string                 `json:"sourceId"`      // Reference to news source
	PublishedDate string                 `json:"publishedDate"` // RFC3339 format
	ImageURL      string                 `json:"imageUrl,omitempty"` // Featured image
	Categories    []string               `json:"categories,omitempty"` // Tags: "Regulations", "Sports Betting", etc
	Authors       []string               `json:"authors,omitempty"` // Article author(s)
	Metadata      map[string]interface{} `json:"metadata,omitempty"` // Extra fields (views, engagement, etc.)
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

// ArticleMetadata represents minimal article info for listings
type ArticleMetadata struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	SourceName    string   `json:"sourceName"`
	PublishedDate string   `json:"publishedDate"`
	URL           string   `json:"url"`
	Categories    []string `json:"categories,omitempty"`
	ImageURL      string   `json:"imageUrl,omitempty"`
}

// RankedArticle represents an article with its ranking score and reason
type RankedArticle struct {
	Article ArticleData `json:"article"`
	Score   float64     `json:"score"`      // Ranking score (0-1)
	Rank    int         `json:"rank"`       // Position in ranking (1-5)
	Reason  string      `json:"reason"`     // Why selected: "trending", "high-engagement", "sector-news"
}

// DailyDigest represents the daily top 5 curated articles
type DailyDigest struct {
	Date     string           `json:"date"`     // YYYY-MM-DD
	Articles []RankedArticle  `json:"articles"` // Top 5 articles
	Headline string           `json:"headline"` // One-sentence super summary
	Summary  string           `json:"summary"`  // Overall day summary
	Created  time.Time        `json:"created"`
}

// ArticleCategory represents article categorization
type ArticleCategory string

const (
	CategoryRegulations    ArticleCategory = "Regulations"
	CategoryBusiness       ArticleCategory = "Business"
	CategoryTechnology     ArticleCategory = "Technology"
	CategorySportsBetting  ArticleCategory = "Sports Betting"
	CategoryMergerAcquisition ArticleCategory = "M&A"
	CategoryInternational  ArticleCategory = "International"
	CategoryPayments       ArticleCategory = "Payments"
	CategoryResponsibleGaming ArticleCategory = "Responsible Gaming"
)

// ErrorCode defines error types for article operations
type ErrorCode string

const (
	ErrorCodeInvalidID     ErrorCode = "INVALID_ID"
	ErrorCodeNotFound      ErrorCode = "ARTICLE_NOT_FOUND"
	ErrorCodeFetchError    ErrorCode = "FETCH_ERROR"
	ErrorCodeRateLimited   ErrorCode = "RATE_LIMITED"
	ErrorCodeInternalError ErrorCode = "INTERNAL_ERROR"
	ErrorCodeValidationError ErrorCode = "VALIDATION_ERROR"
	ErrorCodeSummarizationFailed ErrorCode = "SUMMARIZATION_FAILED"
)

// ApiError defines the structure for an error response
type ApiError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ArticleFilter represents filtering options for article queries
type ArticleFilter struct {
	SourceNames  []string
	Categories   []string
	DateFrom     time.Time
	DateTo       time.Time
	Search       string
	Limit        int
	Offset       int
}

// RankingCriteria defines weights for article ranking
type RankingCriteria struct {
	RecencyWeight      float64 // Recent articles score higher (0-1)
	SourceWeight       float64 // Trusted sources score higher (0-1)
	CategoryWeight     float64 // Category diversity factor (0-1)
	EngagementWeight   float64 // Comments, shares (if available) (0-1)
}

// NewRankingCriteria creates default ranking criteria
func NewRankingCriteria() *RankingCriteria {
	return &RankingCriteria{
		RecencyWeight:    0.40,
		SourceWeight:     0.30,
		EngagementWeight: 0.20,
		CategoryWeight:   0.10,
	}
}
