package feed

import (
	"fmt"
	"main/lib/analytics"
	"main/lib/logger"
	"os"
	"time"
)

// GetFeedRawResult holds the result of fetching the feed.
type GetFeedRawResult struct {
	Data   *RssFeed
	Source string
}

// GetFeedRaw attempts to fetch the latest feed from blob cache, falling back to a fresh parse.
func GetFeedRaw() (*GetFeedRawResult, error) {
	// Feature flag to bypass blob cache entirely
	disableBlob := os.Getenv("DISABLE_BLOB_CACHE") == "1" || os.Getenv("DISABLE_BLOB_CACHE") == "true"

	// 1. First try to get cached feed from blob storage (unless disabled)
	var feed *RssFeed
	var err error
	if !disableBlob {
		feed, err = GetLatestTldrFeed()
		if err != nil {
			// Log the error but don't fail, as we can fall back to a fresh parse.
			logger.Error("Failed to get latest feed from blob cache", err, nil)
		}
	}

	if feed != nil {
		// After 07:05 UTC, ensure the cached feed's date is today; otherwise prefer a fresh parse
		if feed.LastBuildDate != "" {
			if t, err := time.Parse(time.RFC1123Z, feed.LastBuildDate); err == nil {
				now := time.Now().UTC()
				afterSevenOhFive := now.Hour() > 7 || (now.Hour() == 7 && now.Minute() >= 5)
				sameYMD := t.UTC().Year() == now.Year() && t.UTC().Month() == now.Month() && t.UTC().Day() == now.Day()
				if afterSevenOhFive && !sameYMD {
					logger.Warn("Cached TLDR feed appears stale after 07:05 UTC; fetching fresh", map[string]interface{}{
						"cachedDate": feed.LastBuildDate,
						"currentTime": now.Format(time.RFC3339),
					})
				} else {
					_ = analytics.Track("feed_served", "cache", map[string]interface{}{"source": "blob-cache"})
					return &GetFeedRawResult{
						Data:   feed,
						Source: "blob-cache",
					}, nil
				}
			} else {
				logger.Warn("Failed to parse LastBuildDate in cached feed", map[string]interface{}{
					"lastBuildDate": feed.LastBuildDate,
					"error": err.Error(),
				})
			}
		} else {
			return &GetFeedRawResult{
				Data:   feed,
				Source: "blob-cache",
			}, nil
		}
	}

	// 2. If no cached feed, fetch fresh data
	feed, err = ParseRssFeed()
	if err != nil {
		return nil, fmt.Errorf("failed to parse fresh RSS feed: %w", err)
	}

	// 3. Store the fresh feed in blob storage (unless disabled)
	if !disableBlob {
		err = StoreTldrFeed(feed)
		if err != nil {
			logger.Error("Failed to store fresh feed in blob cache", err, map[string]interface{}{
				"feedTitle": feed.Title,
			})
		}
	}

	if feed == nil {
		return nil, fmt.Errorf("feed not available from any source")
	}

	_ = analytics.Track("feed_served", "fresh", map[string]interface{}{"source": "rss-fresh"})

	return &GetFeedRawResult{
		Data:   feed,
		Source: "rss-fresh",
	}, nil
}
