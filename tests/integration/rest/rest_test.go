package rest_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/rest"
	"github.com/Skpow1234/Peervault/internal/api/rest/types/requests"
	"github.com/Skpow1234/Peervault/internal/api/rest/types/responses"
)

func setupTestServer() *rest.Server {
	config := rest.DefaultConfig()
	config.Port = ":0" // Use random port for testing
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return rest.NewServer(config, logger)
}

func TestRESTAPIHealth(t *testing.T) {
	restServer := setupTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	restServer.SystemEndpoints.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response responses.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}
}

func TestRESTAPIMetrics(t *testing.T) {
	restServer := setupTestServer()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	restServer.SystemEndpoints.HandleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response responses.MetricsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.RequestsTotal <= 0 {
		t.Errorf("Expected positive requests total, got %d", response.RequestsTotal)
	}
}

func TestRESTAPISystemInfo(t *testing.T) {
	restServer := setupTestServer()

	req := httptest.NewRequest("GET", "/system", nil)
	w := httptest.NewRecorder()

	restServer.SystemEndpoints.HandleSystemInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response responses.SystemInfoResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Version == "" {
		t.Error("Expected non-empty version")
	}
}

func TestRESTAPIRoot(t *testing.T) {
	restServer := setupTestServer()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	restServer.SystemEndpoints.HandleRoot(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "PeerVault REST API" {
		t.Errorf("Expected message 'PeerVault REST API', got '%v'", response["message"])
	}
}

func TestRESTAPIFiles(t *testing.T) {
	restServer := setupTestServer()

	// Test list files
	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	w := httptest.NewRecorder()

	restServer.FileEndpoints.HandleListFiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response responses.FileListResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Files) == 0 {
		t.Error("Expected at least one file in response")
	}

	// Test get file
	req = httptest.NewRequest("GET", "/api/v1/files/get?key=file1", nil)
	w = httptest.NewRecorder()

	restServer.FileEndpoints.HandleGetFile(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var fileResponse responses.FileResponse
	if err := json.NewDecoder(w.Body).Decode(&fileResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if fileResponse.Key != "file1" {
		t.Errorf("Expected key 'file1', got '%s'", fileResponse.Key)
	}
}

func TestRESTAPIPeers(t *testing.T) {
	restServer := setupTestServer()

	// Test list peers
	req := httptest.NewRequest("GET", "/api/v1/peers", nil)
	w := httptest.NewRecorder()

	restServer.PeerEndpoints.HandleListPeers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response responses.PeerListResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Peers) == 0 {
		t.Error("Expected at least one peer in response")
	}

	// Test get peer
	req = httptest.NewRequest("GET", "/api/v1/peers/get?id=peer1", nil)
	w = httptest.NewRecorder()

	restServer.PeerEndpoints.HandleGetPeer(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var peerResponse responses.PeerResponse
	if err := json.NewDecoder(w.Body).Decode(&peerResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if peerResponse.ID != "peer1" {
		t.Errorf("Expected ID 'peer1', got '%s'", peerResponse.ID)
	}
}

func TestRESTAPIAddPeer(t *testing.T) {
	restServer := setupTestServer()

	request := requests.PeerAddRequest{
		Address: "192.168.1.100",
		Port:    8080,
		Metadata: map[string]string{
			"region": "us-east",
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/peers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	restServer.PeerEndpoints.HandleAddPeer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response responses.PeerResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Address != "192.168.1.100" {
		t.Errorf("Expected address '192.168.1.100', got '%s'", response.Address)
	}
}

func TestRESTAPIWebhook(t *testing.T) {
	restServer := setupTestServer()

	request := requests.WebhookRequest{
		Event:     "file.uploaded",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"file_id": "test123",
			"size":    1024,
		},
		Source: "test-client",
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	restServer.SystemEndpoints.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRESTAPIDocs(t *testing.T) {
	restServer := setupTestServer()

	req := httptest.NewRequest("GET", "/docs", nil)
	w := httptest.NewRecorder()

	restServer.SystemEndpoints.HandleDocs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected content type 'text/html', got '%s'", contentType)
	}
}

func TestRESTAPISwaggerJSON(t *testing.T) {
	restServer := setupTestServer()

	req := httptest.NewRequest("GET", "/swagger.json", nil)
	w := httptest.NewRecorder()

	restServer.SystemEndpoints.HandleSwaggerJSON(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", contentType)
	}

	var swaggerSpec map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&swaggerSpec); err != nil {
		t.Fatalf("Failed to decode swagger spec: %v", err)
	}

	if swaggerSpec["openapi"] != "3.0.3" {
		t.Errorf("Expected OpenAPI version '3.0.3', got '%v'", swaggerSpec["openapi"])
	}
}

func TestRESTAPICORS(t *testing.T) {
	restServer := setupTestServer()

	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	restServer.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Expected CORS header to be set")
	}
}
