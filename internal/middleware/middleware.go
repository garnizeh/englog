package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/garnizeh/englog/internal/logging"
	"github.com/google/uuid"
)

// RequestMiddleware provides request logging and tracking functionality
type RequestMiddleware struct {
	logger *logging.Logger
}

// NewRequestMiddleware creates a new request middleware
func NewRequestMiddleware(logger *logging.Logger) *RequestMiddleware {
	return &RequestMiddleware{
		logger: logger,
	}
}

// LoggingMiddleware adds comprehensive request/response logging with request IDs
func (m *RequestMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate unique request ID
		requestID := uuid.New().String()

		// Add request ID to context
		ctx := context.WithValue(r.Context(), logging.RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Create logger with request ID
		requestLogger := m.logger.WithRequestID(requestID)

		// Log incoming request
		requestLogger.LogHTTPRequest(
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.Header.Get("User-Agent"),
			r.ContentLength,
		)

		// Log query parameters if present
		if r.URL.RawQuery != "" {
			requestLogger.Debug("Request query parameters",
				"query", r.URL.RawQuery,
			)
		}

		// Log request headers in debug mode
		if m.logger.Logger.Enabled(r.Context(), -4) { // Debug level
			headers := make(map[string]string)
			for name, values := range r.Header {
				if len(values) > 0 {
					// Avoid logging sensitive headers
					if !isSensitiveHeader(name) {
						headers[name] = values[0]
					}
				}
			}
			requestLogger.Debug("Request headers", "headers", headers)
		}

		// Create response writer wrapper to capture status code and response size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			responseSize:   0,
		}

		// Add request ID to response headers for debugging
		wrapped.Header().Set("X-Request-ID", requestID)

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log response
		requestLogger.LogHTTPResponse(
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration.Milliseconds(),
		)

		// Log additional response metrics
		requestLogger.Debug("Response details",
			"response_size_bytes", wrapped.responseSize,
			"status_code", wrapped.statusCode,
		)

		// Log slow requests as warnings
		if duration > 5*time.Second {
			requestLogger.Warn("Slow request detected",
				"duration_ms", duration.Milliseconds(),
				"threshold_ms", 5000,
			)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseSize += size
	return size, err
}

// RecoveryMiddleware provides panic recovery with structured logging
func (m *RequestMiddleware) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestLogger := m.logger.WithContext(r.Context())

				requestLogger.Error("Panic recovered",
					"error", err,
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)

				// Return 500 error
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(name string) bool {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-auth-token",
	}

	lowerName := strings.ToLower(name)
	for _, sensitive := range sensitiveHeaders {
		if lowerName == sensitive {
			return true
		}
	}
	return false
}

// PerformanceMiddleware logs performance metrics for requests
func (m *RequestMiddleware) PerformanceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		requestLogger := m.logger.WithContext(r.Context())

		// Log performance metrics
		requestLogger.LogPerformanceMetric("http_request", duration.Milliseconds(), map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": wrapped.statusCode,
		})

		// Log performance warnings for different thresholds
		if duration > 10*time.Second {
			requestLogger.Error("Extremely slow request",
				"duration_ms", duration.Milliseconds(),
				"threshold_type", "critical",
			)
		} else if duration > 5*time.Second {
			requestLogger.Warn("Very slow request",
				"duration_ms", duration.Milliseconds(),
				"threshold_type", "warning",
			)
		} else if duration > 1*time.Second {
			requestLogger.Info("Slow request",
				"duration_ms", duration.Milliseconds(),
				"threshold_type", "info",
			)
		}
	})
}
