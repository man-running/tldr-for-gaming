package handler

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"main/lib/logger"
	"main/lib/middleware"
	"main/lib/paper"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"
)

const proxyTimeout = 30 * time.Second

// mapErrorTypeToStatus maps TEI error_type to HTTP status code
func mapErrorTypeToStatus(errorType string) int {
	switch errorType {
	case "empty":
		return http.StatusBadRequest // 400
	case "validation":
		return http.StatusRequestEntityTooLarge // 413
	case "tokenizer":
		return http.StatusUnprocessableEntity // 422
	case "backend":
		return http.StatusFailedDependency // 424
	case "overloaded":
		return http.StatusTooManyRequests // 429
	default:
		return http.StatusBadGateway // 502
	}
}

// proxyHandler acts as a reverse proxy to the SageMaker embedding endpoint
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	// Get embedding service to access SageMaker client
	embeddingService, err := paper.GetEmbeddingService()
	if err != nil {
		logger.Error("Embedding service initialization failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, middleware.ErrorResponse{
			Error: "Failed to initialize embedding service",
		})
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("Failed to read request body", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "Failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	// Get SageMaker client and endpoint name
	client, endpointName := embeddingService.GetClient()

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(r.Context(), proxyTimeout)
	defer cancel()

	// Invoke SageMaker endpoint directly
	input := &sagemakerruntime.InvokeEndpointInput{
		EndpointName: aws.String(endpointName),
		ContentType:  aws.String("application/json"),
		Body:         body,
	}

	resp, err := client.InvokeEndpoint(reqCtx, input)
	if err != nil {
		// Check if error contains response body (SageMaker may return error responses)
		errMsg := err.Error()
		var errorResp struct {
			Error     string `json:"error"`
			ErrorType string `json:"error_type"`
		}
		
		// Try to extract JSON error from error message
		if jsonStart := strings.Index(errMsg, "{"); jsonStart != -1 {
			jsonStr := errMsg[jsonStart:]
			if jsonEnd := strings.LastIndex(jsonStr, "}"); jsonEnd != -1 {
				jsonStr = jsonStr[:jsonEnd+1]
				if err := json.Unmarshal([]byte(jsonStr), &errorResp); err == nil && errorResp.ErrorType != "" {
					statusCode := mapErrorTypeToStatus(errorResp.ErrorType)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(statusCode)
					// Use the extracted string directly to avoid re-encoding
					w.Write([]byte(jsonStr))
					return
				}
			}
		}
		
		// Fallback: return generic error
		logger.Error("SageMaker invocation failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusBadGateway, middleware.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Optimization: Check for success first without Unmarshal
	// TEI returns a JSON array "[[...]]" on success
	isSuccess := false
	for _, b := range resp.Body {
		switch b {
		case ' ', '\t', '\r', '\n':
			continue
		case '[':
			isSuccess = true
		}
		break
	}

	if isSuccess {
		// Success response - pass through directly
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(resp.Body); err != nil {
			logger.Error("Failed to write response", err, ctx)
		}
		return
	}

	// Check if response body is an error response (TEI returns errors in body even on 200)
	var errorResp struct {
		Error     string `json:"error"`
		ErrorType string `json:"error_type"`
	}
	if err := json.Unmarshal(resp.Body, &errorResp); err == nil && errorResp.ErrorType != "" {
		// It's an error response, map to appropriate status code
		statusCode := mapErrorTypeToStatus(errorResp.ErrorType)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write(resp.Body)
		return
	}

	// Fallback success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(resp.Body); err != nil {
		logger.Error("Failed to write response", err, ctx)
	}
}

// binaryHandler returns embeddings in binary format with header + raw float32 data
func binaryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Log.WithRequest(r)

	// Get embedding service to access SageMaker client
	embeddingService, err := paper.GetEmbeddingService()
	if err != nil {
		logger.Error("Embedding service initialization failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusInternalServerError, middleware.ErrorResponse{
			Error: "Failed to initialize embedding service",
		})
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("Failed to read request body", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "Failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	// Get SageMaker client and endpoint name
	client, endpointName := embeddingService.GetClient()

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(r.Context(), proxyTimeout)
	defer cancel()

	// Invoke SageMaker endpoint directly
	input := &sagemakerruntime.InvokeEndpointInput{
		EndpointName: aws.String(endpointName),
		ContentType:  aws.String("application/json"),
		Body:         body,
	}

	resp, err := client.InvokeEndpoint(reqCtx, input)
	if err != nil {
		// Check if error contains response body
		errMsg := err.Error()
		var errorResp struct {
			Error     string `json:"error"`
			ErrorType string `json:"error_type"`
		}
		
		if jsonStart := strings.Index(errMsg, "{"); jsonStart != -1 {
			jsonStr := errMsg[jsonStart:]
			if jsonEnd := strings.LastIndex(jsonStr, "}"); jsonEnd != -1 {
				jsonStr = jsonStr[:jsonEnd+1]
				if err := json.Unmarshal([]byte(jsonStr), &errorResp); err == nil && errorResp.ErrorType != "" {
					statusCode := mapErrorTypeToStatus(errorResp.ErrorType)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(statusCode)
					w.Write([]byte(jsonStr))
					return
				}
			}
		}
		
		logger.Error("SageMaker invocation failed", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusBadGateway, middleware.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Parse JSON response to extract embeddings
	var embeddings [][]float32
	if err := json.Unmarshal(resp.Body, &embeddings); err != nil {
		// Check if it's an error response
		var errorResp struct {
			Error     string `json:"error"`
			ErrorType string `json:"error_type"`
		}
		if err2 := json.Unmarshal(resp.Body, &errorResp); err2 == nil && errorResp.ErrorType != "" {
			statusCode := mapErrorTypeToStatus(errorResp.ErrorType)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			w.Write(resp.Body)
			return
		}
		
		logger.Error("Failed to parse embeddings response", err, ctx)
		middleware.WriteJSONResponse(w, http.StatusBadGateway, middleware.ErrorResponse{
			Error: "Failed to parse embeddings",
		})
		return
	}

	if len(embeddings) == 0 {
		logger.Warn("Empty embeddings array", ctx)
		middleware.WriteJSONResponse(w, http.StatusBadRequest, middleware.ErrorResponse{
			Error: "No embeddings returned",
		})
		return
	}

	// Validate all embeddings have same dimensions
	batchSize := len(embeddings)
	dims := len(embeddings[0])
	for _, emb := range embeddings {
		if len(emb) != dims {
			logger.Warn("Inconsistent embedding dimensions", ctx)
			middleware.WriteJSONResponse(w, http.StatusBadGateway, middleware.ErrorResponse{
				Error: "Inconsistent embedding dimensions",
			})
			return
		}
	}

	// Create binary response
	// Header format (16 bytes):
	// - Magic: 4 bytes "EMBD"
	// - Version: 1 byte (1)
	// - Batch: 2 bytes uint16
	// - Dims: 2 bytes uint16
	// - Dtype: 1 byte (0 = float32)
	// - Endian: 1 byte (0 = little-endian)
	// - Reserved: 5 bytes
	header := make([]byte, 16)
	copy(header[0:4], []byte("EMBD")) // Magic
	header[4] = 1                      // Version
	binary.LittleEndian.PutUint16(header[5:7], uint16(batchSize))
	binary.LittleEndian.PutUint16(header[7:9], uint16(dims))
	header[9] = 0  // Dtype: float32
	header[10] = 0 // Endian: little-endian
	// header[11:16] reserved (zeros)

	// Write float32 data (batch * dims * 4 bytes)
	dataSize := batchSize * dims * 4
	binaryData := make([]byte, 16+dataSize)
	copy(binaryData[0:16], header)

	// Write embeddings as little-endian float32
	offset := 16
	for _, emb := range embeddings {
		for _, val := range emb {
			binary.LittleEndian.PutUint32(binaryData[offset:offset+4], uint32(math.Float32bits(val)))
			offset += 4
		}
	}

	ctx["batch_size"] = batchSize
	ctx["dims"] = dims
	ctx["binary_size"] = len(binaryData)
	logger.Info("Binary embeddings response prepared", ctx)

	// Set headers and write binary response
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(binaryData)))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(binaryData); err != nil {
		logger.Error("Failed to write binary response", err, ctx)
	}
}

// Handler is the Vercel serverless function entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	// Check if binary format is requested via Accept header or query parameter
	acceptHeader := r.Header.Get("Accept")
	formatParam := r.URL.Query().Get("format")
	
	if formatParam == "binary" || strings.Contains(acceptHeader, "application/octet-stream") {
		middleware.MethodValidator(http.MethodPost)(binaryHandler)(w, r)
	} else {
		middleware.MethodValidator(http.MethodPost)(proxyHandler)(w, r)
	}
}
