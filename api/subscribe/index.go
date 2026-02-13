package handler

import (
	"main/lib/logger"
	"main/lib/middleware"
	"main/lib/subscribe"
	"net/http"
)

// subscribeHandler contains the main logic for the subscribe endpoint
func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	logger.Info("Subscription request started", ctx)

	// 1. Decode the request body
	var reqBody subscribe.RequestBody
	if err := middleware.ParseJSONBody(r, &reqBody); err != nil {
		ctx["parse_error"] = err.Error()
		logger.Warn("Failed to parse request body", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, subscribe.ApiResponse{Error: "Invalid request body"})
		return
	}

	ctx["email"] = reqBody.Email
	ctx["has_turnstile_token"] = reqBody.TurnstileToken != ""
	ctx["turnstile_token_length"] = len(reqBody.TurnstileToken)

	logger.Debug("Request body parsed successfully", ctx)

	// 2. Validate input
	if reqBody.Email == "" {
		logger.Warn("Missing email in subscription request", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, subscribe.ApiResponse{Error: "Email required."})
		return
	}

	if reqBody.TurnstileToken == "" {
		logger.Warn("Missing Turnstile token in subscription request", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, subscribe.ApiResponse{Error: "Turnstile token required."})
		return
	}

	// 3. Verify Turnstile token
	logger.Debug("Verifying Turnstile token", ctx)
	isVerified, err := subscribe.VerifyTurnstileToken(reqBody.TurnstileToken)
	if err != nil {
		logger.Error("Turnstile verification failed with error", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, subscribe.ApiResponse{Error: "Server error"})
		return
	}
	if !isVerified {
		logger.Warn("Turnstile verification failed - invalid token", ctx)
		middleware.WriteJSONResponse(w, http.StatusForbidden, subscribe.ApiResponse{Error: "Verification failed"})
		return
	}

	logger.Info("Turnstile verification successful", ctx)

	// 4. Subscribe the email
	logger.Debug("Processing email subscription", ctx)
	if err := subscribe.SubscribeEmail(reqBody.Email); err != nil {
		logger.Error("Email subscription failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, subscribe.ApiResponse{Error: "Server error"})
		return
	}

	logger.Info("Email subscription completed successfully", ctx)

	// 5. Return success
	middleware.WriteJSONResponse(w, http.StatusOK, subscribe.ApiResponse{Success: true})
}

// Handler is the Vercel serverless function entrypoint for the subscribe API.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Subscription endpoints should not be cached
	middleware.NoCache(http.MethodPost)(subscribeHandler)(w, r)
}
