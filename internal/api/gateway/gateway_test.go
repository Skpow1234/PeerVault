package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/rest/ratelimit"
)

// createTestLogger creates a test logger
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Only log errors in tests
	}))
}

func TestGatewayBasicRouting(t *testing.T) {
	// Create a test upstream server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "hello from upstream"}); err != nil {
			t.Errorf("Failed to encode JSON response: %v", err)
		}
	}))
	defer upstream.Close()

	config := &GatewayConfig{
		ListenAddr:      ":0", // Use random port
		UpstreamTimeout: 5 * time.Second,
		MaxRetries:      1,
		RetryDelay:      100 * time.Millisecond,
		Routes: []RouteConfig{
			{
				Path:        "/api/",
				Methods:     []string{"GET", "POST"},
				UpstreamURL: upstream.URL,
				StripPrefix: "/api",
			},
		},
	}

	gw, err := NewGateway(config, createTestLogger())
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	// Create a test request
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	// Find the route and test the handler
	route := gw.routes["/api/"]
	if route == nil {
		t.Fatal("Route not found")
	}

	handler := gw.createRouteHandler(route)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "hello from upstream" {
		t.Errorf("Expected 'hello from upstream', got '%s'", response["message"])
	}
}

func TestGatewayURLRewriting(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(r.URL.Path)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer upstream.Close()

	config := &GatewayConfig{
		ListenAddr:      ":0",
		UpstreamTimeout: 5 * time.Second,
		Routes: []RouteConfig{
			{
				Path:        "/v1/",
				Methods:     []string{"GET"},
				UpstreamURL: upstream.URL,
				StripPrefix: "/v1",
				RewriteRules: []RewriteRule{
					{Pattern: "/users", Replace: "/api/users"},
				},
			},
		},
	}

	gw, err := NewGateway(config, createTestLogger())
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/users/123", nil)
	w := httptest.NewRecorder()

	route := gw.routes["/v1/"]
	handler := gw.createRouteHandler(route)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expectedPath := "/api/users/123"
	if w.Body.String() != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, w.Body.String())
	}
}

func TestGatewayRequestTransformation(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			t.Errorf("Failed to unmarshal JSON: %v", err)
			return
		}

		// Check if transformed field exists
		if _, exists := data["gateway_processed"]; !exists {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(data); err != nil {
			t.Errorf("Failed to encode JSON response: %v", err)
		}
	}))
	defer upstream.Close()

	config := &GatewayConfig{
		ListenAddr:      ":0",
		UpstreamTimeout: 5 * time.Second,
		Routes: []RouteConfig{
			{
				Path:        "/api/",
				Methods:     []string{"POST"},
				UpstreamURL: upstream.URL,
				Transform: &TransformConfig{
					Request: &TransformRule{
						AddFields: map[string]interface{}{
							"gateway_processed": true,
							"timestamp":         "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		},
	}

	gw, err := NewGateway(config, createTestLogger())
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	requestBody := map[string]string{"name": "test"}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	route := gw.routes["/api/"]
	handler := gw.createRouteHandler(route)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGatewayResponseTransformation(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"original_field": "value",
			"unwanted_field": "remove_me",
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode JSON response: %v", err)
		}
	}))
	defer upstream.Close()

	config := &GatewayConfig{
		ListenAddr:      ":0",
		UpstreamTimeout: 5 * time.Second,
		Routes: []RouteConfig{
			{
				Path:        "/api/",
				Methods:     []string{"GET"},
				UpstreamURL: upstream.URL,
				Transform: &TransformConfig{
					Response: &TransformRule{
						RemoveFields: []string{"unwanted_field"},
						AddFields: map[string]interface{}{
							"gateway_transformed": true,
						},
						RenameFields: map[string]string{
							"original_field": "transformed_field",
						},
					},
				},
			},
		},
	}

	gw, err := NewGateway(config, createTestLogger())
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	route := gw.routes["/api/"]
	handler := gw.createRouteHandler(route)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check transformations
	if _, exists := response["unwanted_field"]; exists {
		t.Error("unwanted_field should have been removed")
	}

	if response["transformed_field"] != "value" {
		t.Errorf("Expected transformed_field to be 'value', got %v", response["transformed_field"])
	}

	if response["gateway_transformed"] != true {
		t.Errorf("Expected gateway_transformed to be true, got %v", response["gateway_transformed"])
	}
}

func TestGatewayMethodFiltering(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	config := &GatewayConfig{
		ListenAddr:      ":0",
		UpstreamTimeout: 5 * time.Second,
		Routes: []RouteConfig{
			{
				Path:        "/api/",
				Methods:     []string{"GET", "POST"},
				UpstreamURL: upstream.URL,
			},
		},
	}

	gw, err := NewGateway(config, createTestLogger())
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	// Test allowed method
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	route := gw.routes["/api/"]
	handler := gw.createRouteHandler(route)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for GET, got %d", w.Code)
	}

	// Test disallowed method
	req = httptest.NewRequest("DELETE", "/api/test", nil)
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for DELETE, got %d", w.Code)
	}
}

func TestGatewayRateLimiting(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	config := &GatewayConfig{
		ListenAddr:      ":0",
		UpstreamTimeout: 5 * time.Second,
		RateLimitConfig: &ratelimit.RateLimitConfig{
			Algorithm:       ratelimit.TokenBucket,
			RequestsPerMin:  2,
			BurstSize:       2,
			CleanupInterval: 5 * time.Minute,
			Enabled:         true,
		},
		Routes: []RouteConfig{
			{
				Path:        "/api/",
				Methods:     []string{"GET"},
				UpstreamURL: upstream.URL,
			},
		},
	}

	gw, err := NewGateway(config, createTestLogger())
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}
	defer func() {
		if err := gw.Stop(context.Background()); err != nil {
			t.Errorf("Failed to stop gateway: %v", err)
		}
	}()

	route := gw.routes["/api/"]
	handler := gw.createRouteHandler(route)

	// Make requests up to the limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "127.0.0.1:12345" // Set a consistent remote address
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed, got status %d", i+1, w.Code)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345" // Same address to trigger rate limiting
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 for rate limited request, got %d", w.Code)
	}
}

func TestGatewayAddRemoveRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	config := &GatewayConfig{
		ListenAddr:      ":0",
		UpstreamTimeout: 5 * time.Second,
	}

	gw, err := NewGateway(config, createTestLogger())
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	// Add a route
	routeConfig := RouteConfig{
		Path:        "/test/",
		Methods:     []string{"GET"},
		UpstreamURL: upstream.URL,
	}

	err = gw.AddRoute(routeConfig)
	if err != nil {
		t.Fatalf("Failed to add route: %v", err)
	}

	// Check if route exists
	routes := gw.ListRoutes()
	if _, exists := routes["/test/"]; !exists {
		t.Error("Route should exist after adding")
	}

	// Remove the route
	gw.RemoveRoute("/test/")

	// Check if route is removed
	routes = gw.ListRoutes()
	if _, exists := routes["/test/"]; exists {
		t.Error("Route should not exist after removing")
	}
}
