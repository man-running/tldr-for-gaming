package paper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	vercelBlobAPIURL = "https://blob.vercel-storage.com"
	papersPrefix     = "papers/"
	metadataPrefix   = "metadata/"
)

// VercelListBlob is a simplified representation of a blob item.
type VercelListBlob struct {
	URL string `json:"url"`
}

// VercelListResponse is the structure of the list API response.
type VercelListResponse struct {
	Blobs []VercelListBlob `json:"blobs"`
}

// GetPaperURL retrieves the blob URL for a paper without fetching the content.
// Returns empty string if not found.
func GetPaperURL(arxivId string) (string, error) {
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return "", fmt.Errorf("BLOB_READ_WRITE_TOKEN not set")
	}
	blobPath := papersPrefix + arxivId + ".json"

	// We must list to get the full public URL, as it contains a hash.
	req, err := http.NewRequest("GET", vercelBlobAPIURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create list request for get: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	q := req.URL.Query()
	q.Add("prefix", blobPath)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute list request for get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("blob list API returned non-200 for get: %s", resp.Status)
	}

	var listResponse VercelListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return "", fmt.Errorf("failed to decode blob list response for get: %w", err)
	}

	if len(listResponse.Blobs) == 0 {
		return "", nil // Not found, which is a valid cache miss.
	}

	return listResponse.Blobs[0].URL, nil
}

// GetPaper retrieves a paper's data from Vercel Blob storage. Returns nil if not found.
func GetPaper(arxivId string) (*PaperData, error) {
	blobURL, err := GetPaperURL(arxivId)
	if err != nil {
		return nil, err
	}
	if blobURL == "" {
		return nil, nil // Not found
	}

	// Fetch the actual blob content
	contentResp, err := http.Get(blobURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blob content: %w", err)
	}
	defer func() { _ = contentResp.Body.Close() }()

	if contentResp.StatusCode == http.StatusNotFound {
		return nil, nil // Not found
	}
	if contentResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get blob content, status: %s", contentResp.Status)
	}

	var paper PaperData
	if err := json.NewDecoder(contentResp.Body).Decode(&paper); err != nil {
		return nil, fmt.Errorf("failed to decode paper JSON from blob: %w", err)
	}

	return &paper, nil
}

// StorePaper saves a paper's data to Vercel Blob storage.
func StorePaper(arxivId string, paper *PaperData) error {
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return fmt.Errorf("BLOB_READ_WRITE_TOKEN not set")
	}
	blobPath := papersPrefix + arxivId + ".json"

	jsonData, err := json.Marshal(paper)
	if err != nil {
		return fmt.Errorf("failed to marshal paper data for storage: %w", err)
	}

	// The Vercel Blob API for PUT requires the pathname in the URL.
	putURL := fmt.Sprintf("%s/%s", vercelBlobAPIURL, blobPath)
	req, err := http.NewRequest("PUT", putURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	// These headers are hints for the blob store.
	req.Header.Set("x-add-random-suffix", "0")
	req.Header.Set("x-cache-control-max-age", "31536000") // 1 year

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute PUT request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("blob store PUT returned non-200 status: %s - %s", resp.Status, string(body))
	}

	// Store metadata for quick listing
	metadata := PaperMetadata{
		Title:         paper.Title,
		Authors:       paper.Authors,
		PublishedDate: paper.PublishedDate,
		ArxivID:       arxivId,
		CachedAt:      time.Now().Format(time.RFC3339),
	}

	metadataBlobPath := metadataPrefix + arxivId + ".json"
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
