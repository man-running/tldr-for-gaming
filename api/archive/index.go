package handler

import (
	"main/lib/logger"
	"main/lib/middleware"
	"main/lib/tldr"
	"net/http"
)

// archiveHandler contains the main logic for the archive endpoint
func archiveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	date := r.URL.Query().Get("date")
	ctx["requested_date"] = date
	ctx["has_date_param"] = date != ""

	if date != "" {
		logger.Info("Processing archive request for specific date (fallback)", ctx)

		// This endpoint is now primarily a fallback - clients should fetch directly from blob storage
		// Pattern: https://l0m9dfhwc2c0qq2u.public.blob.vercel-storage.com/tldr-feeds/{date}.json
		logger.Debug("Fetching TLDR feed content for date", ctx)
		feed, err := tldr.GetTldrFeed(date)
		if err != nil {
			logger.Error("Failed to fetch archive feed", err, ctx)
			middleware.WriteJSONResponse(w, http.StatusInternalServerError, tldr.ErrorResponse{Error: "Internal server error"})
			return
		}

		if feed == nil {
			logger.Warn("No feed found for requested date", ctx)
			middleware.WriteJSONResponse(w, http.StatusNotFound, tldr.ErrorResponse{Error: "Feed not found for this date"})
			return
		}

		ctx["feed_date"] = feed.LastBuildDate
		logger.Info("Archive feed retrieved successfully", ctx)
		// Write response (caching handled by middleware)
		middleware.WriteJSONResponse(w, http.StatusOK, feed)
	} else {
		logger.Info("Processing archive request for date list", ctx)

		// Handle request to list all available dates
		logger.Debug("Fetching available archive dates", ctx)
		dates, err := tldr.ListTldrFeedDates()
		if err != nil {
			logger.Error("Failed to fetch archive dates", err, ctx)
			middleware.WriteJSONResponse(w, http.StatusInternalServerError, tldr.ErrorResponse{Error: "Internal server error"})
			return
		}

		ctx["total_dates"] = len(dates)
		logger.Info("Archive dates retrieved successfully", ctx)
		// Write response (caching handled by middleware)
		middleware.WriteJSONResponse(w, http.StatusOK, tldr.DatesResponse{Dates: dates})
	}
}

// Handler is the Vercel serverless function entrypoint for the Archive API.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Configure caching for Archive endpoint
	cacheOpts := middleware.CacheOptions{
		Config: middleware.CacheConfig{
			MaxAge:               0,    // No browser caching
			SMaxAge:              300,  // 5 minutes CDN cache
			StaleWhileRevalidate: 3600, // 1 hour stale-while-revalidate
			StaleIfError:         0,    // No stale-if-error
		},
		ETagKey: "archive",
		Enabled: true,
	}
	middleware.MethodAndCache(http.MethodGet, cacheOpts)(archiveHandler)(w, r)
}
