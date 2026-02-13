package paper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"main/lib/logger"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"
)

const (
	// Default SageMaker endpoint name (can be overridden via env var)
	defaultEndpointName = "ds1-serverless-1762961395"
	// Default region
	defaultRegion = "us-east-1"
	// Request timeout
	embeddingTimeout = 30 * time.Second
	// Maximum batch size for embedding requests
	maxBatchSize = 32
	// Maximum concurrent embedding requests
	maxConcurrency = 15
	// Stagger delay between batches to allow SageMaker warm-up (milliseconds)
	batchStaggerDelay = 100
)

// EmbeddingService handles embedding generation via SageMaker
type EmbeddingService struct {
	client       *sagemakerruntime.Client
	endpointName string
	region       string
	mu           sync.RWMutex
	// Global semaphore to limit total concurrent requests (respects endpoint max concurrency)
	semaphore    chan struct{}
	// Cache for embeddings (text hash -> embedding)
	cache        map[string][]float32
}

// EmbedRequest represents the TEI /embed endpoint request format
type EmbedRequest struct {
	Inputs              interface{} `json:"inputs"` // string or []string
	Normalize           *bool       `json:"normalize,omitempty"`
	Truncate            *bool       `json:"truncate,omitempty"`
	TruncationDirection *string     `json:"truncation_direction,omitempty"` // "Left" or "Right"
	Dimensions          *int        `json:"dimensions,omitempty"`
	PromptName          *string     `json:"prompt_name,omitempty"`
}

// EmbedResponse represents the TEI /embed endpoint response format
// Response is [[float32, ...]] - array of arrays
type EmbedResponse [][]float32

var globalEmbeddingService *EmbeddingService
var embeddingServiceOnce sync.Once

// GetEmbeddingService returns the global embedding service instance
func GetEmbeddingService() (*EmbeddingService, error) {
	var initErr error
	embeddingServiceOnce.Do(func() {
		globalEmbeddingService, initErr = NewEmbeddingService()
	})
	return globalEmbeddingService, initErr
}

// NewEmbeddingService creates a new embedding service instance
func NewEmbeddingService() (*EmbeddingService, error) {
	endpointName := os.Getenv("SAGEMAKER_ENDPOINT_NAME")
	if endpointName == "" {
		endpointName = defaultEndpointName
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configure HTTP client for low latency with connection pooling
	httpClient := awshttp.NewBuildableClient().
		WithDialerOptions(func(d *net.Dialer) {
			d.KeepAlive = 30 * time.Second // Enable keep-alive for connection reuse
			d.Timeout = 5 * time.Second     // Connection timeout
		}).
		WithTransportOptions(func(tr *http.Transport) {
			tr.MaxIdleConns = 100                    // Max idle connections across all hosts
			tr.MaxIdleConnsPerHost = 10              // Max idle connections per host (SageMaker endpoint)
			tr.IdleConnTimeout = 90 * time.Second    // Keep idle connections alive
			tr.TLSHandshakeTimeout = 5 * time.Second // TLS handshake timeout
		}).
		WithTimeout(embeddingTimeout + 5*time.Second) // Request timeout (slightly longer than embedding timeout)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		logger.Error("Failed to load AWS config", err, map[string]interface{}{
			"endpoint":   endpointName,
			"region":     region,
			"error_type": "AWS_CONFIG_ERROR",
			"suggestion": "Check AWS credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY) or ~/.aws/credentials",
		})
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sagemakerruntime.NewFromConfig(cfg)

	// Generate instance ID for tracking (first 8 chars of hash of timestamp + random)
	instanceID := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))[:8]
	
	logger.Info("SageMaker embedding service initialized", map[string]interface{}{
		"endpoint":       endpointName,
		"region":         region,
		"max_concurrency": maxConcurrency,
		"instance_id":    instanceID,
	})

	svc := &EmbeddingService{
		client:       client,
		endpointName: endpointName,
		region:       region,
		semaphore:    make(chan struct{}, maxConcurrency),
		cache:        make(map[string][]float32),
	}
	
	// Store instance ID in a way we can log it later
	_ = instanceID
	
	return svc, nil
}

