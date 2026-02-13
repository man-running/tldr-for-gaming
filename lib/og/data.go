package og

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/lib/analytics"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	papersPrefix   = "papers/"
	metadataPrefix = "metadata/"
)

var (
	// In-memory cache for titles to persist across warm invocations.
	titleCache = struct {
		sync.RWMutex
		m map[string]string
	}{m: make(map[string]string)}

	httpClient = &http.Client{Timeout: 5 * time.Second}
)

// fetchBlobObject fetches and unmarshals a JSON object from Vercel Blob.
func fetchBlobObject(ctx context.Context, key string, v interface{}) error {
	token := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if token == "" {
		return fmt.Errorf("BLOB_READ_WRITE_TOKEN not set")
	}

	// list first to get the full URL
	url := "https://blob.vercel-storage.com"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create list request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	q := req.URL.Query()
	q.Add("prefix", key)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute list request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("blob list API returned non-200: %s", resp.Status)
	}

	var listResponse struct {
		Blobs []struct {
			URL string `json:"url"`
		} `json:"blobs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return fmt.Errorf("failed to decode blob list response: %w", err)
	}

	if len(listResponse.Blobs) == 0 {
		return fmt.Errorf("blob not found: %s", key)
	}

	// Fetch the actual blob content using the full URL
	contentResp, err := http.Get(listResponse.Blobs[0].URL)
	if err != nil {
		return fmt.Errorf("failed to fetch blob content: %w", err)
	}
	defer func() { _ = contentResp.Body.Close() }()

	if contentResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get blob content, status: %s", contentResp.Status)
	}

	body, err := io.ReadAll(contentResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read blob body: %w", err)
	}

	return json.Unmarshal(body, v)
}

// GetTitle fetches a paper's title, using an in-memory cache and falling back to Vercel Blob.
func GetTitle(arxivId string) (string, error) {
	// 1. Check in-memory cache first
	titleCache.RLock()
	cachedTitle, found := titleCache.m[arxivId]
	titleCache.RUnlock()
	if found {
		return cachedTitle, nil
	}

	// 2. If not cached, fetch from Vercel Blob
	var wg sync.WaitGroup
	var metadata PaperMetadata
	var paper PaperData
	var metaErr, paperErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		key := fmt.Sprintf("%s%s.json", metadataPrefix, arxivId)
		metaErr = fetchBlobObject(context.TODO(), key, &metadata)
	}()
	go func() {
		defer wg.Done()
		key := fmt.Sprintf("%s%s.json", papersPrefix, arxivId)
		paperErr = fetchBlobObject(context.TODO(), key, &paper)
	}()
	wg.Wait()

	// 3. Resolve title from fetched data (prefer metadata)
	resolvedTitle := ""
	if metaErr == nil && strings.TrimSpace(metadata.Title) != "" {
		resolvedTitle = metadata.Title
	} else if paperErr == nil && strings.TrimSpace(paper.Title) != "" {
		resolvedTitle = paper.Title
	}

	if resolvedTitle != "" {
		_ = analytics.Track("og_image_generated", arxivId, map[string]interface{}{"arxiv_id": arxivId})
	}

	if resolvedTitle == "" {
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			return "AI Research Paper", fmt.Errorf("BASE_URL environment variable not set")
		}

		paperURL := fmt.Sprintf("%s/api/paper?id=%s", baseURL, arxivId)
		resp, err := http.Get(paperURL)
		if err != nil {
			return "AI Research Paper", fmt.Errorf("failed to fetch paper title from %s: %w", paperURL, err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return "AI Research Paper", fmt.Errorf("failed to get paper title, status: %s", resp.Status)
		}

		var paperData struct {
			Data struct {
				Title string `json:"title"`
			} `json:"data"`
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "AI Research Paper", fmt.Errorf("failed to read paper body: %w", err)
		}

		if err := json.Unmarshal(body, &paperData); err != nil {
			return "AI Research Paper", fmt.Errorf("failed to unmarshal paper data: %w", err)
		}

		if strings.TrimSpace(paperData.Data.Title) != "" {
			resolvedTitle = paperData.Data.Title
			finalTitle := shortenTitle(resolvedTitle)
			titleCache.Lock()
			titleCache.m[arxivId] = finalTitle
			titleCache.Unlock()
			return finalTitle, nil
		}

		return "AI Research Paper", nil // Default title
	}

	// 4. Shorten title if necessary and cache it
	finalTitle := shortenTitle(resolvedTitle)
	titleCache.Lock()
	titleCache.m[arxivId] = finalTitle
	titleCache.Unlock()

	return finalTitle, nil
}

// shortenTitle replicates the title shortening logic from the original source.
func shortenTitle(title string) string {
	trimmedTitle := strings.TrimSpace(title)
	if len(trimmedTitle) > 120 {
		if colonIndex := strings.Index(trimmedTitle, ":"); colonIndex > 0 {
			return strings.TrimSpace(trimmedTitle[:colonIndex])
		}
	}
	return trimmedTitle
}
