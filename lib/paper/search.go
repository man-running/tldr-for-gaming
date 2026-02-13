package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/lib/logger"
	"net/http"
	"net/url"
	"time"
)

const huggingFaceSearchURL = "https://huggingface.co/api/papers/search"

var (
	// Shared HTTP client - reusing the same client enables connection pooling
	// Go's http.Client automatically handles connection reuse, keep-alive, and gzip decompression
	hfHTTPClient = &http.Client{
		Timeout: 10 * time.Second,
		// Uses http.DefaultTransport which has connection pooling enabled by default:
		// - MaxIdleConns: 100 (default)
		// - MaxIdleConnsPerHost: 2 (default, but we can increase for better performance)
		// - IdleConnTimeout: 90s (default)
		// - DisableCompression: false (default - auto-decompresses gzip responses)
	}
)

// SearchResult represents a paper search result
type SearchResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	PublishedAt string `json:"publishedAt"`
}

// huggingFaceSearchItem represents a single search result from HuggingFace API
type huggingFaceSearchItem struct {
	Paper *huggingFacePaper `json:"paper"`
}

// huggingFacePaper represents the paper object from HuggingFace API
type huggingFacePaper struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	PublishedAt string `json:"publishedAt"`
}

// SearchPapersOnHuggingFace searches for papers on HuggingFace and returns simplified results
// Results are cached at CDN level, embeddings cached in vector DB for similarity search
// Automatically generates embeddings and reranks results if embedding service is available
func SearchPapersOnHuggingFace(ctx context.Context, query string) ([]SearchResult, error) {
	startTime := time.Now()
	
	// Start query embedding generation in parallel with HuggingFace fetch
	// (query embedding doesn't depend on results, so we can do it early)
	embeddingService, _ := GetEmbeddingService()
	queryEmbeddingChan := make(chan struct {
		embedding []float32
		err       error
	}, 1)
	
	if embeddingService != nil {
		go func() {
			embedding, err := embeddingService.GenerateEmbedding(ctx, query)
			queryEmbeddingChan <- struct {
				embedding []float32
				err       error
			}{embedding: embedding, err: err}
		}()
	} else {
		queryEmbeddingChan <- struct {
			embedding []float32
			err       error
		}{embedding: nil, err: nil}
	}
	
	results, err := fetchFromHuggingFace(ctx, query)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil, context.Canceled
		}
		return nil, err
	}
	
	queryEmbeddingRes := <-queryEmbeddingChan
	var queryEmbedding []float32
	if queryEmbeddingRes.err == nil && queryEmbeddingRes.embedding != nil {
		queryEmbedding = queryEmbeddingRes.embedding
	}
	
	rerankedResults, finalQueryEmbedding := tryRerankWithEmbeddings(ctx, query, results, queryEmbedding)
	totalDuration := time.Since(startTime)
	
	logger.Info("Search completed", map[string]interface{}{
		"query":        query,
		"result_count": len(rerankedResults),
		"duration_ms":  totalDuration.Milliseconds(),
	})
	
	if finalQueryEmbedding != nil {
		cache := GetVectorDBCache()
		queryHash := cache.HashQuery(query)
		_ = cache.AddEmbeddingWithText(queryHash, query, finalQueryEmbedding)
	}
	
	return rerankedResults, nil
}

// tryRerankWithEmbeddings attempts to generate embeddings and rerank results
// queryEmbedding: optional pre-generated query embedding (if nil, will generate)
// Returns reranked results and query embedding (or original results if embedding fails)
func tryRerankWithEmbeddings(ctx context.Context, query string, results []SearchResult, queryEmbedding []float32) ([]SearchResult, []float32) {
	cache := GetVectorDBCache()
	
	// Get or generate query embedding
	if cache.dbEnabled {
		maxWait := 2 * time.Second
		checkInterval := 50 * time.Millisecond
		waited := time.Duration(0)
		
		for !IsDBEnabled() && waited < maxWait {
			time.Sleep(checkInterval)
			waited += checkInterval
		}
		
		if IsDBEnabled() && queryEmbedding == nil {
			queryHash := cache.HashQuery(query)
			emb, err := cache.GetQueryEmbedding(ctx, queryHash)
			if err == nil && emb != nil {
				queryEmbedding = emb
			}
		}
	}
	
	if queryEmbedding == nil {
		embeddingService, err := GetEmbeddingService()
		if err != nil {
			return results, nil
		}
		
		emb, err := embeddingService.GenerateEmbedding(ctx, query)
		if err != nil {
			return results, nil
		}
		queryEmbedding = emb
	}
	
	// Step 3: Do similarity search in DB to get top K results (fast path)
	// This uses the HNSW index for optimized vector search
	if cache.dbEnabled && IsDBEnabled() {
		// Create result map for quick lookup
		resultMap := make(map[string]SearchResult, len(results))
		for _, result := range results {
			resultMap[result.ID] = result
		}
		
		// Do similarity search - get top K from all stored embeddings
		topK := len(results)
		if topK > 200 {
			topK = 200 // Limit for performance
		}
		
		similarPaperIDs, err := cache.SearchSimilarInDB(ctx, queryEmbedding, topK)
		if err == nil && len(similarPaperIDs) > 0 {
			// Build reranked results from similarity search
			reranked := make([]SearchResult, 0, len(similarPaperIDs))
			seen := make(map[string]bool, len(similarPaperIDs))
			
			// Add results that match our input results (in similarity order)
			for _, paperID := range similarPaperIDs {
				if result, exists := resultMap[paperID]; exists && !seen[paperID] {
					reranked = append(reranked, result)
					seen[paperID] = true
				}
			}
			
			// Add any results that weren't in similarity search
			for _, result := range results {
				if !seen[result.ID] {
					reranked = append(reranked, result)
				}
			}
			
			// Backfill missing embeddings in background
			go func() {
				bgCtx := context.Background()
				embeddingService, err := GetEmbeddingService()
				if err != nil {
					return
				}
				
				paperIDs := make([]string, 0, len(results))
				for _, result := range results {
					paperIDs = append(paperIDs, result.ID)
				}
				
				existingEmbeddings, err := cache.GetResultEmbeddingsBatch(bgCtx, paperIDs)
				if err != nil {
					return
				}
				
				missingTexts := make([]string, 0)
				missingResults := make([]SearchResult, 0)
				for _, result := range results {
					if _, exists := existingEmbeddings[result.ID]; !exists {
						text := result.Title
						if result.Summary != "" {
							text += ". " + result.Summary
						}
						if text != "" {
							missingTexts = append(missingTexts, text)
							missingResults = append(missingResults, result)
						}
					}
				}
				
				if len(missingTexts) > 0 {
					embeddings, err := embeddingService.GenerateEmbeddings(bgCtx, missingTexts)
					if err == nil && len(embeddings) == len(missingResults) {
						embeddingsToStore := make(map[string][]float32)
						for i, result := range missingResults {
							if i < len(embeddings) {
								embeddingsToStore[result.ID] = embeddings[i]
							}
						}
						if len(embeddingsToStore) > 0 {
							_ = cache.AddResultEmbeddingsBatch(embeddingsToStore)
						}
					}
				}
			}()
			
			return reranked, queryEmbedding
		}
	}
	
	return results, queryEmbedding
}

