package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// CacheConfig defines configurable caching behavior
type CacheConfig struct {
	MaxAge               int // Browser cache max-age (typically 0)
	SMaxAge              int // CDN cache max-age
	StaleWhileRevalidate int // Seconds to serve stale content while revalidating
	StaleIfError         int // Seconds to serve stale content on errors (0 = disabled)
}

// GenerateETag creates a unique ETag based on the SHA-256 hash of the payload.
// Uses URL-safe base64 encoding for compatibility.
func GenerateETag(payload []byte, identifier string) string {
	hash := sha256.Sum256(payload)
	encodedHash := base64.RawURLEncoding.EncodeToString(hash[:])
	return fmt.Sprintf(`"%s-%s"`, identifier, encodedHash)
}

// CheckETagMatch checks if the given ETag is present in the If-None-Match header value.
// The header can contain a comma-separated list of ETags.
func CheckETagMatch(etag string, ifNoneMatchHeader string) bool {
	if ifNoneMatchHeader == "" {
		return false
	}
	tags := strings.Split(ifNoneMatchHeader, ",")
	for _, tag := range tags {
		if strings.TrimSpace(tag) == etag {
			return true
		}
	}
	return false
}

// CreateCommonHeaders generates caching headers based on the provided configuration.
func CreateCommonHeaders(etag string, config CacheConfig) map[string]string {
	headers := map[string]string{
		"ETag": etag,
	}

	// Build Cache-Control header
	cacheControlParts := []string{"public"}
	if config.MaxAge > 0 {
		cacheControlParts = append(cacheControlParts, fmt.Sprintf("max-age=%d", config.MaxAge))
	} else {
		cacheControlParts = append(cacheControlParts, "max-age=0")
	}

	if config.SMaxAge > 0 {
		cacheControlParts = append(cacheControlParts, fmt.Sprintf("s-maxage=%d", config.SMaxAge))
	}

	if config.StaleWhileRevalidate > 0 {
		cacheControlParts = append(cacheControlParts, fmt.Sprintf("stale-while-revalidate=%d", config.StaleWhileRevalidate))
	}

	if config.StaleIfError > 0 {
		cacheControlParts = append(cacheControlParts, fmt.Sprintf("stale-if-error=%d", config.StaleIfError))
	}

	headers["Cache-Control"] = strings.Join(cacheControlParts, ", ")

	// CDN cache control (typically matches s-maxage or a custom value)
	if config.SMaxAge > 0 {
		cdnCacheControl := fmt.Sprintf("public, max-age=%d", config.SMaxAge)
		
		// Add stale directives for CDN if configured
		if config.StaleWhileRevalidate > 0 {
			cdnCacheControl += fmt.Sprintf(", stale-while-revalidate=%d", config.StaleWhileRevalidate)
		}
		if config.StaleIfError > 0 {
			cdnCacheControl += fmt.Sprintf(", stale-if-error=%d", config.StaleIfError)
		}
		
		headers["CDN-Cache-Control"] = cdnCacheControl
	}

	return headers
}

// ResponseCapture wraps http.ResponseWriter to capture response data
type ResponseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	written    bool
}

// NewResponseCapture creates a new ResponseCapture
func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           bytes.NewBuffer(nil),
		written:        false,
	}
}

// WriteHeader captures the status code
func (rc *ResponseCapture) WriteHeader(code int) {
	rc.statusCode = code
}

// Write captures the response body
func (rc *ResponseCapture) Write(data []byte) (int, error) {
	rc.written = true
	return rc.body.Write(data)
}

// Flush sends the captured response to the original ResponseWriter
func (rc *ResponseCapture) Flush() {
	if !rc.written {
		rc.WriteHeader(rc.statusCode)
		return
	}

	// Copy all headers
	for key, values := range rc.Header() {
		rc.Header()[key] = values
	}

	rc.WriteHeader(rc.statusCode)
	// Ignore error as per Go best practices for HTTP response writing
	_, _ = rc.Write(rc.body.Bytes())
}

// GetBody returns the captured response body
func (rc *ResponseCapture) GetBody() []byte {
	return rc.body.Bytes()
}

// GetStatusCode returns the captured status code
func (rc *ResponseCapture) GetStatusCode() int {
	return rc.statusCode
}

// CachingMiddleware creates middleware that sets cache headers and ETags
func CachingMiddleware(cacheOpts CacheOptions) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !cacheOpts.Enabled {
				next(w, r)
				return
			}

			// Generate ETag from the route identifier
			etag := GenerateETag([]byte(cacheOpts.ETagKey), cacheOpts.ETagKey)

			// Set cache headers with ETag
			commonHeaders := CreateCommonHeaders(etag, cacheOpts.Config)
			for key, value := range commonHeaders {
				w.Header().Set(key, value)
			}

			// Execute the handler
			next(w, r)
		}
	}
}

// MethodAndCache combines method validation and caching
func MethodAndCache(method string, cacheOpts CacheOptions) func(http.HandlerFunc) http.HandlerFunc {
	return CombineMiddlewares(
		MethodValidator(method),
		CachingMiddleware(cacheOpts),
	)
}

// QuickCache provides simple caching setup for common use cases
func QuickCache(method, etagKey string, maxAge, sMaxAge, staleWhileRevalidate int) func(http.HandlerFunc) http.HandlerFunc {
	cacheOpts := CacheOptions{
		Config: CacheConfig{
			MaxAge:               maxAge,
			SMaxAge:              sMaxAge,
			StaleWhileRevalidate: staleWhileRevalidate,
			StaleIfError:         0,
		},
		ETagKey: etagKey,
		Enabled: true,
	}

	return MethodAndCache(method, cacheOpts)
}

// NoCache disables caching for a route
func NoCache(method string) func(http.HandlerFunc) http.HandlerFunc {
	return MethodValidator(method)
}

// DefaultCache creates a zero-cache configuration (no caching by default)
func DefaultCache() CacheOptions {
	return CacheOptions{
		Config: CacheConfig{
			MaxAge:               0, // No browser caching
			SMaxAge:              0, // No CDN caching
			StaleWhileRevalidate: 0, // No stale-while-revalidate
			StaleIfError:         0, // No stale-if-error
		},
		Enabled: false, // Caching disabled by default
	}
}
