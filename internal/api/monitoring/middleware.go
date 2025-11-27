package monitoring

import (
	"log/slog"
	"net/http"
	"time"
)

// ResponseWriter wraps http.ResponseWriter to capture response details
type ResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(size)
	return size, err
}

// Middleware creates performance monitoring middleware
func Middleware(service *Service, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !service.config.Enabled {
				next.ServeHTTP(w, r)
				return
			}
			
			// Skip monitoring endpoints to avoid recursion
			if shouldSkipMonitoring(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			
			// Track active requests
			service.IncrementActiveRequests()
			defer service.DecrementActiveRequests()
			
			// Record start time
			start := time.Now()
			
			// Wrap response writer
			rw := NewResponseWriter(w)
			
			// Get request size
			requestSize := r.ContentLength
			if requestSize < 0 {
				requestSize = 0
			}
			
			// Process request
			next.ServeHTTP(rw, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Create metric
			metric := PerformanceMetric{
				Timestamp:    start,
				Endpoint:     normalizeEndpoint(r.URL.Path),
				Method:       r.Method,
				ResponseTime: duration,
				StatusCode:   rw.statusCode,
				RequestSize:  requestSize,
				ResponseSize: rw.responseSize,
			}
			
			// Record metric asynchronously
			go service.RecordMetric(metric)
			
			// Log slow requests
			if duration > 1*time.Second {
				logger.Warn("Slow request detected",
					"method", r.Method,
					"path", r.URL.Path,
					"duration_ms", duration.Milliseconds(),
					"status", rw.statusCode,
				)
			}
		})
	}
}

// shouldSkipMonitoring determines if a path should skip monitoring
func shouldSkipMonitoring(path string) bool {
	skipPaths := []string{
		"/monitoring",
		"/health",
		"/metrics",
	}
	
	for _, skip := range skipPaths {
		if len(path) >= len(skip) && path[:len(skip)] == skip {
			return true
		}
	}
	
	return false
}

// normalizeEndpoint normalizes endpoint path for grouping
func normalizeEndpoint(path string) string {
	// Simple normalization - in production, use more sophisticated logic
	// This should match analytics.normalizeEndpoint for consistency
	
	// Remove trailing slash
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	
	return path
}

