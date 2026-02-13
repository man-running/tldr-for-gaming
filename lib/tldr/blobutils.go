package tldr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	vercelBlobAPIURL  = "https://blob.vercel-storage.com"
	vercelBlobBaseURL = "https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com"
	tldrFeedsPrefix   = "tldr-feeds/"
)

// VercelListBlob represents a single blob item in the Vercel Blob List API response.
type VercelListBlob struct {
	URL      string `json:"url"`
	Pathname string `json:"pathname"`
}

// VercelListResponse is the structure of the response from the Vercel Blob List API.
type VercelListResponse struct {
	Blobs []VercelListBlob `json:"blobs"`
}

// listBlobsManually performs a GET request to the Vercel Blob List API.
func listBlobsManually(prefix string) (*VercelListResponse, error) {
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("BLOB_READ_WRITE_TOKEN environment variable not set")
	}

	req, err := http.NewRequest("GET", vercelBlobAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	q := req.URL.Query()
	q.Add("prefix", prefix)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute list request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("blob storage list API returned non-200 status: %s - %s", resp.Status, string(body))
	}

	var listResponse VercelListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to decode blob list response: %w", err)
	}

	return &listResponse, nil
}

// constructBlobURL constructs the direct Vercel Blob URL for a feed file.
// Vercel Blob URLs follow the pattern: https://blob.vercel-storage.com/{accountId}/{pathname}
func constructBlobURL(pathname string) string {
	return fmt.Sprintf("%s/%s", vercelBlobBaseURL, pathname)
}

// ListTldrFeedDates lists all available feed dates from blob storage.
// First attempts to read from a cached dates-index.json file for performance.
// Falls back to listing all blobs if the index doesn't exist (graceful degradation).
func ListTldrFeedDates() ([]string, error) {
	const indexPathname = "tldr-summaries/dates-index.json"
	indexURL := fmt.Sprintf("%s/%s", vercelBlobBaseURL, indexPathname)

	// Try to fetch the index file first (fast path)
	resp, err := http.Get(indexURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer func() { _ = resp.Body.Close() }()
		var indexFile struct {
			LastUpdated string   `json:"lastUpdated"`
			Dates       []string `json:"dates"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&indexFile); err == nil {
			return indexFile.Dates, nil
		}
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	// TODO: Fallback disabled - index should always be available after migration period
	// Uncomment if needed for emergency recovery:
	/*
	// Fallback: List blobs manually (expensive, but ensures data integrity)
	listResponse, err := listBlobsManually(tldrFeedsPrefix)
	if err != nil {
		return nil, fmt.Errorf("could not list tldr feeds from blob: %w", err)
	}

	var dates []string
	for _, blob := range listResponse.Blobs {
		// Filter out metadata and index files, only include feed files
		if !strings.Contains(blob.Pathname, "/metadata/") && !strings.Contains(blob.Pathname, "dates-index") && strings.HasSuffix(blob.Pathname, ".json") {
			// Extract date from "tldr-feeds/YYYY-MM-DD.json"
			fileName := strings.TrimSuffix(blob.Pathname, ".json")
			parts := strings.Split(fileName, "/")
			if len(parts) > 1 {
				dates = append(dates, parts[len(parts)-1])
			}
		}
	}

	// Sort dates descending (most recent first)
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	return dates, nil
	*/

	// Index file not found - this shouldn't happen in normal operation
	return nil, fmt.Errorf("dates index not found at %s", indexPathname)
}

// GetTldrFeedURL constructs the blob URL for a specific feed by date without fetching content.
// Returns the URL if the feed exists, empty string if not found.
func GetTldrFeedURL(date string) string {
	// Construct the blob pathname and URL directly
	// Pattern: tldr-feeds/YYYY-MM-DD.json
	pathname := fmt.Sprintf("%s%s.json", tldrFeedsPrefix, date)
	return constructBlobURL(pathname)
}

// GetTldrFeed fetches a specific feed by date from blob storage.
// Optimized: constructs URL directly instead of calling expensive list API.
func GetTldrFeed(date string) (*RssFeed, error) {
	blobURL := GetTldrFeedURL(date)

	resp, err := http.Get(blobURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed blob for date %s: %w", date, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Treat 404 as not found.
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status when fetching feed blob for date %s: %s", date, resp.Status)
	}

	var feed RssFeed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to decode feed content for date %s: %w", date, err)
	}

	return &feed, nil
}
