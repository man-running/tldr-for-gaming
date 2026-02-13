package feed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"main/lib/logger"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	vercelBlobAPIURL = "https://blob.vercel-storage.com"
	tldrFeedsPrefix  = "tldr-feeds/"
	metadataPrefix   = "metadata/"
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

// GetLatestTldrFeed fetches the most recent TLDR feed from Vercel Blob storage.
func GetLatestTldrFeed() (*RssFeed, error) {
	listResponse, err := listBlobsManually(tldrFeedsPrefix)
	if err != nil {
		return nil, fmt.Errorf("could not list tldr feeds from blob: %w", err)
	}

	var feedBlobs []VercelListBlob
	for _, blob := range listResponse.Blobs {
		// Filter out metadata files
		if !strings.Contains(blob.Pathname, "/metadata/") && strings.HasSuffix(blob.Pathname, ".json") {
			feedBlobs = append(feedBlobs, blob)
		}
	}

	if len(feedBlobs) == 0 {
		return nil, nil // No cached feed found, not an error
	}

	// Sort by pathname (which includes the date) descending to find the latest
	sort.Slice(feedBlobs, func(i, j int) bool {
		return feedBlobs[i].Pathname > feedBlobs[j].Pathname
	})

	latestBlob := feedBlobs[0]

	// Fetch the content of the latest blob
	resp, err := http.Get(latestBlob.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest feed blob content: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status when fetching latest feed blob: %s", resp.Status)
	}

	var feed RssFeed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to decode latest feed content: %w", err)
	}

	return &feed, nil
}

func StoreTldrFeed(feed *RssFeed) error {
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		logger.Error("BLOB_READ_WRITE_TOKEN environment variable not set", nil, nil)
		return fmt.Errorf("BLOB_READ_WRITE_TOKEN environment variable not set")
	}

	// Robustly parse LastBuildDate across common layouts; fall back to now
	var pubDate time.Time
	layouts := []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, time.RubyDate, time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, feed.LastBuildDate); err == nil {
			pubDate = t.UTC()
			break
		}
	}
	if pubDate.IsZero() {
		pubDate = time.Now().UTC()
	}
	pathDate := pubDate.Format("2006-01-02")

	blobPath := tldrFeedsPrefix + pathDate + ".json"
	jsonData, err := json.Marshal(feed)
	if err != nil {
		return fmt.Errorf("failed to marshal feed data: %w", err)
	}

	putURL := fmt.Sprintf("%s/%s", vercelBlobAPIURL, blobPath)
	req, err := http.NewRequest("PUT", putURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-add-random-suffix", "0")
	req.Header.Set("x-cache-control-max-age", "31536000") // 1 year

	client := &http.Client{Timeout: 15 * time.Second}
	resp2, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute PUT request: %w", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	if resp2.StatusCode != http.StatusOK {
		return fmt.Errorf("blob storage PUT API returned non-200 status: %s", resp2.Status)
	}

	// Store metadata for quick listing
	metadata := TldrFeedMetadata{
		Title:         feed.Title,
		Description:   feed.Description,
		LastBuildDate: feed.LastBuildDate,
		ItemCount:     len(feed.Items),
		CachedAt:      time.Now().Format(time.RFC3339),
	}

	metadataBlobPath := tldrFeedsPrefix + metadataPrefix + pathDate + ".json"
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metaPutURL := fmt.Sprintf("%s/%s", vercelBlobAPIURL, metadataBlobPath)
	metaReq, err := http.NewRequest("PUT", metaPutURL, bytes.NewBuffer(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to create PUT request for metadata: %w", err)
	}

	metaReq.Header.Set("Authorization", "Bearer "+token)
	metaReq.Header.Set("Content-Type", "application/json")
	metaReq.Header.Set("x-add-random-suffix", "0")
	metaReq.Header.Set("x-cache-control-max-age", "31536000") // 1 year

	metaResp, err := client.Do(metaReq)
	if err != nil {
		return fmt.Errorf("failed to execute PUT request for metadata: %w", err)
	}
	defer func() { _ = metaResp.Body.Close() }()

	if metaResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(metaResp.Body)
		return fmt.Errorf("blob storage PUT API returned non-200 status for metadata: %s - %s", metaResp.Status, string(body))
	}

	return nil
}
