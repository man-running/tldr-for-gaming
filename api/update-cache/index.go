package handler

import (
	"crypto/subtle"
	"main/lib/middleware"
	"main/lib/summary"
	"net/http"
	"os"
	"strings"
	"time"
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

// updateCacheHandler contains the main logic for the update cache endpoint
func updateCacheHandler(w http.ResponseWriter, r *http.Request) {
	// Check for secret key to prevent unauthorized updates
	secretKey := r.Header.Get("X-Update-Key")
	expectedKey := os.Getenv("UPDATE_KEY")

	if expectedKey == "" || subtle.ConstantTimeCompare([]byte(secretKey), []byte(expectedKey)) != 1 {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	service := summary.NewService()

	// Construct absolute URL using BASE_URL (use base URL for canonical cache content)
	requestURL := constructAbsoluteURL("api/tldr")
	// Update both papers and summary caches with fresh data
	err := service.UpdateCache(r.Context(), requestURL)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Error updating cache: "+err.Error())
		return
	}

	// Return success response
	response := map[string]interface{}{
		"status":    "Cache updated successfully",
		"message":   "Both papers and summary caches have been refreshed with fresh data",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	middleware.WriteJSONResponse(w, http.StatusOK, response)
}

// Handler is the Vercel serverless function entrypoint for the update cache API.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Configure caching for update endpoint (disabled - this is a maintenance endpoint)
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
	middleware.MethodAndCache(http.MethodPost, cacheOpts)(updateCacheHandler)(w, r)
}
