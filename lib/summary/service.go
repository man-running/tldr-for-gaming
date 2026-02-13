package summary

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"main/lib/analytics"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
)

// Service provides the main functionality for summary operations
type Service struct {
	logger *slog.Logger
}

// NewService creates a new summary service
func NewService() *Service {
	return &Service{
		logger: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
	}
}

// GetSummaryRaw fetches the summary, trying cache first, then generating if needed
func (s *Service) GetSummaryRaw(ctx context.Context, requestURL string) (*GetSummaryRawResult, error) {
	s.logger.Info("GetSummaryRaw called", "requestURL", requestURL)

	// Try to get cached summary URL first (without fetching content)
	summaryBlobURL, err := GetLatestSummaryURL()
	if err != nil {
		s.logger.Warn("Failed to get cached summary URL", "error", err)
	}
	if summaryBlobURL != "" {
		s.logger.Info("Returning cached summary URL", "url", summaryBlobURL)
		_ = analytics.Track("summary_served", "cache", map[string]interface{}{"source": "blob-cache"})
		return &GetSummaryRawResult{
			Data:    nil, // Client will fetch from blob URL
			Source:  "blob-cache",
			BlobURL: &summaryBlobURL,
		}, nil
	}

	s.logger.Info("No cached summary found, generating new summary")

	// Get the papers RSS data (reusing existing scraped data)
	// Extract base URL and construct papers endpoint URL
	baseURL := strings.Split(requestURL, "?")[0] // Remove query parameters
	papersURL := strings.Replace(baseURL, "/api/tldr", "/api/papers", 1)
	papersResult, err := s.GetPapersRaw(ctx, papersURL)
	if err != nil {
		s.logger.Warn("Failed to get papers data, falling back to direct generation", "error", err)
		// Fallback to direct generation if papers endpoint fails
		summaryData, err := s.generateSummaryDirect(ctx, requestURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate summary: %w", err)
		}
		return &GetSummaryRawResult{
			Data:   summaryData,
			Source: "generated-fallback",
		}, nil
	}

	// Generate summary from existing RSS data
	s.logger.Info("Generating summary from RSS data", "rss_size", len(papersResult.Data))
	summaryData, err := s.GenerateSummaryFromRSS(ctx, papersResult.Data, requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary from RSS: %w", err)
	}
	s.logger.Info("Summary generation completed", "summary_size", len(summaryData))
	_ = analytics.Track("summary_generated", "generated", map[string]interface{}{"source": "rss"})

	// Cache the new summary asynchronously
	go func() {
		if err := StoreSummary(summaryData); err != nil {
			s.logger.Error("Failed to cache summary", "error", err)
		}
	}()

	return &GetSummaryRawResult{
		Data:   summaryData,
		Source: "generated",
	}, nil
}

// generateSummaryDirect generates a summary from scratch (legacy method)
func (s *Service) generateSummaryDirect(ctx context.Context, requestURL string) ([]byte, error) {
	// Get fresh feed data
	papers, err := scrapePapers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape papers: %w", err)
	}

	// Generate RSS from papers
	feedData, err := GeneratePapersRSS(papers, requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSS feed: %w", err)
	}

	// Generate summary from RSS data
	return s.GenerateSummaryFromRSS(ctx, feedData, requestURL)
}

