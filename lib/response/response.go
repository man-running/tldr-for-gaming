package response

import (
	"encoding/json"
	"net/http"
)

// WriteJSONResponse is a helper to marshal data and write a JSON response with the specified status code.
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Fallback if JSON encoding fails
		http.Error(w, `{"error":"Failed to encode JSON response"}`, http.StatusInternalServerError)
	}
}
