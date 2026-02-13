package media

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	vercelBlobAPIURL  = "https://blob.vercel-storage.com"
	vercelBlobBaseURL = "https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com"
)

// StoreImageBlob stores an image in Vercel Blob storage and returns the public URL
func StoreImageBlob(key string, imageData []byte, contentType string) (string, error) {
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return "", fmt.Errorf("BLOB_READ_WRITE_TOKEN not set")
	}

	// The Vercel Blob API for PUT requires the pathname in the URL
	putURL := fmt.Sprintf("%s/%s", vercelBlobAPIURL, key)
	req, err := http.NewRequest("PUT", putURL, bytes.NewBuffer(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to create PUT request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)
	// Use x-add-random-suffix: 0 for deterministic URLs
	req.Header.Set("x-add-random-suffix", "0")
	req.Header.Set("x-cache-control-max-age", "31536000") // 1 year

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute PUT request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("blob store PUT returned non-200 status: %s - %s", resp.Status, string(body))
	}

	// Construct public URL directly (since we use x-add-random-suffix: 0, the URL is deterministic)
	publicURL := fmt.Sprintf("%s/%s", vercelBlobBaseURL, key)
	return publicURL, nil
}

