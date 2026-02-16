package feed

import (
	"fmt"
	"main/lib/article"
	"math"
	"sort"
	"strings"
	"time"
)

// RankingEngine calculates article scores based on RankingCriteria
type RankingEngine struct {
	criteria      *article.RankingCriteria
	sourceManager *SourceManager
}

// ScoreBreakdown provides detailed scoring information
type ScoreBreakdown struct {
	RecencyScore    float64 // 0-1
	SourceScore     float64 // 0-1
	EngagementScore float64 // 0-1
	CategoryScore   float64 // 0-1
	FinalScore      float64 // 0-1 (weighted sum)
	Reason          string  // Why ranked: "trending", "authoritative", etc
}

// NewRankingEngine creates a new ranking engine
func NewRankingEngine(criteria *article.RankingCriteria, sourceMgr *SourceManager) *RankingEngine {
	if criteria == nil {
		criteria = article.NewRankingCriteria()
	}
	return &RankingEngine{
		criteria:      criteria,
		sourceManager: sourceMgr,
	}
}

// CalculateScore calculates the score breakdown for a single article
func (re *RankingEngine) CalculateScore(art *article.ArticleData) (*ScoreBreakdown, error) {
	if art == nil {
		return nil, fmt.Errorf("article cannot be nil")
	}

	sb := &ScoreBreakdown{}

	// 1. Recency Score (decay by age)
	sb.RecencyScore = re.calculateRecencyScore(art.PublishedDate)

	// 2. Source Score (based on priority)
	sb.SourceScore = re.calculateSourceScore(art.SourceID)

	// 3. Engagement Score (from metadata)
	sb.EngagementScore = re.calculateEngagementScore(art)

	// 4. Category Score (diversity factor)
	sb.CategoryScore = re.calculateCategoryScore(art)

	// Calculate final weighted score
	sb.FinalScore = (sb.RecencyScore * re.criteria.RecencyWeight) +
		(sb.SourceScore * re.criteria.SourceWeight) +
		(sb.EngagementScore * re.criteria.EngagementWeight) +
		(sb.CategoryScore * re.criteria.CategoryWeight)

	// Ensure score is in [0, 1] range
	if sb.FinalScore < 0 {
		sb.FinalScore = 0
	} else if sb.FinalScore > 1 {
		sb.FinalScore = 1
	}

	// Assign reason(s)
	sb.Reason = re.assignReason(sb)

	return sb, nil
}

// calculateRecencyScore uses exponential decay based on hours since publication
// Score 1.0 for today, decays to ~0 by day 14
func (re *RankingEngine) calculateRecencyScore(publishedDate string) float64 {
	if publishedDate == "" {
		return 0.5 // Neutral score for missing date
	}

	pubTime, err := time.Parse(time.RFC3339, publishedDate)
	if err != nil {
		return 0.5 // Neutral score for unparseable date
	}

	hoursOld := time.Since(pubTime).Hours()
	if hoursOld < 0 {
		hoursOld = 0 // Handle future dates
	}

	// Exponential decay: exp(-hoursOld / 24)
	score := math.Exp(-hoursOld / 24)

	// Clamp to [0, 1]
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}

	return score
}

// calculateSourceScore looks up source priority and normalizes it
func (re *RankingEngine) calculateSourceScore(sourceID string) float64 {
	if sourceID == "" || re.sourceManager == nil {
		return 0.5 // Neutral score for unknown source
	}

	source, err := re.sourceManager.GetSource(sourceID)
	if err != nil || source == nil {
		return 0.5 // Neutral score if source not found
	}

	// Normalize priority (1-10) to (0.1-1.0)
	if source.Priority < 1 {
		source.Priority = 1
	} else if source.Priority > 10 {
		source.Priority = 10
	}

	return float64(source.Priority) / 10.0
}

// calculateEngagementScore extracts engagement metrics from metadata
func (re *RankingEngine) calculateEngagementScore(art *article.ArticleData) float64 {
	if art.Metadata == nil || len(art.Metadata) == 0 {
		return 0.5 // Neutral score for no metadata
	}

	// Check for various engagement metrics
	engagementKeys := []string{"views", "shares", "comments", "engagement_score", "engagement"}

	for _, key := range engagementKeys {
		if val, exists := art.Metadata[key]; exists {
			// Try to extract numeric value
			switch v := val.(type) {
			case float64:
				// Normalize to 0-1 range (assuming reasonable max values)
				// For simplicity, divide by expected max and clamp
				normalized := v / 1000.0 // Assume max 1000 views/shares
				if normalized > 1 {
					normalized = 1
				}
				return normalized
			case int:
				normalized := float64(v) / 1000.0
				if normalized > 1 {
					normalized = 1
				}
				return normalized
			}
		}
	}

	return 0.5 // Neutral score if no recognized metrics
}

// calculateCategoryScore implements category diversity bonus/penalty
func (re *RankingEngine) calculateCategoryScore(art *article.ArticleData) float64 {
	// Base score is neutral
	// This can be enhanced later to track category frequency across batch
	// For now, return neutral
	return 0.5
}

// assignReason generates human-readable reasons for the ranking
func (re *RankingEngine) assignReason(sb *ScoreBreakdown) string {
	var reasons []string

	if sb.RecencyScore > 0.8 {
		reasons = append(reasons, "trending")
	}
	if sb.SourceScore > 0.8 {
		reasons = append(reasons, "authoritative")
	}
	if sb.EngagementScore > 0.8 {
		reasons = append(reasons, "high-engagement")
	}
	if sb.CategoryScore > 0.5 {
		reasons = append(reasons, "diverse")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "featured")
	}

	return strings.Join(reasons, ", ")
}

// RankArticles scores and ranks all articles by final score
func (re *RankingEngine) RankArticles(articles []article.ArticleData) ([]article.RankedArticle, error) {
	if len(articles) == 0 {
		return []article.RankedArticle{}, nil
	}

	// Score all articles
	rankedArticles := make([]article.RankedArticle, len(articles))
	for i, art := range articles {
		scoreBreakdown, err := re.CalculateScore(&art)
		if err != nil {
			// Log error but continue with default score
			fmt.Printf("Failed to score article %s: %v\n", art.ID, err)
			scoreBreakdown = &ScoreBreakdown{
				RecencyScore:    0.5,
				SourceScore:     0.5,
				EngagementScore: 0.5,
				CategoryScore:   0.5,
				FinalScore:      0.5,
				Reason:          "unscored",
			}
		}

		rankedArticles[i] = article.RankedArticle{
			Article: art,
			Score:   scoreBreakdown.FinalScore,
			Rank:    0, // Will be set after sorting
			Reason:  scoreBreakdown.Reason,
		}
	}

	// Sort by score descending
	sort.Slice(rankedArticles, func(i, j int) bool {
		return rankedArticles[i].Score > rankedArticles[j].Score
	})

	// Assign ranks
	for i := range rankedArticles {
		rankedArticles[i].Rank = i + 1
	}

	return rankedArticles, nil
}

// GetTopN returns the top N ranked articles
func (re *RankingEngine) GetTopN(articles []article.ArticleData, n int) ([]article.RankedArticle, error) {
	ranked, err := re.RankArticles(articles)
	if err != nil {
		return nil, err
	}

	if n <= 0 || n >= len(ranked) {
		return ranked, nil
	}

	return ranked[:n], nil
}
