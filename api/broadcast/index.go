package handler

import (
	"crypto/subtle"
	"main/lib/broadcast"
	"main/lib/logger"
	"main/lib/middleware"
	"net/http"
	"os"
)

// broadcastHandler contains the main logic for the broadcast endpoint
func broadcastHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	logger.Info("Broadcast request initiated", ctx)

	// 1. Validate environment configuration
	expectedSecret := os.Getenv("DAILY_BROADCAST_KEY")
	if expectedSecret == "" {
		logger.Error("Broadcast key not configured", nil, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Server configuration error")
		return
	}

	// 2. Check secret key authentication
	incomingSecret := r.Header.Get("secret")
	ctx["has_secret_header"] = incomingSecret != ""
	ctx["secret_length"] = len(incomingSecret)

	if incomingSecret == "" {
		logger.Warn("Missing authentication secret", ctx)
		middleware.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Use constant-time comparison to prevent timing attacks
	logger.Debug("Validating authentication secret", ctx)
	if subtle.ConstantTimeCompare([]byte(incomingSecret), []byte(expectedSecret)) != 1 {
		ctx["secret_valid"] = false
		logger.Warn("Invalid authentication secret provided", ctx)
		middleware.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx["secret_valid"] = true
	logger.Info("Authentication successful", ctx)

	// 3. Trigger the broadcast
	logger.Info("Starting broadcast process", ctx)
	if err := broadcast.SendDailyBroadcast(); err != nil {
		logger.Error("Broadcast process failed", err, ctx)
		middleware.WriteJSONError(w, http.StatusInternalServerError, "Server error")
		return
	}

	logger.Info("Broadcast completed successfully", ctx)

	// 4. Return success
	middleware.WriteJSONResponse(w, http.StatusOK, middleware.SuccessResponse{Success: true})
}

// Handler is the Vercel serverless function entrypoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Broadcast endpoints should not be cached
	middleware.NoCache(http.MethodPost)(broadcastHandler)(w, r)
}
