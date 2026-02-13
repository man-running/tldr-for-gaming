package handler

import (
	"encoding/json"
	"main/lib/logger"
	"main/lib/middleware"
	"main/lib/paper"
	"net/http"
)

// paperHandler contains the main logic for the paper endpoint
func paperHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	// Extract arxivId from the path, e.g., "/api/paper/1706.03762"
	arxivId := r.URL.Query().Get("id")
	ctx["arxiv_id"] = arxivId

	logger.Info("Processing paper request", ctx)

	// Validate input
	if arxivId == "" {
		logger.Warn("Missing arxiv_id parameter", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, paper.FinalApiResponse{
			Success: false,
			Error:   &paper.ApiError{Code: paper.ErrorCodeInvalidArxivID, Message: "arxiv_id parameter required"},
		})
		return
	}

	// 1. Get the raw paper data
	logger.Debug("Fetching paper data", ctx)
	result, err := paper.GetPaperRaw(arxivId)
	if err != nil {
		switch err.(type) {
		case *paper.InvalidIdError:
			logger.Warn("Invalid ArXiv ID provided", ctx)
			middleware.WriteJSONResponse(w, http.StatusBadRequest, paper.FinalApiResponse{
				Success: false,
				Error:   &paper.ApiError{Code: paper.ErrorCodeInvalidArxivID, Message: err.Error()},
			})
		case *paper.PaperNotFoundError:
			logger.Warn("Paper not found", ctx)
			middleware.WriteJSONResponse(w, http.StatusNotFound, paper.FinalApiResponse{
				Success: false,
				Error:   &paper.ApiError{Code: paper.ErrorCodePaperNotFound, Message: err.Error()},
			})
		default:
			logger.Error("Internal error fetching paper", err, ctx)
			middleware.WriteJSONResponse(w, http.StatusInternalServerError, paper.FinalApiResponse{
				Success: false,
				Error:   &paper.ApiError{Code: paper.ErrorCodeInternalError, Message: "Internal server error"},
			})
		}
		return
	}

	// 2. Prepare response data
	logger.Debug("Marshalling response data", ctx)
	payload, err := json.Marshal(paper.FinalApiResponse{
		Success: true,
		Data:    result.Data,
		BlobURL: result.BlobURL,
	})
	if err != nil {
		logger.Error("Failed to marshal response", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, paper.FinalApiResponse{
			Success: false,
			Error:   &paper.ApiError{Code: paper.ErrorCodeInternalError, Message: "Internal server error"},
		})
		return
	}

	// 3. Write response (caching is handled by middleware)
	logger.Debug("Sending response", ctx)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(payload); err != nil {
		logger.Error("Failed to write response", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	logger.Info("Paper request completed successfully", ctx)
}

// Handler is the Vercel serverless function entrypoint for the paper API.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Configure caching for paper endpoint
	cacheOpts := middleware.CacheOptions{
		Config: middleware.CacheConfig{
			MaxAge:               0,     // No browser caching
			SMaxAge:              3600,  // 1 hour CDN cache
			StaleWhileRevalidate: 86400, // 24 hours stale-while-revalidate
			StaleIfError:         86400, // 24 hours stale-if-error
		},
		ETagKey: "paper",
		Enabled: true,
	}
	middleware.MethodAndCache(http.MethodGet, cacheOpts)(paperHandler)(w, r)
}