// GenerateSummaryFromRSS generates a summary from existing RSS data (public method)
func (s *Service) GenerateSummaryFromRSS(ctx context.Context, rssData []byte, requestURL string) ([]byte, error) {
	s.logger.Info("Starting summary generation from RSS", "rss_size", len(rssData))

	// Parse RSS to markdown
	originalMarkdown, err := s.parseRSSToMarkdown(string(rssData))
	if err != nil {
		s.logger.Error("Failed to parse RSS to markdown", "error", err)
		return nil, fmt.Errorf("failed to parse RSS to markdown: %w", err)
	}
	s.logger.Info("Parsed RSS to markdown", "markdown_length", len(originalMarkdown))

	// Extract links from the markdown
	feedURLs := s.extractLinksFromMarkdown(originalMarkdown)
	s.logger.Info("Extracted links from markdown", "link_count", len(feedURLs))

	// Generate summary with LLM
	s.logger.Info("Calling LLM for summary generation")
	summaryMarkdown, err := summarizeWithLLM(ctx, originalMarkdown, feedURLs)
	if err != nil {
		s.logger.Error("LLM summary generation failed after retries", "error", err)
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}
	s.logger.Info("LLM summary generation successful", "summary_length", len(summaryMarkdown))

	// Convert summary markdown to HTML
	htmlBytes := markdown.ToHTML([]byte(summaryMarkdown), nil, nil)
	htmlSummary := string(htmlBytes)
	s.logger.Info("Converted summary to HTML", "html_length", len(htmlSummary))

	// Generate summary RSS
	now := time.Now().UTC()
	summaryRSSBytes, err := GenerateSummaryRSS(htmlSummary, requestURL, now)
	if err != nil {
		s.logger.Error("Failed to generate summary RSS", "error", err)
		return nil, fmt.Errorf("failed to generate summary RSS: %w", err)
	}
	s.logger.Info("Generated summary RSS", "rss_size", len(summaryRSSBytes))

	return summaryRSSBytes, nil
}

// parseRSSToMarkdown converts RSS XML to markdown format
func (s *Service) parseRSSToMarkdown(xmlContent string) (string, error) {
	var rss RSS
	err := xml.Unmarshal([]byte(xmlContent), &rss)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal RSS XML: %w", err)
	}

	// Format date
	var formattedDate string
	parsedDate, err := time.Parse(time.RFC1123Z, rss.Channel.LastBuildDate)
	if err != nil {
		formattedDate = rss.Channel.LastBuildDate
	} else {
		formattedDate = parsedDate.Format("2006-01-02")
	}

	// Create markdown
	var markdown strings.Builder
	markdown.WriteString(fmt.Sprintf("# %s\n\n", rss.Channel.Title))
	markdown.WriteString(fmt.Sprintf("*%s*\n\n", rss.Channel.Description))
	markdown.WriteString(fmt.Sprintf("*Last updated: %s*\n\n", formattedDate))
	markdown.WriteString("---\n\n")

	// Process each item
	for _, item := range rss.Channel.Items {
		title := strings.ReplaceAll(item.Title, "\n", " ")
		title = strings.TrimSpace(title)

		markdown.WriteString(fmt.Sprintf("## [%s](%s)\n\n", title, item.Link))
		markdown.WriteString(fmt.Sprintf("%s\n\n", item.Description.Text))
		markdown.WriteString("---\n\n")
	}

	return markdown.String(), nil
}

// extractLinksFromMarkdown parses the input markdown to find ## [Title](URL) lines
func (s *Service) extractLinksFromMarkdown(markdownContent string) map[string]string {
	links := make(map[string]string)
	re := regexp.MustCompile(`(?m)^##\s*\[([^\]]+)\]\(([^)]+)\)$`)
	matches := re.FindAllStringSubmatch(markdownContent, -1)
	for _, match := range matches {
		if len(match) == 3 {
			title := strings.TrimSpace(match[1])
			url := strings.TrimSpace(match[2])
			links[title] = toTLDRLink(url)
		}
	}
	return links
}