// GetClient returns the SageMaker client and endpoint name for direct access
func (e *EmbeddingService) GetClient() (*sagemakerruntime.Client, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.client, e.endpointName
}

// hashText creates a SHA256 hash of the text for cache key
func (e *EmbeddingService) hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// GenerateEmbedding generates an embedding for a single text
// Uses global semaphore to respect endpoint max concurrency
func (e *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	hash := e.hashText(text)
	
	e.mu.RLock()
	if cached, exists := e.cache[hash]; exists {
		e.mu.RUnlock()
		return cached, nil
	}
	e.mu.RUnlock()
	
	e.mu.RLock()
	semaphore := e.semaphore
	e.mu.RUnlock()
	
	semaphore <- struct{}{} // Acquire semaphore
	defer func() { <-semaphore }() // Release semaphore
	
	embeddings, err := e.generateEmbeddingsBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	
	e.mu.Lock()
	e.cache[hash] = embeddings[0]
	e.mu.Unlock()
	
	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts (batch)
// Automatically batches requests into chunks of maxBatchSize (32)
func (e *EmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Check cache concurrently for large batches
	cached := make([][]float32, len(texts))
	uncachedIndices := make([]int, 0, len(texts))
	uncachedTexts := make([]string, 0, len(texts))
	
	if len(texts) > 100 {
		// Parallel cache lookup for large batches
		type cacheResult struct {
			index int
			emb   []float32
			found bool
		}
		
		results := make(chan cacheResult, len(texts))
		var wg sync.WaitGroup
		
		for i, text := range texts {
			wg.Add(1)
			go func(idx int, txt string) {
				defer wg.Done()
				hash := e.hashText(txt)
				e.mu.RLock()
				emb, exists := e.cache[hash]
				e.mu.RUnlock()
				results <- cacheResult{index: idx, emb: emb, found: exists}
			}(i, text)
		}
		
		go func() {
			wg.Wait()
			close(results)
		}()
		
		for res := range results {
			if res.found {
				cached[res.index] = res.emb
			} else {
				uncachedIndices = append(uncachedIndices, res.index)
				uncachedTexts = append(uncachedTexts, texts[res.index])
			}
		}
	} else {
		// Sequential lookup for small batches (faster due to no goroutine overhead)
		e.mu.RLock()
		for i, text := range texts {
			hash := e.hashText(text)
			if emb, exists := e.cache[hash]; exists {
				cached[i] = emb
			} else {
				uncachedIndices = append(uncachedIndices, i)
				uncachedTexts = append(uncachedTexts, text)
			}
		}
		e.mu.RUnlock()
	}
	
	if len(uncachedTexts) == 0 {
		return cached, nil
	}
	
	// Generate embeddings for uncached texts only (already filtered from cache)
	var uncachedEmbeddings [][]float32
	var err error
	if len(uncachedTexts) <= maxBatchSize {
		// Single batch - no need to check cache again, we already filtered
		uncachedEmbeddings, err = e.generateEmbeddingsBatchUncached(ctx, uncachedTexts)
	} else {
		numBatches := (len(uncachedTexts) + maxBatchSize - 1) / maxBatchSize

		batches := make([][]string, 0, numBatches)
		for i := 0; i < len(uncachedTexts); i += maxBatchSize {
			end := i + maxBatchSize
			if end > len(uncachedTexts) {
				end = len(uncachedTexts)
			}
			batches = append(batches, uncachedTexts[i:end])
		}

		type batchResult struct {
			index     int
			embeddings [][]float32
			err       error
		}

		e.mu.RLock()
		semaphore := e.semaphore
		e.mu.RUnlock()
		
		results := make(chan batchResult, numBatches)
		
		for i, batch := range batches {
			semaphore <- struct{}{}
			go func(batchIndex int, batchTexts []string) {
				defer func() { <-semaphore }()
				// Stagger batches slightly to allow SageMaker warm-up
				// First batch starts immediately, subsequent batches wait progressively
				if batchIndex > 0 {
					staggerDelay := time.Duration(batchIndex*batchStaggerDelay) * time.Millisecond
					select {
					case <-ctx.Done():
						results <- batchResult{
							index:      batchIndex,
							embeddings: nil,
							err:        ctx.Err(),
						}
						return
					case <-time.After(staggerDelay):
						// Continue after stagger delay
					}
				}
				// No cache check needed - already filtered in GenerateEmbeddings
				embeddings, err := e.generateEmbeddingsBatchUncached(ctx, batchTexts)
				results <- batchResult{
					index:      batchIndex,
					embeddings: embeddings,
					err:        err,
				}
			}(i, batch)
		}

		batchResults := make([]batchResult, numBatches)
		for i := 0; i < numBatches; i++ {
			result := <-results
			batchResults[result.index] = result
		}

		allEmbeddings := make([][]float32, 0, len(uncachedTexts))
		for _, result := range batchResults {
			if result.err != nil {
				return nil, fmt.Errorf("batch %d failed: %w", result.index+1, result.err)
			}
			allEmbeddings = append(allEmbeddings, result.embeddings...)
		}
		uncachedEmbeddings = allEmbeddings
	}
	if err != nil {
		return nil, err
	}
	
	// Merge cached and uncached results (generateEmbeddingsBatch already cached uncached ones)
	result := make([][]float32, len(texts))
	uncachedIdx := 0
	for i := range texts {
		if cached[i] != nil {
			result[i] = cached[i]
		} else {
			result[i] = uncachedEmbeddings[uncachedIdx]
			uncachedIdx++
		}
	}
	
	return result, nil
}

// generateEmbeddingsBatch generates embeddings for a single batch (max maxBatchSize)
// Checks cache and handles both cached and uncached texts
func (e *EmbeddingService) generateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) > maxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum %d", len(texts), maxBatchSize)
	}

	// Check cache concurrently for large batches
	cached := make([][]float32, len(texts))
	uncachedIndices := make([]int, 0, len(texts))
	uncachedTexts := make([]string, 0, len(texts))
	
	if len(texts) > 10 {
		// Parallel cache lookup for larger batches
		type cacheResult struct {
			index int
			emb   []float32
			found bool
		}
		
		results := make(chan cacheResult, len(texts))
		var wg sync.WaitGroup
		
		for i, text := range texts {
			wg.Add(1)
			go func(idx int, txt string) {
				defer wg.Done()
				hash := e.hashText(txt)
				e.mu.RLock()
				emb, exists := e.cache[hash]
				e.mu.RUnlock()
				results <- cacheResult{index: idx, emb: emb, found: exists}
			}(i, text)
		}
		
		go func() {
			wg.Wait()
			close(results)
		}()
		
		for res := range results {
			if res.found {
				cached[res.index] = res.emb
			} else {
				uncachedIndices = append(uncachedIndices, res.index)
				uncachedTexts = append(uncachedTexts, texts[res.index])
			}
		}
	} else {
		// Sequential lookup for small batches
		e.mu.RLock()
		for i, text := range texts {
			hash := e.hashText(text)
			if emb, exists := e.cache[hash]; exists {
				cached[i] = emb
			} else {
				uncachedIndices = append(uncachedIndices, i)
				uncachedTexts = append(uncachedTexts, text)
			}
		}
		e.mu.RUnlock()
	}
	
	// If all cached, return immediately
	if len(uncachedTexts) == 0 {
		logger.Debug("All batch embeddings from cache", map[string]interface{}{
			"total": len(texts),
		})
		return cached, nil
	}
	

	// Generate embeddings for uncached texts only
	uncachedEmbeddings, err := e.generateEmbeddingsBatchUncached(ctx, uncachedTexts)
	if err != nil {
		return nil, err
	}
	
	// Merge cached and uncached results
	result := make([][]float32, len(texts))
	uncachedIdx := 0
	for i := range texts {
		if cached[i] != nil {
			result[i] = cached[i]
		} else {
			result[i] = uncachedEmbeddings[uncachedIdx]
			uncachedIdx++
		}
	}
	
	return result, nil
}

