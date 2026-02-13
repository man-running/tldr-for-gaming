package subscribe

import (
	"encoding/json"
	"fmt"
	"main/lib/logger"
	"net/http"
	"net/url"
	"os"
	"time"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// VerifyTurnstileToken sends the user's token to Cloudflare for server-side validation.
func VerifyTurnstileToken(token string) (bool, error) {
	secretKey := os.Getenv("TURNSTILE_SECRET_KEY")
	if secretKey == "" {
		return false, fmt.Errorf("TURNSTILE_SECRET_KEY is not set")
	}

	// Create the form data payload.
	formData := url.Values{}
	formData.Set("secret", secretKey)
	formData.Set("response", token)

	// Make the POST request to Cloudflare.
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(turnstileVerifyURL, formData)
	if err != nil {
		return false, fmt.Errorf("failed to send verification request to Cloudflare: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Decode the JSON response.
	var result TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode Cloudflare response: %w", err)
	}

	// Log failures for debugging.
	if !result.Success {
		ctx := map[string]interface{}{
			"error_codes": result.ErrorCodes,
		}
		logger.Error("Turnstile verification failed", nil, ctx)
	}

	return result.Success, nil
}
