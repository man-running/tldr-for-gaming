package handler

import (
	"fmt"
	"io"
	"main/lib/logger"
	"main/lib/middleware"
	"main/lib/summary"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

// constructAbsoluteURL constructs an absolute URL using BASE_URL and Vercel fallbacks
func constructAbsoluteURL(path string) string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		// Try Vercel environment variables in order of preference
		if prodURL := os.Getenv("VERCEL_PROJECT_PRODUCTION_URL"); prodURL != "" {
			baseURL = "https://" + prodURL
		} else if deployURL := os.Getenv("VERCEL_URL"); deployURL != "" {
			baseURL = "https://" + deployURL
		} else if branchURL := os.Getenv("VERCEL_BRANCH_URL"); branchURL != "" {
			baseURL = "https://" + branchURL
		} else {
			baseURL = "https://tldr.takara.ai"
		}
	}
	// Remove leading slash from path if present
	path = strings.TrimPrefix(path, "/")
	return baseURL + "/" + path
}

// feedHandler contains the main logic for the feed endpoint
// This is now primarily a fallback endpoint - clients should fetch directly from blob storage
// Pattern: https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com/tldr-feeds/{latestDate}.json
// Only used when DISABLE_BLOB_CACHE is true or when blob fetch fails
func feedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	// Check if blob cache is disabled - if so, generate fresh data
	disableBlob := os.Getenv("DISABLE_BLOB_CACHE") == "1" || os.Getenv("DISABLE_BLOB_CACHE") == "true"
	if disableBlob {
		logger.Info("Blob cache disabled, generating fresh feed", ctx)
	} else {
		logger.Info("Feed API called (fallback - clients should use blob storage directly)", ctx)
	}

	// 1. Get the TLDR summary data directly using the summary service
	logger.LogRequestStart(r)
	service := summary.NewService()

	// Construct the request URL for the TLDR endpoint
	requestURL := constructAbsoluteURL("api/tldr")
	ctx["request_url"] = requestURL

	result, err := service.GetSummaryRaw(r.Context(), requestURL)
	if err != nil {
		logger.Error("Failed to get TLDR summary data", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// If blob URL is available but data is nil, fetch from blob URL
	var rssData []byte
	if result.BlobURL != nil && *result.BlobURL != "" && result.Data == nil {
		resp, err := http.Get(*result.BlobURL)
		if err != nil {
			logger.Error("Failed to fetch from blob URL", err, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.Error("Blob URL returned non-200", nil, map[string]interface{}{
				"status": resp.StatusCode,
				"url":    *result.BlobURL,
			})
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		var readErr error
		rssData, readErr = io.ReadAll(resp.Body)
		if readErr != nil {
			logger.Error("Failed to read blob content", readErr, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	} else {
		rssData = result.Data
	}

	// 2. Parse the RSS data to convert to JSON format expected by feed endpoint
	feedData, err := parseRSSToFeedData(rssData)
	if err != nil {
		logger.Error("Failed to parse RSS to feed data", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// 3. Write JSON response with proper content type
	middleware.WriteJSONResponse(w, http.StatusOK, feedData)

	logger.LogRequestComplete(r, http.StatusOK, 0) // Duration will be tracked by middleware
}

// FeedItem matches the structure expected by the frontend
type FeedItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	PubDate     string `json:"pubDate"`
	GUID        string `json:"guid"`
}

// RssFeed matches the structure expected by the frontend
type RssFeed struct {
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Link          string     `json:"link"`
	LastBuildDate string     `json:"lastBuildDate,omitempty"`
	Items         []FeedItem `json:"items"`
}

var (
	// Capture the inner content of the outermost div (non-greedy, dotall, allow attributes)
	divRegex = regexp.MustCompile(`(?is)<div[^>]*>(.*?)</div>`)
	// Capture the entire first paragraph following the Morning Headline h2 (including links/inline HTML)
	headlineRegex = regexp.MustCompile(`(?is)<h2[^>]*>\s*Morning\s+Headline\s*</h2>\s*(<p[\s\S]*?>[\s\S]*?</p>)`)
	// Find all h2 tags to split content into sections
	h2TagRegex = regexp.MustCompile(`(?is)<h2[^>]*>.*?</h2>`)
)

// processRssItem extracts the headline and sections from RSS content
func processRssItem(item *gofeed.Item) (string, []FeedItem) {
	// Prefer full content when available
	content := item.Content
	if strings.TrimSpace(content) == "" {
		content = item.Description
	}

	// Extract content from wrapper div if present
	divMatches := divRegex.FindStringSubmatch(content)
	if len(divMatches) >= 2 {
		content = divMatches[1]
	}

	// Extract headline from Morning Headline paragraph
	headline := ""
	if m := headlineRegex.FindStringSubmatch(content); len(m) > 1 {
		// Extract inner content from <p> tag
		p := strings.TrimSpace(m[1])
		if idx := strings.Index(p, ">"); idx != -1 {
			inner := p[idx+1:]
			if end := strings.LastIndex(inner, "</p>"); end != -1 {
				headline = strings.TrimSpace(inner[:end])
			}
		}
	}

	// Split content by h2 sections
	var feedItems []FeedItem
	h2Positions := h2TagRegex.FindAllStringIndex(content, -1)

	for i, pos := range h2Positions {
		// Extract section title
		h2Match := regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`).FindStringSubmatch(content[pos[0]:pos[1]])
		if len(h2Match) < 2 {
			continue
		}
		title := strings.TrimSpace(h2Match[1])

		// Skip Morning Headline section
		if strings.EqualFold(title, "Morning Headline") {
			continue
		}

		// Extract section content (from current h2 to next h2 or end)
		var endPos int
		if i+1 < len(h2Positions) {
			endPos = h2Positions[i+1][0]
		} else {
			endPos = len(content)
		}

		sectionContent := strings.TrimSpace(content[pos[0]:endPos])

		feedItems = append(feedItems, FeedItem{
			Title:       title,
			Link:        item.Link,
			Description: sectionContent,
			PubDate:     item.Published,
			GUID:        fmt.Sprintf("%s-section-%d", item.GUID, i),
		})
	}

	return headline, feedItems
}

// parseRSSToFeedData converts RSS XML bytes to RssFeed JSON structure with proper section parsing
func parseRSSToFeedData(rssData []byte) (*RssFeed, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(string(rssData))
	if err != nil {
		return nil, err
	}

	if len(feed.Items) == 0 {
		return &RssFeed{
			Title:         "Takara TLDR",
			Description:   "Daily AI research summaries",
			Link:          "https://tldr.takara.ai",
			LastBuildDate: "",
			Items:         []FeedItem{},
		}, nil
	}

	// Process the first item to extract headline and sections
	firstItem := feed.Items[0]
	headline, feedItems := processRssItem(firstItem)

	lastBuildDate := feed.Published
	if firstItem.Published != "" {
		lastBuildDate = firstItem.Published
	}

	result := &RssFeed{
		Title:         feed.Title,
		Description:   headline,
		Link:          feed.Link,
		LastBuildDate: lastBuildDate,
		Items:         feedItems,
	}

	return result, nil
}

// Handler is the Vercel serverless function entrypoint for the feed API.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Configure caching for feed endpoint
	cacheOpts := middleware.CacheOptions{
		Config: middleware.CacheConfig{
			MaxAge:               0,   // No browser caching
			SMaxAge:              300, // 5 minutes CDN cache
			StaleWhileRevalidate: 0,   // 1 hour stale-while-revalidate
			StaleIfError:         0,   // No stale-if-error
		},
		ETagKey: "tldr-feed",
		Enabled: true,
	}
	middleware.MethodAndCache(http.MethodGet, cacheOpts)(feedHandler)(w, r)
}