// GetPapersRaw fetches papers, trying cache first, then scraping if needed
func (s *Service) GetPapersRaw(ctx context.Context, requestURL string) (*GetSummaryRawResult, error) {
	s.logger.Info("GetPapersRaw called", "requestURL", requestURL)

	// Try to get cached papers URL first (without fetching content)
	papersBlobURL, err := GetLatestPapersURL()
	if err != nil {
		s.logger.Warn("Failed to get cached papers URL", "error", err.Error())
		// Don't return error, just continue to generate fresh
	} else if papersBlobURL != "" {
		s.logger.Info("Returning cached papers URL", "url", papersBlobURL)
		_ = analytics.Track("papers_served", "cache", map[string]interface{}{"source": "blob-cache"})
		return &GetSummaryRawResult{
			Data:    nil, // Client will fetch from blob URL
			Source:  "blob-cache",
			BlobURL: &papersBlobURL,
		}, nil
	}

	s.logger.Info("No cached papers found, generating fresh papers")

	// Generate fresh papers data
	s.logger.Info("Starting scrapePapers call")
	papers, err := scrapePapers(ctx)
	if err != nil {
		s.logger.Error("scrapePapers failed", "error", err)
		return nil, fmt.Errorf("failed to scrape papers: %w", err)
	}
	s.logger.Info("scrapePapers completed", "paper_count", len(papers))

	// Always use the papers endpoint URL for the self-link in papers RSS
	baseURL := strings.Split(requestURL, "?")[0] // Remove query parameters
	var papersURL string
	if strings.Contains(baseURL, "/api/tldr") {
		papersURL = strings.Replace(baseURL, "/api/tldr", "/api/papers", 1)
	} else {
		papersURL = baseURL
	}

	s.logger.Info("Starting RSS generation", "papersURL", papersURL)
	feedData, err := GeneratePapersRSS(papers, papersURL)
	if err != nil {
		s.logger.Error("GeneratePapersRSS failed", "error", err)
		return nil, fmt.Errorf("failed to generate RSS feed: %w", err)
	}
	s.logger.Info("RSS generation completed", "feed_size", len(feedData))

	// Cache the papers data
	go func() {
		if err := StorePapers(feedData); err != nil {
			s.logger.Error("Failed to cache papers data", "error", err)
		} else {
			s.logger.Info("Successfully cached papers data")
		}
	}()

	s.logger.Info("GetPapersRaw completed successfully", "source", "scraped")
	_ = analytics.Track("papers_generated", "scraped", map[string]interface{}{"count": len(papers)})
	return &GetSummaryRawResult{
		Data:   feedData,
		Source: "scraped",
	}, nil
}

// forceGeneratePapersRaw generates fresh papers data bypassing cache (for cache updates)
func (s *Service) forceGeneratePapersRaw(ctx context.Context, requestURL string) (*GetSummaryRawResult, error) {
	s.logger.Info("Force generating fresh papers data (bypassing cache)")

	// Always generate fresh papers data (bypass cache)
	papers, err := scrapePapers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape papers: %w", err)
	}

	// Always use the papers endpoint URL for the self-link in papers RSS
	baseURL := strings.Split(requestURL, "?")[0] // Remove query parameters
	var papersURL string
	if strings.Contains(baseURL, "/api/tldr") {
		papersURL = strings.Replace(baseURL, "/api/tldr", "/api/papers", 1)
	} else {
		papersURL = baseURL
	}

	s.logger.Info("Force generating RSS", "papersURL", papersURL)
	feedData, err := GeneratePapersRSS(papers, papersURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSS feed: %w", err)
	}

	return &GetSummaryRawResult{
		Data:   feedData,
		Source: "fresh-scraped",
	}, nil
}

// UpdateCache forces a refresh of both papers and summary caches (like rss_old.go updateAllCaches)
func (s *Service) UpdateCache(ctx context.Context, requestURL string) error {
	s.logger.Info("Starting comprehensive cache update for papers and summary")

	// 1. Generate fresh papers data (force bypass cache)
	s.logger.Info("Generating fresh papers data")
	papersData, err := s.forceGeneratePapersRaw(ctx, requestURL)
	if err != nil {
		return fmt.Errorf("failed to generate fresh papers data: %w", err)
	}

	// 2. Store papers in cache (this will be skipped if DISABLE_BLOB_CACHE is true)
	if err := StorePapers(papersData.Data); err != nil {
		s.logger.Warn("Failed to store papers in cache", "error", err)
		// Continue with summary generation even if papers cache fails
	} else {
		s.logger.Info("Successfully updated papers cache")
	}

	// 3. Generate summary from the fresh papers data
	s.logger.Info("Generating summary from fresh papers data")
	summaryData, err := s.GenerateSummaryFromRSS(ctx, papersData.Data, requestURL)
	if err != nil {
		return fmt.Errorf("failed to generate summary from fresh papers: %w", err)
	}

	// 4. Store summary in cache (this will be skipped if DISABLE_BLOB_CACHE is true)
	if err := StoreSummary(summaryData); err != nil {
		s.logger.Warn("Failed to store summary in cache", "error", err)
		// Continue - at least we generated fresh data
	} else {
		s.logger.Info("Successfully updated summary cache")
	}

	s.logger.Info("Comprehensive cache update completed successfully")
	_ = analytics.Track("cache_updated", "update", map[string]interface{}{"type": "comprehensive"})
	return nil
}
