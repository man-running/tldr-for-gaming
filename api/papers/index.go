package handler

import (
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

// papersHandler contains the main logic for the papers endpoint (raw scraped feed)
func papersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)
	logger.Info("papersHandler called", ctx)

	service := summary.NewService()

	// Construct absolute URL using BASE_URL
	requestURL := constructAbsoluteURL(strings.TrimPrefix(r.URL.Path, "/"))

	logger.Info("Starting GetPapersRaw", map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"requestURL": requestURL,
	})

	result, err := service.GetPapersRaw(r.Context(), requestURL)
	if err != nil {
		logger.Error("GetPapersRaw failed", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	logger.Info("papersHandler success", map[string]interface{}{
		"data_size": len(result.Data),
		"source":    result.Source,
		"has_blob_url": result.BlobURL != nil,
	})

	// If blob URL is available, redirect client to fetch directly
	if result.BlobURL != nil && *result.BlobURL != "" {
		w.Header().Set("Location", *result.BlobURL)
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	// Otherwise, return the data directly
	w.Header().Set("Content-Type", "application/rss+xml")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(result.Data); err != nil {
		logger.Error("Failed to write response", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// Handler is the Vercel serverless function entrypoint for the papers API.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Configure caching for papers endpoint (disabled for daily feed)
	cacheOpts := middleware.CacheOptions{
		Config: middleware.CacheConfig{
			MaxAge:               300, // No browser caching
			SMaxAge:              300, // No CDN caching
			StaleWhileRevalidate: 0,   // No stale-while-revalidate
			StaleIfError:         0,   // No stale-if-error
		},
		ETagKey: "", // Disable ETags
		Enabled: true,
	}
	middleware.MethodAndCache(http.MethodGet, cacheOpts)(papersHandler)(w, r)
}