// RerankSearchResults reranks existing search results with embeddings
func RerankSearchResults(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error) {
	return RerankSearchResultsWithEmbedding(ctx, query, results, nil)
}

// RerankSearchResultsWithEmbedding reranks results with optional pre-generated query embedding
// Reranks all results - embeddings are cached and generated in parallel batches for efficiency
func RerankSearchResultsWithEmbedding(ctx context.Context, query string, results []SearchResult, queryEmbedding []float32) ([]SearchResult, error) {
	reranked, _ := tryRerankWithEmbeddings(ctx, query, results, queryEmbedding)
	return reranked, nil
}

// SearchPapersOnHuggingFaceWithRerank searches for papers and reranks them using provided embeddings
// queryEmbedding: embedding vector for the search query (optional, will generate if nil)
// resultEmbeddings: embeddings for each result (optional, will generate if nil)
func SearchPapersOnHuggingFaceWithRerank(
	ctx context.Context,
	query string,
	queryEmbedding []float32,
	resultEmbeddings [][]float32,
) ([]SearchResult, error) {
	// Fetch from HuggingFace API (CDN handles result caching)
	results, err := fetchFromHuggingFace(ctx, query)
	if err != nil {
		return nil, err
	}
	
	// Generate embeddings if not provided
	embeddingService, _ := GetEmbeddingService()
	if embeddingService != nil {
		if queryEmbedding == nil {
			if qe, err := embeddingService.GenerateEmbedding(ctx, query); err == nil {
				queryEmbedding = qe
			}
		}
		
		if resultEmbeddings == nil && queryEmbedding != nil {
			resultTexts := make([]string, len(results))
			for i, result := range results {
				resultTexts[i] = result.Title
				if result.Summary != "" {
					resultTexts[i] += ". " + result.Summary
				}
			}
			if re, err := embeddingService.GenerateEmbeddings(ctx, resultTexts); err == nil {
				resultEmbeddings = re
			}
		}
	}
	
	// Rerank if embeddings provided
	cache := GetVectorDBCache()
	if queryEmbedding != nil && resultEmbeddings != nil && len(resultEmbeddings) == len(results) {
		reranked, err := cache.RerankResults(ctx, queryEmbedding, results, resultEmbeddings)
		if err == nil {
			results = reranked
		}
	}
	
	if queryEmbedding != nil {
		queryHash := cache.HashQuery(query)
		_ = cache.AddEmbeddingWithText(queryHash, query, queryEmbedding)
	}
	
	return results, nil
}

// fetchFromHuggingFace performs the actual API call to HuggingFace
// Results are cached at CDN level via middleware, embeddings cached in embedding service
func fetchFromHuggingFace(ctx context.Context, query string) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("%s?q=%s", huggingFaceSearchURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Takara-TLDR/1.0")
	// Note: HuggingFace API does not compress responses, so no gzip handling needed
	// Go's http.Client automatically uses HTTP/2 and connection pooling via shared client

	resp, err := hfHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var hfResults []huggingFaceSearchItem
	if err := json.Unmarshal(body, &hfResults); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(hfResults))
	for _, item := range hfResults {
		if item.Paper != nil && item.Paper.ID != "" && item.Paper.Title != "" {
			results = append(results, SearchResult{
				ID:          item.Paper.ID,
				Title:       item.Paper.Title,
				Summary:     item.Paper.Summary,
				PublishedAt: item.Paper.PublishedAt,
			})
		}
	}

	return results, nil
}

