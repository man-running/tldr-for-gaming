package middleware

import (
	"main/lib/response"
	"net/http"
)

// MethodValidator is a middleware that validates HTTP methods
func MethodValidator(allowedMethods ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if the request method is allowed
			for _, method := range allowedMethods {
				if r.Method == method {
					next(w, r)
					return
				}
			}

			// Method not allowed
			response.WriteJSONResponse(w, http.StatusMethodNotAllowed, ErrorResponse{
				Error: "Method not allowed",
			})
		}
	}
}

// ErrorResponse is a common error response structure
type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// SuccessResponse is a common success response structure
type SuccessResponse struct {
	Success bool   `json:"success,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CacheOptions defines caching configuration for middleware
type CacheOptions struct {
	Config  CacheConfig
	ETagKey string
	Enabled bool
}

// DefaultCacheOptions returns sensible default cache options
func DefaultCacheOptions() CacheOptions {
	return CacheOptions{
		Config: CacheConfig{
			MaxAge:               0,    // No browser caching
			SMaxAge:              300,  // 5 minutes CDN cache
			StaleWhileRevalidate: 3600, // 1 hour stale-while-revalidate
			StaleIfError:         0,    // No stale-if-error
		},
		Enabled: true,
	}
}

// WithMethodAndCache combines method validation and caching middleware
func WithMethodAndCache(method string, cacheOpts CacheOptions) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 1. Method validation
			if r.Method != method {
				WriteJSONResponse(w, http.StatusMethodNotAllowed, ErrorResponse{
					Error: "Method not allowed",
				})
				return
			}

			// 2. Set cache headers if enabled
			if cacheOpts.Enabled {
				commonHeaders := CreateCommonHeaders("", cacheOpts.Config)
				for key, value := range commonHeaders {
					w.Header().Set(key, value)
				}
			}

			// 3. Execute the handler
			next(w, r)
		}
	}
}

// WithCaching is a middleware that adds caching headers and ETag support
func WithCaching(cacheOpts CacheOptions, etagKey string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !cacheOpts.Enabled {
				next(w, r)
				return
			}

			next(w, r)
		}
	}
}

// CachedResponse represents a response that can be cached
type CachedResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Status int         `json:"-"`
}

// CombineMiddlewares combines multiple middlewares into one
func CombineMiddlewares(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		// Apply middlewares in reverse order (last middleware wraps first)
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
