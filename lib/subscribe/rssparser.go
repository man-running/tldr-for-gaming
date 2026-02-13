package subscribe

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/mmcdole/gofeed"
)

// This file is a copy of the RSS parser logic from other packages.

var (
	divRegex      = regexp.MustCompile(`(?s)<div>(.*?)</div>`)
	headlineRegex = regexp.MustCompile(`<h2>Morning Headline</h2>\s*<p>([^<]+)</p>`)
	h2SplitRegex  = regexp.MustCompile(`<h2>`)
)

// processRssItem replicates the logic of RssContentProcessor.processRssItem
func processRssItem(item *gofeed.Item) (string, []FeedItem) {
	content := item.Description
	divMatches := divRegex.FindStringSubmatch(content)
	if len(divMatches) < 2 {
		return "", []FeedItem{}
	}
	innerContent := divMatches[1]

	headlineMatches := headlineRegex.FindStringSubmatch(innerContent)
	headline := ""
	if len(headlineMatches) > 1 {
		headline = headlineMatches[1]
	}

	sectionParts := h2SplitRegex.Split(innerContent, -1)
	if len(sectionParts) <= 1 {
		return headline, []FeedItem{}
	}

	var feedItems []FeedItem
	for i, section := range sectionParts[1:] {
		feedItems = append(feedItems, FeedItem{
			Title:       fmt.Sprintf("Section %d", i+1),
			Link:        item.Link,
			Description: section,
			PubDate:     item.Published,
			GUID:        fmt.Sprintf("%s-section-%d", item.GUID, i),
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
			baseURL = "https://papers.takara.ai"
		}
	}
	feedURL := baseURL + "/api/tldr"

	fp := gofeed.NewParser()
	fp.Client = &http.Client{Timeout: 10 * time.Second}
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
