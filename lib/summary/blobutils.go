package summary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"main/lib/logger"
)

const (
	summaryPrefix = "tldr-summaries/"
	papersPrefix  = "tldr-papers/"
)

// isBlobCacheDisabled checks if DISABLE_BLOB_CACHE environment variable is set to true
func isBlobCacheDisabled() bool {
	disableStr := os.Getenv("DISABLE_BLOB_CACHE")
	if disableStr == "" {
		return false
	}
	disabled, err := strconv.ParseBool(disableStr)
	if err != nil {
		// If parsing fails, default to false (cache enabled)
		return false
	}
	return disabled
}

type SummaryMetadata struct {
	Date      string `json:"date"`
	WordCount int    `json:"wordCount"`
	CachedAt  string `json:"cachedAt"`
}

// DateIndexFile represents the structure of the dates index file in blob storage
type DateIndexFile struct {
	LastUpdated string   `json:"lastUpdated"`
	Dates       []string `json:"dates"`
}

// listBlobsManually performs a GET request to the Vercel Blob List API.
func listBlobsManually(prefix string) ([]VercelListBlob, error) {
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("BLOB_READ_WRITE_TOKEN environment variable not set")
	}

	req, err := http.NewRequest("GET", "https://blob.vercel-storage.com", nil)
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

	var listResponse struct {
		Blobs []VercelListBlob `json:"blobs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to decode blob list response: %w", err)
	}

	return listResponse.Blobs, nil
}

// VercelListBlob represents a single blob item in the Vercel Blob List API response.
type VercelListBlob struct {
	URL      string `json:"url"`
	Pathname string `json:"pathname"`
}

// GetLatestSummaryURL retrieves the blob URL for the most recent summary without fetching content.
// Returns empty string if not found.
func GetLatestSummaryURL() (string, error) {
	if isBlobCacheDisabled() {
		return "", nil // Return empty to indicate no cache found
	}

	blobs, err := listBlobsManually(summaryPrefix)
	if err != nil {
		return "", fmt.Errorf("could not list summaries from blob: %w", err)
	}

	var summaryBlobs []VercelListBlob
	for _, blob := range blobs {
		// Filter out metadata files and only include .xml files
		if !strings.Contains(blob.Pathname, "/metadata/") && strings.HasSuffix(blob.Pathname, ".xml") {
			summaryBlobs = append(summaryBlobs, blob)
		}
	}

	if len(summaryBlobs) == 0 {
		return "", nil // No cached summary found, not an error
	}

	// Sort by pathname (which includes the date) descending to find the latest
	sort.Slice(summaryBlobs, func(i, j int) bool {
		return summaryBlobs[i].Pathname > summaryBlobs[j].Pathname
	})

	return summaryBlobs[0].URL, nil
}

// GetLatestSummary fetches the most recent summary from Vercel Blob storage.
func GetLatestSummary() ([]byte, error) {
	blobURL, err := GetLatestSummaryURL()
	if err != nil {
		return nil, err
	}
	if blobURL == "" {
		return nil, nil // No cached summary found
	}

	// Fetch the content of the latest blob
	resp, err := http.Get(blobURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest summary blob content: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status when fetching latest summary blob: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read summary blob content: %w", err)
	}

	return content, nil
}

// GetLatestPapersURL retrieves the blob URL for the most recent papers without fetching content.
// Returns empty string if not found.
func GetLatestPapersURL() (string, error) {
	if isBlobCacheDisabled() {
		return "", nil // Return empty to indicate no cache found
	}

	blobs, err := listBlobsManually(papersPrefix)
	if err != nil {
		return "", fmt.Errorf("could not list papers from blob: %w", err)
	}

	var papersBlobs []VercelListBlob
	for _, blob := range blobs {
		// Filter out metadata files and only include .xml files
		if !strings.Contains(blob.Pathname, "/metadata/") && strings.HasSuffix(blob.Pathname, ".xml") {
			papersBlobs = append(papersBlobs, blob)
		}
	}

	if len(papersBlobs) == 0 {
		return "", nil // No cached papers found, not an error
	}

	// Sort by pathname (which includes the date) descending to find the latest
	sort.Slice(papersBlobs, func(i, j int) bool {
		return papersBlobs[i].Pathname > papersBlobs[j].Pathname
	})

	return papersBlobs[0].URL, nil
}

// GetLatestPapers fetches the most recent papers from Vercel Blob storage.
func GetLatestPapers() ([]byte, error) {
	blobURL, err := GetLatestPapersURL()
	if err != nil {
		return nil, err
	}
	if blobURL == "" {
		return nil, nil // No cached papers found
	}

	// Fetch the content of the latest blob
	resp, err := http.Get(blobURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest papers blob content: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch papers blob: status code %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read papers blob content: %w", err)
	}

	return content, nil
}

// StorePapers stores papers in Vercel Blob storage.
func StorePapers(papersData []byte) error {
	if isBlobCacheDisabled() {
		return nil // Silently skip storing to cache
	}

	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return fmt.Errorf("BLOB_READ_WRITE_TOKEN environment variable not set")
	}

	// Generate filename based on current date
	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")
	blobPath := papersPrefix + dateStr + ".xml"

	putURL := fmt.Sprintf("https://blob.vercel-storage.com/%s", blobPath)
	req, err := http.NewRequest("PUT", putURL, bytes.NewBuffer(papersData))
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/rss+xml")
	req.Header.Set("x-add-random-suffix", "0")
	req.Header.Set("x-cache-control-max-age", "31536000") // 1 year

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute PUT request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("blob storage returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// StoreSummary stores a summary in Vercel Blob storage.
func StoreSummary(summaryData []byte) error {
	if isBlobCacheDisabled() {
		return nil // Silently skip storing to cache
	}

	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return fmt.Errorf("BLOB_READ_WRITE_TOKEN environment variable not set")
	}

	// Generate filename based on current date
	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")
	blobPath := summaryPrefix + dateStr + ".xml"

	putURL := fmt.Sprintf("https://blob.vercel-storage.com/%s", blobPath)
	req, err := http.NewRequest("PUT", putURL, bytes.NewBuffer(summaryData))
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/rss+xml")
	req.Header.Set("x-add-random-suffix", "0")
	req.Header.Set("x-cache-control-max-age", "31536000") // 1 year
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute PUT request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("blob storage PUT API returned non-200 status: %s - %s", resp.Status, string(body))
	}

	// Store metadata
	metadata := SummaryMetadata{
		Date:      dateStr,
		WordCount: len(strings.Fields(string(summaryData))), // Rough word count
		CachedAt:  now.Format(time.RFC3339),
	}

	metadataBlobPath := summaryPrefix + "metadata/" + dateStr + ".json"
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metaPutURL := fmt.Sprintf("https://blob.vercel-storage.com/%s", metadataBlobPath)
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

	// Update the dates index file for fast retrieval
	if err := UpdateDatesIndex(); err != nil {
		// Log the error but don't fail the entire operation
		// The dates index is a performance optimization; if it fails, the fallback list API still works
		logger.Warn("failed to update dates index", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return nil
}

// UpdateDatesIndex rebuilds and stores the dates index file in blob storage.
// This is called after storing a new daily summary to keep the index fresh.
func UpdateDatesIndex() error {
	if isBlobCacheDisabled() {
		return nil
	}

	// Fetch all current dates from blob storage
	dates, err := listTldrFeedDatesInternal()
	if err != nil {
		return fmt.Errorf("failed to list dates for index update: %w", err)
	}

	// Create index file
	indexFile := DateIndexFile{
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Dates:       dates,
	}

	indexJSON, err := json.Marshal(indexFile)
	if err != nil {
		return fmt.Errorf("failed to marshal dates index: %w", err)
	}

	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return fmt.Errorf("BLOB_READ_WRITE_TOKEN environment variable not set")
	}

	// Write index file
	indexPath := summaryPrefix + "dates-index.json"
	indexURL := fmt.Sprintf("https://blob.vercel-storage.com/%s", indexPath)
	indexReq, err := http.NewRequest("PUT", indexURL, bytes.NewBuffer(indexJSON))
	if err != nil {
		return fmt.Errorf("failed to create PUT request for index: %w", err)
	}

	indexReq.Header.Set("Authorization", "Bearer "+token)
	indexReq.Header.Set("Content-Type", "application/json")
	indexReq.Header.Set("x-add-random-suffix", "0")
	indexReq.Header.Set("x-cache-control-max-age", "3600") // 1 hour, more frequent updates

	client := &http.Client{Timeout: 15 * time.Second}
	indexResp, err := client.Do(indexReq)
	if err != nil {
		return fmt.Errorf("failed to execute PUT request for index: %w", err)
	}
	defer func() { _ = indexResp.Body.Close() }()

	if indexResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(indexResp.Body)
		return fmt.Errorf("blob storage PUT API returned non-200 status for index: %s - %s", indexResp.Status, string(body))
	}

	return nil
}

// listTldrFeedDatesInternal fetches and returns a sorted list of dates from the tldr-summaries/ directory.
func listTldrFeedDatesInternal() ([]string, error) {
	blobs, err := listBlobsManually(summaryPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list summaries for dates: %w", err)
	}

	var dates []string
	for _, blob := range blobs {
		// Filter out metadata files and only include .xml files
		if !strings.Contains(blob.Pathname, "/metadata/") && strings.HasSuffix(blob.Pathname, ".xml") {
			// Extract date from pathname (e.g., "tldr-summaries/2023-10-27.xml")
			dateStr := strings.TrimPrefix(blob.Pathname, summaryPrefix)
			dateStr = strings.TrimSuffix(dateStr, ".xml")
			dates = append(dates, dateStr)
		}
	}

	// Sort dates descending
	sort.Slice(dates, func(i, j int) bool {
		return dates[i] > dates[j]
	})

	return dates, nil
}
