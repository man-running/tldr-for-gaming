package middleware

import (
	"encoding/json"
	"main/lib/response"
	"net/http"
)

// WriteJSONError is a helper for common JSON error responses
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	response.WriteJSONResponse(w, statusCode, ErrorResponse{Error: message})
}

// WriteJSONSuccess is a helper for common JSON success responses
func WriteJSONSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	response.WriteJSONResponse(w, statusCode, data)
}

// WriteJSONResponse is a middleware wrapper for response.WriteJSONResponse
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	response.WriteJSONResponse(w, statusCode, data)
}

// ParseJSONBody is a helper to parse JSON request bodies with proper error handling
func ParseJSONBody(r *http.Request, v interface{}) error {
	defer func() {
		_ = r.Body.Close() // Ignore error as per Go best practices for HTTP request bodies
	}()
	return json.NewDecoder(r.Body).Decode(v)
}
