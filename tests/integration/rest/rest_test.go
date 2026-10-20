package rest_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Skpow1234/Shugo/internal/api/rest"
	"github.com/Skpow1234/Shugo/internal/api/rest/types/requests"
	"github.com/Skpow1234/Shugo/internal/api/rest/types/responses"
	"github.com/Skpow1234/Shugo/internal/app/fileserver"
	"github.com/Skpow1234/Shugo/internal/crypto"
	"github.com/Skpow1234/Shugo/internal/peer"
	"github.com/Skpow1234/Shugo/internal/storage"
	netp2p "github.com/Skpow1234/Shugo/internal/transport/p2p"
)

func setupTestServer() *rest.Server {
	config := rest.DefaultConfig()
	config.Port = ":0"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return rest.NewServer(config, logger, nil)
}

func setupTestServerWithFS(t *testing.T) *rest.Server {
	t.Helper()
	km, err := crypto.NewKeyManagerFromClusterKey([]byte("shugo-test-cluster-key-32byt"))
	if err != nil {
		t.Fatal(err)
	}
	nodeID := crypto.GenerateID()
	tr := netp2p.NewTCPTransport(netp2p.TCPTransportOpts{
		ListenAddr:    ":0",
		HandshakeFunc: netp2p.NOPHandshakeFunc,
		Decoder:       netp2p.LengthPrefixedDecoder{},
	})
	fs := fileserver.New(fileserver.Options{
		ID:                nodeID,
		KeyManager:        km,
		StorageRoot:       t.TempDir(),
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tr,
		ResourceLimits:    peer.DefaultResourceLimits(),
	})
	tr.OnPeer = fs.OnPeer
	tr.OnStream = fs.OnStream

	config := rest.DefaultConfig()
	config.Port = ":0"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return rest.NewServer(config, logger, fs)
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

	if response["message"] != "Shugo REST API" {
		t.Errorf("Expected message 'Shugo REST API', got '%v'", response["message"])
	}
}

func TestRESTAPIFiles(t *testing.T) {
	restServer := setupTestServerWithFS(t)

	// Empty list initially
	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	w := httptest.NewRecorder()
	restServer.FileEndpoints.HandleListFiles(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Upload via multipart
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("hello shugo")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest("POST", "/api/v1/files", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w = httptest.NewRecorder()
	restServer.FileEndpoints.HandleUploadFile(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d body=%s", w.Code, w.Body.String())
	}

	var uploaded responses.FileResponse
	if err := json.NewDecoder(w.Body).Decode(&uploaded); err != nil {
		t.Fatalf("Failed to decode upload response: %v", err)
	}
	if uploaded.Key != "hello.txt" {
		t.Fatalf("Expected key hello.txt, got %q", uploaded.Key)
	}

	// List should include the file
	req = httptest.NewRequest("GET", "/api/v1/files", nil)
	w = httptest.NewRecorder()
	restServer.FileEndpoints.HandleListFiles(w, req)
	var list responses.FileListResponse
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(list.Files))
	}

	// Get metadata
	req = httptest.NewRequest("GET", "/api/v1/files/get?key=hello.txt", nil)
	w = httptest.NewRecorder()
	restServer.FileEndpoints.HandleGetFile(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
}

func TestRESTAPIPeers(t *testing.T) {
	restServer := setupTestServerWithFS(t)

	req := httptest.NewRequest("GET", "/api/v1/peers", nil)
	w := httptest.NewRecorder()
	restServer.PeerEndpoints.HandleListPeers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response responses.PeerListResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(response.Peers) == 0 {
		t.Fatal("Expected at least self peer in response")
	}

	selfID := response.Peers[0].ID
	req = httptest.NewRequest("GET", "/api/v1/peers/get?id="+selfID, nil)
	w = httptest.NewRecorder()
	restServer.PeerEndpoints.HandleGetPeer(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
}

func TestRESTAPIAddPeer(t *testing.T) {
	restServer := setupTestServerWithFS(t)

	// Dial to an unused address should fail closed (no mock success).
	request := requests.PeerAddRequest{
		Address: "127.0.0.1",
		Port:    1,
	}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/peers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	restServer.PeerEndpoints.HandleAddPeer(w, req)
	if w.Code == http.StatusCreated {
		t.Fatal("Expected add-peer to fail for unreachable address")
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
