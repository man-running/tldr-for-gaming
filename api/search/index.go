package handler

import (
	"encoding/json"
	"main/lib/logger"
	"main/lib/middleware"
	"main/lib/paper"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// embedQueryHandler generates query embedding (GET /api/search?q={query})
func embedQueryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	// Get query from URL query parameter
	queryParam := r.URL.Query().Get("q")
	if queryParam == "" {
		logger.Warn("Empty search query", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "Query parameter 'q' is required",
		})
		return
	}

	// URL decode the query
	decodedQuery, err := url.QueryUnescape(queryParam)
	if err != nil {
		decodedQuery = queryParam
	}

	query := strings.TrimSpace(decodedQuery)
	if query == "" {
		logger.Warn("Empty search query after trim", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "Query is required",
		})
		return
	}

	ctx["search_query"] = query
	logger.Debug("Generating query embedding", ctx)

	embeddingService, err := paper.GetEmbeddingService()
	if err != nil {
		logger.Error("Embedding service initialization failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, middleware.ErrorResponse{
			Error: "Failed to initialize embedding service",
		})
		return
	}

	embedding, err := embeddingService.GenerateEmbedding(r.Context(), query)
	if err != nil {
		logger.Error("Query embedding generation failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, middleware.ErrorResponse{
			Error: "Failed to generate query embedding",
		})
		return
	}

	middleware.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"queryEmbedding": embedding,
	})
}

// embedBatchHandler generates embeddings for multiple texts (GET /api/search?text=hello&text=you)
func embedBatchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	// Get all text parameters from URL
	textParams := r.URL.Query()["text"]
	if len(textParams) == 0 {
		logger.Warn("No text parameters provided", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "At least one 'text' parameter is required",
		})
		return
	}

	// Limit to max batch size (32)
	const maxBatchSize = 32
	if len(textParams) > maxBatchSize {
		logger.Warn("Too many text parameters", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "Maximum 32 text parameters allowed",
		})
		return
	}

	// URL decode and trim all texts
	texts := make([]string, 0, len(textParams))
	for _, textParam := range textParams {
		decodedText, err := url.QueryUnescape(textParam)
		if err != nil {
			decodedText = textParam
		}
		trimmedText := strings.TrimSpace(decodedText)
		if trimmedText != "" {
			texts = append(texts, trimmedText)
		}
	}

	if len(texts) == 0 {
		logger.Warn("No valid text parameters after processing", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "At least one valid text parameter is required",
		})
		return
	}

	ctx["text_count"] = len(texts)
	logger.Debug("Generating batch embeddings", ctx)

	embeddingService, err := paper.GetEmbeddingService()
	if err != nil {
		logger.Error("Embedding service initialization failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, middleware.ErrorResponse{
			Error: "Failed to initialize embedding service",
		})
		return
	}

	embeddings, err := embeddingService.GenerateEmbeddings(r.Context(), texts)
	if err != nil {
		logger.Error("Batch embedding generation failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, middleware.ErrorResponse{
			Error: "Failed to generate embeddings",
		})
		return
	}

	// Return single embedding if one text, array if multiple
	if len(embeddings) == 1 {
		middleware.WriteJSONResponse(w, http.StatusOK, embeddings[0])
	} else {
		middleware.WriteJSONResponse(w, http.StatusOK, embeddings)
	}
}

// rerankHandler reranks HuggingFace results with embeddings (POST /api/search)
func rerankHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)
	requestStart := time.Now()

	// Initialize database connection and schema in parallel (non-blocking)
	// This starts early so it's ready when needed during reranking
	// InitDB() already calls InitSchema() internally, so we just need to call InitDB()
	go func() {
		_ = paper.InitDB()
	}()

	var requestBody struct {
		Query          string                  `json:"query"`
		Results        []paper.SearchResult    `json:"results"`
		QueryEmbedding []float32               `json:"queryEmbedding,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logger.Warn("Invalid request body", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	query := strings.TrimSpace(requestBody.Query)
	if query == "" {
		logger.Warn("Empty search query", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "Query is required",
		})
		return
	}
	
	// Security: Limit result array size to prevent DoS
	const maxResults = 1000
	if len(requestBody.Results) > maxResults {
		logger.Warn("Too many results in request", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "Too many results",
		})
		return
	}

	ctx["search_query"] = query
	ctx["result_count"] = len(requestBody.Results)
	logger.Info("Reranking HuggingFace results", ctx)

	// Separate first result (fastest match) from rest
	var firstResult *paper.SearchResult
	var resultsToRerank []paper.SearchResult
	if len(requestBody.Results) > 0 {
		firstResult = &requestBody.Results[0]
		resultsToRerank = requestBody.Results[1:]
	} else {
		resultsToRerank = requestBody.Results
	}

	rerankedResults, err := paper.RerankSearchResultsWithEmbedding(r.Context(), query, resultsToRerank, requestBody.QueryEmbedding)
	if err != nil {
		logger.Error("Reranking failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, middleware.ErrorResponse{
			Error: "Failed to rerank results",
		})
		return
	}

	// Prepend first result (fastest match) to reranked results
	var finalResults []paper.SearchResult
	if firstResult != nil {
		finalResults = append([]paper.SearchResult{*firstResult}, rerankedResults...)
	} else {
		finalResults = rerankedResults
	}

	totalDuration := time.Since(requestStart)
	ctx["result_count"] = len(finalResults)
	ctx["total_duration_ms"] = totalDuration.Milliseconds()
	logger.Info("Reranking completed", ctx)

	middleware.WriteJSONResponse(w, http.StatusOK, finalResults)
}

// applyCORSHeaders enables a permissive CORS policy for this API.
func applyCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// searchHandler routes requests based on HTTP method.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	applyCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Check if batch embedding is requested (text parameters present)
		if len(r.URL.Query()["text"]) > 0 {
			// GET /api/search?text=hello&text=you - Generate batch embeddings
			// Configure aggressive caching for batch embeddings
			// Model is stable for at least 1 year, so same texts = same embeddings
			cacheOpts := middleware.CacheOptions{
				Config: middleware.CacheConfig{
					MaxAge:               31536000, // 1 year browser cache
					SMaxAge:              31536000, // 1 year CDN cache
					StaleWhileRevalidate: 31536000, // 1 year stale-while-revalidate
					StaleIfError:         31536000, // 1 year stale-if-error
				},
				ETagKey: "batch-embedding",
				Enabled: true,
			}
			middleware.MethodAndCache(http.MethodGet, cacheOpts)(embedBatchHandler)(w, r)
		} else {
			// GET /api/search?q={query} - Generate query embedding
			// Configure aggressive caching for query embeddings
			// Model is stable for at least 1 year, so same query = same embedding
			cacheOpts := middleware.CacheOptions{
				Config: middleware.CacheConfig{
					MaxAge:               31536000, // 1 year browser cache
					SMaxAge:              31536000, // 1 year CDN cache
					StaleWhileRevalidate: 31536000, // 1 year stale-while-revalidate
					StaleIfError:         31536000, // 1 year stale-if-error
				},
				ETagKey: "query-embedding",
				Enabled: true,
			}
			middleware.MethodAndCache(http.MethodGet, cacheOpts)(embedQueryHandler)(w, r)
		}
	case http.MethodPost:
		// POST /api/search - Rerank results
		middleware.MethodValidator(http.MethodPost)(rerankHandler)(w, r)
	default:
		middleware.WriteJSONResponse(w, http.StatusMethodNotAllowed, middleware.ErrorResponse{
			Error: "Method not allowed",
		})
	}
}

// Handler is the Vercel serverless function entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	searchHandler(w, r)
}

