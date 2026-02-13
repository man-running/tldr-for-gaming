package handler

import (
	"io"
	"main/lib/feed"
	"main/lib/logger"
	"main/lib/middleware"
	"main/lib/summary"
	"net/http"
	"os"
	"strings"
)

// constructAbsoluteURL constructs an absolute URL using BASE_URL
func constructAbsoluteURL(path string) string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://tldr.takara.ai"
	}
	// Remove leading slash from path if present
	path = strings.TrimPrefix(path, "/")
	return baseURL + "/" + path
}

// tldrHandler contains the main logic for the TLDR endpoint (LLM summary by default)
func tldrHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	// Determine requested format from query parameter
	// Default to RSS for RSS reader compatibility
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "rss" // Default to RSS for RSS readers
	}

	// Determine content type: "feed" for raw papers, "summary" for AI-generated
	contentType := r.URL.Query().Get("type")
	if contentType == "" {
		contentType = "summary" // Default to summary (LLM-generated)
	}

	ctx["format"] = format
	ctx["content_type"] = contentType

	logger.Info("Processing TLDR request", ctx)

	switch contentType {
	case "summary":
		logger.Debug("Serving AI-generated summary", ctx)
		// Serve AI-generated summary
		handleSummary(w, r, format, ctx)
	case "feed":
		logger.Debug("Serving raw feed data", ctx)
		// Serve raw feed data
		handleFeed(w, r, format, ctx)
	default:
		logger.Warn("Invalid content type requested", ctx)
		middleware.WriteJSONError(w, http.StatusBadRequest, "Invalid type parameter. Use 'feed' or 'summary'")
	}
}

// handleFeed serves the raw feed data in the requested format
func handleFeed(w http.ResponseWriter, r *http.Request, format string, ctx map[string]interface{}) {
	service := summary.NewService()

	// Construct absolute URL using BASE_URL
	requestURL := constructAbsoluteURL(strings.TrimPrefix(r.URL.Path, "/"))
	ctx["request_url"] = requestURL
	logger.Debug("Fetching raw papers feed", ctx)
	result, err := service.GetPapersRaw(r.Context(), requestURL)
	if err != nil {
		logger.Error("Failed to fetch raw papers feed", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	switch format {
	case "rss":
		logger.Debug("Serving RSS feed format", ctx)
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(result.Data); err != nil {
			logger.Error("Failed to write RSS response", err, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		}
	case "json":
		logger.Debug("Serving JSON feed format", ctx)
		// For JSON format, we need to get the feed data and marshal it
		result, err := feed.GetFeedRaw()
		if err != nil {
			logger.Error("Failed to fetch feed for JSON format", err, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		middleware.WriteJSONResponse(w, http.StatusOK, result.Data)
	default:
		logger.Warn("Invalid format requested for feed", ctx)
		middleware.WriteJSONError(w, http.StatusBadRequest, "Invalid format. Use 'json' or 'rss'. Default is RSS for RSS reader compatibility")
	}
}

// handleSummary serves the AI-generated summary in the requested format
func handleSummary(w http.ResponseWriter, r *http.Request, format string, ctx map[string]interface{}) {
	service := summary.NewService()

	// Construct absolute URL using BASE_URL
	requestURL := constructAbsoluteURL(strings.TrimPrefix(r.URL.Path, "/"))
	ctx["request_url"] = requestURL
	logger.Debug("Fetching AI-generated summary", ctx)
	result, err := service.GetSummaryRaw(r.Context(), requestURL)
	if err != nil {
		logger.Error("Failed to fetch AI summary", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// If blob URL is available but data is nil, fetch from blob URL
	var rssData []byte
	if result.BlobURL != nil && *result.BlobURL != "" && result.Data == nil {
		logger.Debug("Fetching summary from blob URL", map[string]interface{}{"blob_url": *result.BlobURL})
		resp, err := http.Get(*result.BlobURL)
		if err != nil {
			logger.Error("Failed to fetch from blob URL", err, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.Error("Blob URL returned non-200", nil, map[string]interface{}{
				"status": resp.StatusCode,
				"url":    *result.BlobURL,
			})
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		var readErr error
		rssData, readErr = io.ReadAll(resp.Body)
		if readErr != nil {
			logger.Error("Failed to read blob content", readErr, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	} else {
		rssData = result.Data
	}

	switch format {
	case "rss":
		logger.Debug("Serving RSS summary format", ctx)
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(rssData); err != nil {
			logger.Error("Failed to write RSS summary response", err, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		}
	case "json":
		logger.Debug("Serving JSON summary format", ctx)
		// For JSON format with summary, we'd need to parse the RSS and return JSON
		// For now, just serve the RSS as-is
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(rssData); err != nil {
			logger.Error("Failed to write JSON summary response", err, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		}
	default:
		logger.Warn("Invalid format requested for summary", ctx)
		middleware.WriteJSONError(w, http.StatusBadRequest, "Invalid format. Use 'json' or 'rss'. Default is RSS for RSS reader compatibility")
	}
}

// Handler is the Vercel serverless function entrypoint for the TLDR API (LLM summary).
func Handler(w http.ResponseWriter, r *http.Request) {
	// Configure caching for TLDR endpoint (disabled for daily feed)
	cacheOpts := middleware.CacheOptions{
		Config: middleware.CacheConfig{
			MaxAge:               0, // No browser caching
			SMaxAge:              0, // No CDN caching
			StaleWhileRevalidate: 0, // No stale-while-revalidate
			StaleIfError:         0, // No stale-if-error
		},
		ETagKey: "", // Disable ETags
		Enabled: true,
	}
	middleware.MethodAndCache(http.MethodGet, cacheOpts)(tldrHandler)(w, r)
}
