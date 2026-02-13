package handler

import (
	"bytes"
	"encoding/json"
	"main/lib/logger"
	"main/lib/media"
	"main/lib/middleware"
	"main/lib/paper"
	"main/lib/rendering"
	"net/http"
)

type spectrogramResponse struct {
	URL string `json:"url"`
}

// spectrogramHandler generates a spectrogram image from a title
func spectrogramHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	query := r.URL.Query().Get("q")
	if query == "" {
		logger.Warn("Missing query parameter", ctx)
		middleware.WriteJSONError(w, http.StatusBadRequest, "Missing query parameter")
		return
	}

	format := r.URL.Query().Get("format")
	ctx["query"] = query
	ctx["format"] = format
	ctx["query_length"] = len(query)

	logger.Info("Processing spectrogram request", ctx)

	// Get embedding service
	logger.Debug("Initializing embedding service", ctx)
	embeddingService, err := paper.GetEmbeddingService()
	if err != nil {
		logger.Error("Embedding service initialization failed", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to initialize embedding service")
		return
	}

	// Generate embedding from title
	logger.Debug("Generating embedding from title", ctx)
	embedding, err := embeddingService.GenerateEmbedding(r.Context(), query)
	if err != nil {
		logger.Error("Failed to generate embedding", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to generate embedding")
		return
	}
	ctx["embedding_dim"] = len(embedding)
	logger.Debug("Embedding generated successfully", ctx)

	// Convert float32 to float64 for spectrogram
	vector := make([]float64, len(embedding))
	for i, v := range embedding {
		vector[i] = float64(v)
	}

	// Generate spectrogram image to buffer
	logger.Debug("Generating spectrogram image", ctx)
	var imageBuf bytes.Buffer
	if err := rendering.GenerateSpectrogramImage(vector, 2048, 512, 32, 1.0, &imageBuf); err != nil {
		logger.Error("Failed to generate spectrogram", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to generate spectrogram")
		return
	}
	imageSize := imageBuf.Len()
	ctx["image_size_bytes"] = imageSize
	logger.Debug("Spectrogram image generated successfully", ctx)

	// If format=json, store in blob and return JSON
	if format == "json" {
		blobPath := media.GenerateSpectrogramBlobPath(query)
		ctx["blob_path"] = blobPath
		logger.Debug("Storing spectrogram in blob storage", ctx)

		blobURL, err := media.StoreImageBlob(blobPath, imageBuf.Bytes(), "image/webp")
		if err != nil {
			logger.Error("Failed to store spectrogram in blob", err, ctx)
			middleware.WriteJSONError(w, http.StatusInternalServerError, "Failed to store image")
			return
		}
		ctx["blob_url"] = blobURL
		logger.Info("Spectrogram stored in blob successfully", ctx)

		w.Header().Set("Content-Type", "application/json")
		response := spectrogramResponse{URL: blobURL}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Failed to encode JSON response", err, ctx)
			return
		}
		logger.Info("Spectrogram request completed successfully (JSON format)", ctx)
		return
	}

	// Otherwise, return image directly (backward compatibility)
	logger.Debug("Returning image directly (backward compatibility)", ctx)
	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("X-Image-Alt", query)
	if _, err := w.Write(imageBuf.Bytes()); err != nil {
		logger.Error("Failed to write image response", err, ctx)
		return
	}
	logger.Info("Spectrogram request completed successfully (image format)", ctx)
}

// Handler is the Vercel serverless function entrypoint for the Spectrogram API.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Spectrogram images can be cached aggressively (1 year)
	cacheOpts := middleware.CacheOptions{
		Config: middleware.CacheConfig{
			MaxAge:               0,        // No browser caching
			SMaxAge:              31536000, // 1 year CDN cache
			StaleWhileRevalidate: 604800,   // 7 days stale-while-revalidate
			StaleIfError:         86400,    // 1 day stale-if-error
		},
		ETagKey: "spectrogram",
		Enabled: true,
	}
	middleware.MethodAndCache(http.MethodGet, cacheOpts)(spectrogramHandler)(w, r)
}

