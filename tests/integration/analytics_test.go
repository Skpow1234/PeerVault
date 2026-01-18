package integration

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/analytics"
	"github.com/Skpow1234/Peervault/internal/api/rest"
)

func TestAnalyticsIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	// Create server with analytics enabled
	config := rest.DefaultConfig()
	config.AnalyticsConfig.Enabled = true
	server := rest.NewServer(config, logger)
	
	// Create test server
	mux := http.NewServeMux()
	
	// Add a test endpoint
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test response")); err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	})
	
	// Wrap with analytics middleware
	analyticsMiddleware := analytics.Middleware(server.GetAnalyticsService(), logger)
	handler := analyticsMiddleware(mux)
	
	ts := httptest.NewServer(handler)
	defer ts.Close()
	
	// Make test requests
	t.Run("Record API Calls", func(t *testing.T) {
		// Make multiple requests
		for i := 0; i < 10; i++ {
			resp, err := http.Get(ts.URL + "/test")
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}
		
		// Wait for async recording
		time.Sleep(100 * time.Millisecond)
		
		// Verify analytics recorded the calls
		service := server.GetAnalyticsService()
		now := time.Now()
		metrics, err := service.GetUsageMetrics(now.Add(-1*time.Minute), now)
		if err != nil {
			t.Fatalf("Failed to get usage metrics: %v", err)
		}
		
		if metrics.TotalRequests < 10 {
			t.Errorf("Expected at least 10 requests, got %d", metrics.TotalRequests)
		}
	})
	
	t.Run("Get Usage Metrics", func(t *testing.T) {
		service := server.GetAnalyticsService()
		now := time.Now()
		
		metrics, err := service.GetUsageMetrics(now.Add(-1*time.Hour), now)
		if err != nil {
			t.Fatalf("Failed to get usage metrics: %v", err)
		}
		
		if metrics.TotalRequests == 0 {
			t.Error("Expected some requests to be recorded")
		}
		
		if metrics.TimeWindow.Start.IsZero() {
			t.Error("Expected time window to be set")
		}
	})
	
	t.Run("Get Analytics Summary", func(t *testing.T) {
		service := server.GetAnalyticsService()
		
		summary, err := service.GetAnalyticsSummary()
		if err != nil {
			t.Fatalf("Failed to get analytics summary: %v", err)
		}
		
		if summary.GeneratedAt.IsZero() {
			t.Error("Expected generated_at to be set")
		}
		
		if summary.SystemHealth.Status == "" {
			t.Error("Expected system health status")
		}
	})
	
	t.Run("Get Usage Trends", func(t *testing.T) {
		service := server.GetAnalyticsService()
		now := time.Now()
		
		trends, err := service.GetUsageTrends(now.Add(-2*time.Hour), now, "hour")
		if err != nil {
			t.Fatalf("Failed to get usage trends: %v", err)
		}
		
		if len(trends) == 0 {
			t.Error("Expected some trend data")
		}
	})
	
	t.Run("Get Popularity Metrics", func(t *testing.T) {
		service := server.GetAnalyticsService()
		now := time.Now()
		
		popularity, err := service.GetPopularityMetrics(now.Add(-1*time.Hour), now)
		if err != nil {
			t.Fatalf("Failed to get popularity metrics: %v", err)
		}
		
		if popularity.TimeWindow.Start.IsZero() {
			t.Error("Expected time window to be set")
		}
	})
}

func TestAnalyticsEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	// Create analytics service
	storage := analytics.NewMemoryStorage(1000, 30)
	service := analytics.NewService(storage, logger, nil)
	handler := analytics.NewHandler(service, logger)
	
	// Populate with test data
	now := time.Now()
	for i := 0; i < 50; i++ {
		call := &analytics.APICall{
			Timestamp:  now.Add(time.Duration(-i) * time.Minute),
			Method:     "GET",
			Path:       "/api/v1/files",
			StatusCode: 200,
			Duration:   time.Duration(40+i) * time.Millisecond,
			Endpoint:   "/api/v1/files",
			UserID:     "test-user",
		}
		if err := service.RecordAPICall(call); err != nil {
			t.Fatalf("Failed to record API call: %v", err)
		}
	}
	
	t.Run("GET /analytics/summary", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/analytics/summary", nil)
		rec := httptest.NewRecorder()
		
		handler.HandleGetSummary(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		var summary analytics.AnalyticsSummary
		if err := json.NewDecoder(rec.Body).Decode(&summary); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if summary.OverallMetrics.TotalRequests == 0 {
			t.Error("Expected some requests in summary")
		}
	})
	
	t.Run("GET /analytics/usage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/analytics/usage?period=24h", nil)
		rec := httptest.NewRecorder()
		
		handler.HandleGetUsageMetrics(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		var metrics analytics.UsageMetrics
		if err := json.NewDecoder(rec.Body).Decode(&metrics); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if metrics.TotalRequests != 50 {
			t.Errorf("Expected 50 requests, got %d", metrics.TotalRequests)
		}
	})
	
	t.Run("GET /analytics/endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/analytics/endpoint?endpoint=/api/v1/files&period=24h", nil)
		rec := httptest.NewRecorder()
		
		handler.HandleGetEndpointStats(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		var stats analytics.EndpointStats
		if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if stats.TotalCalls != 50 {
			t.Errorf("Expected 50 calls, got %d", stats.TotalCalls)
		}
	})
	
	t.Run("GET /analytics/user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/analytics/user?user_id=test-user&period=24h", nil)
		rec := httptest.NewRecorder()
		
		handler.HandleGetUserBehavior(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		var behavior analytics.UserBehaviorMetrics
		if err := json.NewDecoder(rec.Body).Decode(&behavior); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if behavior.UserID != "test-user" {
			t.Errorf("Expected user ID 'test-user', got '%s'", behavior.UserID)
		}
		
		if behavior.TotalRequests != 50 {
			t.Errorf("Expected 50 requests, got %d", behavior.TotalRequests)
		}
	})
	
	t.Run("GET /analytics/trends", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/analytics/trends?period=24h&interval=hour", nil)
		rec := httptest.NewRecorder()
		
		handler.HandleGetUsageTrends(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		var trends []*analytics.UsageTrend
		if err := json.NewDecoder(rec.Body).Decode(&trends); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if len(trends) == 0 {
			t.Error("Expected some trend data")
		}
	})
	
	t.Run("GET /analytics/popularity", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/analytics/popularity?period=24h", nil)
		rec := httptest.NewRecorder()
		
		handler.HandleGetPopularityMetrics(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		var popularity analytics.PopularityMetrics
		if err := json.NewDecoder(rec.Body).Decode(&popularity); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if len(popularity.TopEndpoints) == 0 {
			t.Error("Expected some top endpoints")
		}
	})
	
	t.Run("GET /analytics/dashboard", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/analytics/dashboard?period=24h", nil)
		rec := httptest.NewRecorder()
		
		handler.HandleGetDashboard(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		var dashboard map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&dashboard); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if dashboard["summary"] == nil {
			t.Error("Expected summary in dashboard")
		}
		if dashboard["metrics"] == nil {
			t.Error("Expected metrics in dashboard")
		}
	})
}

func TestAnalyticsMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	storage := analytics.NewMemoryStorage(1000, 30)
	service := analytics.NewService(storage, logger, nil)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	})
	
	middleware := analytics.Middleware(service, logger)
	wrappedHandler := middleware(handler)
	
	t.Run("Records API Call", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		req.Header.Set("X-User-ID", "middleware-test-user")
		req.Header.Set("User-Agent", "test-agent")
		rec := httptest.NewRecorder()
		
		wrappedHandler.ServeHTTP(rec, req)
		
		// Wait for async recording
		time.Sleep(100 * time.Millisecond)
		
		// Verify recording
		now := time.Now()
		calls, err := service.QueryAPICalls(&analytics.AnalyticsQuery{
			StartTime: now.Add(-1 * time.Minute),
			EndTime:   now,
			UserID:    "middleware-test-user",
		})
		
		if err != nil {
			t.Fatalf("Failed to query calls: %v", err)
		}
		
		if len(calls) == 0 {
			t.Error("Expected call to be recorded")
		} else {
			if calls[0].UserID != "middleware-test-user" {
				t.Errorf("Expected user ID 'middleware-test-user', got '%s'", calls[0].UserID)
			}
			if calls[0].UserAgent != "test-agent" {
				t.Errorf("Expected user agent 'test-agent', got '%s'", calls[0].UserAgent)
			}
		}
	})
}

