package handler

import (
	"image"
	"main/lib/logger"
	"main/lib/middleware"
	"main/lib/og"
	"net/http"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/srwiley/oksvg"
)

// ogHandler contains the main logic for the OG image endpoint
func ogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	// 1. Parse arxivId from the query string, e.g., "/api/og?id=1706.03762"
	arxivId := r.URL.Query().Get("id")
	ctx["arxiv_id"] = arxivId

	logger.Info("Processing OG image request", ctx)

	if arxivId == "" {
		logger.Warn("Missing arxiv_id parameter in OG request", ctx)
		http.Error(w, "Missing arxiv_id parameter", http.StatusBadRequest)
		return
	}

	// 2. Load all assets concurrently
	logger.Debug("Loading OG image assets concurrently", ctx)
	var bg image.Image
	var logo *oksvg.SvgIcon
	var boldFont, blackFont *truetype.Font
	var title string
	var assetsErr, fontsErr, titleErr error
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		logger.Debug("Loading image and logo assets", ctx)
		bg, logo, assetsErr = og.LoadImageAndLogo()
	}()
	go func() {
		defer wg.Done()
		logger.Debug("Loading font assets", ctx)
		boldFont, blackFont, fontsErr = og.LoadFonts()
	}()
	go func() {
		defer wg.Done()
		logger.Debug("Fetching paper title", ctx)
		title, titleErr = og.GetTitle(arxivId)
	}()
	wg.Wait()

	// Check for errors after all concurrent operations are done
	if assetsErr != nil {
		logger.Error("Failed to load image assets", assetsErr, ctx)
		http.Error(w, "Internal server error: could not load assets", http.StatusInternalServerError)
		return
	}
	if fontsErr != nil {
		logger.Error("Failed to load fonts", fontsErr, ctx)
		http.Error(w, "Internal server error: could not load fonts", http.StatusInternalServerError)
		return
	}
	if titleErr != nil {
		logger.Error("Failed to get paper title", titleErr, ctx)
		http.Error(w, "Internal server error: could not get title", http.StatusInternalServerError)
		return
	}

	ctx["title_length"] = len(title)
	logger.Info("All assets loaded successfully", ctx)

	// 3. Render the image
	logger.Debug("Rendering OG image", ctx)
	dc, err := og.RenderImage(title, bg, logo, boldFont, blackFont)
	if err != nil {
		logger.Error("Failed to render OG image", err, ctx)
		http.Error(w, "Internal server error: could not render image", http.StatusInternalServerError)
		return
	}

	// 4. Set Content-Type and write response (caching handled by middleware)
	logger.Debug("Encoding and sending OG image", ctx)
	w.Header().Set("Content-Type", "image/png")

	if err := dc.EncodePNG(w); err != nil {
		logger.Error("Failed to encode PNG image", err, ctx)
		// Header might already be written, so we can't send a new http.Error
		return
	}

	logger.Info("OG image generated and sent successfully", ctx)
}

// Handler is the Vercel serverless function entrypoint for the OG Image API.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Configure caching for OG image endpoint (long-term caching for static images)
	cacheOpts := middleware.CacheOptions{
		Config: middleware.CacheConfig{
			MaxAge:               31536000, // 1 year browser cache
			SMaxAge:              31536000, // 1 year CDN cache
			StaleWhileRevalidate: 604800,    // 1 week stale-while-revalidate
			StaleIfError:         0,        // No stale-if-error for images
		},
		ETagKey: "og-image",
		Enabled: true,
	}
	middleware.MethodAndCache(http.MethodGet, cacheOpts)(ogHandler)(w, r)
}