// generateEmbeddingsBatchUncached generates embeddings for texts that are known to be uncached
// No cache check performed - assumes all texts need embedding generation
func (e *EmbeddingService) generateEmbeddingsBatchUncached(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) > maxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum %d", len(texts), maxBatchSize)
	}

	e.mu.RLock()
	endpointName := e.endpointName
	client := e.client
	e.mu.RUnlock()

	// Prepare request payload - match Python boto3 format exactly
	// {"inputs": ["text1", "text2", ...]}
	payload, err := json.Marshal(map[string]interface{}{
		"inputs": texts,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, embeddingTimeout)
	defer cancel()

	// Invoke SageMaker endpoint (matches Python boto3 invoke_endpoint)
	input := &sagemakerruntime.InvokeEndpointInput{
		EndpointName: aws.String(endpointName),
		ContentType:  aws.String("application/json"),
		Body:         payload,
	}

	resp, err := client.InvokeEndpoint(reqCtx, input)
	if err != nil {
		// Check for specific AWS error types
		errorDetails := map[string]interface{}{
			"endpoint":     endpointName,
			"texts":        len(texts),
			"payload_size": len(payload),
			"error":        err.Error(),
		}
		
		// Check for specific error types
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "credential") || strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "access denied") {
			errorDetails["error_type"] = "AWS_CREDENTIALS_ERROR"
			errorDetails["suggestion"] = "Check AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables"
		} else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "context deadline") {
			errorDetails["error_type"] = "TIMEOUT_ERROR"
		} else if strings.Contains(errMsg, "endpoint") || strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "notfound") {
			errorDetails["error_type"] = "ENDPOINT_ERROR"
			errorDetails["suggestion"] = "Verify endpoint name and region are correct"
		} else if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") {
			errorDetails["error_type"] = "NETWORK_ERROR"
		} else {
			errorDetails["error_type"] = "UNKNOWN_ERROR"
		}
		
		logger.Error("SageMaker invocation failed", err, errorDetails)
		return nil, fmt.Errorf("sagemaker invocation failed: %w", err)
	}

	body := resp.Body

	// TEI /embed endpoint returns [[float32, ...]] - array of arrays
	// Try parsing as array of arrays of float32 first
	var float32Resp [][]float32
	if err := json.Unmarshal(body, &float32Resp); err == nil && len(float32Resp) > 0 {
		if len(float32Resp) != len(texts) {
			return nil, fmt.Errorf("embedding count mismatch: expected %d, got %d", len(texts), len(float32Resp))
		}
		
		e.mu.Lock()
		for i, text := range texts {
			hash := e.hashText(text)
			e.cache[hash] = float32Resp[i]
		}
		e.mu.Unlock()
		
		return float32Resp, nil
	}

	// Try parsing as array of arrays of float64 (convert to float32)
	var float64Resp [][]float64
	if err := json.Unmarshal(body, &float64Resp); err == nil && len(float64Resp) > 0 {
		if len(float64Resp) != len(texts) {
			return nil, fmt.Errorf("embedding count mismatch: expected %d, got %d", len(texts), len(float64Resp))
		}
		float32Result := make([][]float32, len(float64Resp))
		for i, vec := range float64Resp {
			float32Result[i] = make([]float32, len(vec))
			for j, val := range vec {
				float32Result[i][j] = float32(val)
			}
		}
		
		e.mu.Lock()
		for i, text := range texts {
			hash := e.hashText(text)
			e.cache[hash] = float32Result[i]
		}
		e.mu.Unlock()
		
		return float32Result, nil
	}

	// Check for error response from TEI
	var errorResp struct {
		Error     string `json:"error"`
		ErrorType string `json:"error_type"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
		logger.Error("TEI embedding error", fmt.Errorf("%s: %s", errorResp.ErrorType, errorResp.Error), map[string]interface{}{
			"endpoint":   endpointName,
			"error_type": errorResp.ErrorType,
		})
		return nil, fmt.Errorf("TEI embedding error (%s): %s", errorResp.ErrorType, errorResp.Error)
	}

	logger.Warn("Failed to parse TEI embedding response", map[string]interface{}{
		"endpoint":     endpointName,
		"body_size":    len(body),
		"body_preview": string(body[:min(200, len(body))]),
		"body_full":    string(body),
	})

	return nil, fmt.Errorf("unable to parse embedding response: unexpected format")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

