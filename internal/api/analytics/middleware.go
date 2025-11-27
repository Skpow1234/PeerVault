package analytics

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// ResponseRecorder wraps http.ResponseWriter to capture response details
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Size       int64
}

// NewResponseRecorder creates a new ResponseRecorder
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code
func (r *ResponseRecorder) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.Size += int64(size)
	return size, err
}

// Middleware creates analytics tracking middleware
func Middleware(service *Service, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip analytics for certain paths
			if shouldSkip(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			
			start := time.Now()
			
			// Wrap response writer to capture details
			recorder := NewResponseRecorder(w)
			
			// Get request size
			requestSize := r.ContentLength
			if requestSize < 0 {
				requestSize = 0
			}
			
			// Process request
			next.ServeHTTP(recorder, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Extract user ID from context or header
			userID := extractUserID(r)
			
			// Create API call record
			call := &APICall{
				Timestamp:    start,
				Method:       r.Method,
				Path:         r.URL.Path,
				StatusCode:   recorder.StatusCode,
				Duration:     duration,
				UserID:       userID,
				UserAgent:    r.UserAgent(),
				IPAddress:    getClientIP(r),
				RequestSize:  requestSize,
				ResponseSize: recorder.Size,
				APIVersion:   extractAPIVersion(r),
				Endpoint:     normalizeEndpoint(r.URL.Path),
				QueryParams:  extractQueryParams(r),
			}
			
			// Record the API call asynchronously
			go func() {
				if err := service.RecordAPICall(call); err != nil {
					logger.Error("Failed to record analytics",
						"error", err,
						"path", r.URL.Path,
						"method", r.Method,
					)
				}
			}()
		})
	}
}

// shouldSkip determines if a path should be skipped for analytics
func shouldSkip(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/analytics", // Don't track analytics endpoints themselves
		"/favicon.ico",
		"/robots.txt",
	}
	
	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	
	return false
}

// extractUserID extracts user ID from request
func extractUserID(r *http.Request) string {
	// Try to get from header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Try to get from authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		// Simple extraction - in production you'd decode the token
		return strings.TrimPrefix(auth, "Bearer ")
	}
	
	return ""
}

// getClientIP extracts the client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	
	return ip
}

// extractAPIVersion extracts API version from path or header
func extractAPIVersion(r *http.Request) string {
	// Try to extract from path (e.g., /api/v1/...)
	if strings.HasPrefix(r.URL.Path, "/api/") {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 3 && strings.HasPrefix(parts[2], "v") {
			return parts[2]
		}
	}
	
	// Try to get from header
	if version := r.Header.Get("X-API-Version"); version != "" {
		return version
	}
	
	return "v1"
}

// normalizeEndpoint normalizes endpoint path for grouping
func normalizeEndpoint(path string) string {
	// Remove leading/trailing slashes
	path = strings.Trim(path, "/")
	
	// Split by slash
	parts := strings.Split(path, "/")
	
	// Replace UUIDs and IDs with placeholders
	for i, part := range parts {
		// Check if it looks like a UUID
		if len(part) == 36 && strings.Count(part, "-") == 4 {
			parts[i] = "{id}"
		} else if isNumeric(part) {
			parts[i] = "{id}"
		}
	}
	
	return "/" + strings.Join(parts, "/")
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// extractQueryParams extracts query parameters from request
func extractQueryParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			// Only store first value for simplicity
			params[key] = values[0]
		}
	}
	
	return params
}

// BodySizeMiddleware creates middleware to track request body size
func BodySizeMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If ContentLength is not set, try to read and measure body
			if r.ContentLength < 0 && r.Body != nil {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					r.ContentLength = int64(len(body))
					// Reset body for downstream handlers
					r.Body = io.NopCloser(strings.NewReader(string(body)))
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

