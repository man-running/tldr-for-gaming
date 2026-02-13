package feed

import (
	"fmt"
	"main/lib/logger"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

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
			GUID:        GUIDString(fmt.Sprintf("%s-section-%d", item.GUID, i)),
		})
	}

	return headline, feedItems
}

// ParseRssFeed fetches and parses the RSS feed, returning a structured RssFeed object.
func ParseRssFeed() (*RssFeed, error) {
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
			logger.Error("BASE_URL environment variable not set and no Vercel environment variables available", nil, nil)
			return nil, fmt.Errorf("BASE_URL environment variable not set and no Vercel environment variables available")
		}
	}
	feedURL := baseURL + "/api/tldr"

	fp := gofeed.NewParser()
	fp.Client = &http.Client{Timeout: 30 * time.Second}

	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	if len(feed.Items) == 0 {
		return &RssFeed{
			Title:         "Takara TLDR",
			Description:   "Daily AI research summaries",
			Link:          "https://tldr.takara.ai",
			LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			Items:         []FeedItem{},
		}, nil
	}

	firstItem := feed.Items[0]
	headline, feedItems := processRssItem(firstItem)

	lastBuildDate := feed.Published
	if firstItem.Published != "" {
		lastBuildDate = firstItem.Published
	}

	finalFeed := &RssFeed{
		Title:         feed.Title,
		Description:   headline,
		Link:          feed.Link,
		LastBuildDate: lastBuildDate,
		Items:         feedItems,
	}
	return finalFeed, nil
}
