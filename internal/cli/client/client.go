package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/config"
)

// Client represents a PeerVault API client
type Client struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
	authToken  string
	connected  bool
	retryCount int
}

// New creates a new client instance
func New(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    cfg.ServerURL,
		authToken:  cfg.AuthToken,
		connected:  false,
		retryCount: 3,
	}
}

// SetServerURL sets the server URL
func (c *Client) SetServerURL(url string) {
	c.baseURL = url
}

// SetAuthToken sets the authentication token
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.connected
}

// Connect tests the connection to the server
func (c *Client) Connect(ctx context.Context) error {
	// Test connection with a health check
	_, err := c.GetHealth(ctx)
	if err != nil {
		c.connected = false
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	c.connected = true
	return nil
}

// Disconnect disconnects from the server
func (c *Client) Disconnect() {
	c.connected = false
}

// SetRetryCount sets the number of retries for failed requests
func (c *Client) SetRetryCount(count int) {
	c.retryCount = count
}

// makeRequest makes an HTTP request to the API with retry logic
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		if attempt > 0 {
			// Wait before retry (exponential backoff)
			waitTime := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}

		resp, err := c.doRequest(ctx, method, endpoint, body)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryCount+1, lastErr)
}

// doRequest performs a single HTTP request
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// Add authentication header
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	// Add content type for JSON requests
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// Get makes a GET request
func (c *Client) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.makeRequest(ctx, "GET", endpoint, nil)
}

// Post makes a POST request
func (c *Client) Post(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	return c.makeRequest(ctx, "POST", endpoint, body)
}

// Put makes a PUT request
func (c *Client) Put(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	return c.makeRequest(ctx, "PUT", endpoint, body)
}

// Delete makes a DELETE request
func (c *Client) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.makeRequest(ctx, "DELETE", endpoint, nil)
}

// ParseResponse parses an HTTP response into the target interface
func (c *Client) ParseResponse(resp *http.Response, target interface{}) error {
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if target == nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

// File operations
type FileInfo struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Size      int64     `json:"size"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	Owner     string    `json:"owner"`
}

type FileListResponse struct {
	Files []FileInfo `json:"files"`
	Total int        `json:"total"`
}

// StoreFile stores a file with real upload functionality
func (c *Client) StoreFile(ctx context.Context, filePath string) (*FileInfo, error) {
	// Check if file exists
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Get file info
	_, err = file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/files", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check response
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result FileInfo
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// DownloadFile downloads a file to the specified path
func (c *Client) DownloadFile(ctx context.Context, fileID, outputPath string) error {
	// Get file info first
	_, err := c.GetFile(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	// Download file
	resp, err := c.Get(ctx, "/api/v1/files/"+fileID+"/download")
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check response
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed %d: %s", resp.StatusCode, string(body))
	}

	// Copy content to file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// GetFile retrieves file information
func (c *Client) GetFile(ctx context.Context, fileID string) (*FileInfo, error) {
	resp, err := c.Get(ctx, "/api/v1/files/"+fileID)
	if err != nil {
		return nil, err
	}

	var file FileInfo
	err = c.ParseResponse(resp, &file)
	return &file, err
}

// ListFiles lists all files
func (c *Client) ListFiles(ctx context.Context) (*FileListResponse, error) {
	resp, err := c.Get(ctx, "/api/v1/files")
	if err != nil {
		return nil, err
	}

	var files FileListResponse
	err = c.ParseResponse(resp, &files)
	return &files, err
}

// DeleteFile deletes a file
func (c *Client) DeleteFile(ctx context.Context, fileID string) error {
	resp, err := c.Delete(ctx, "/api/v1/files/"+fileID)
	if err != nil {
		return err
	}

	return c.ParseResponse(resp, nil)
}

// Peer operations
type PeerInfo struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	Status   string    `json:"status"`
	Latency  int64     `json:"latency"`
	Storage  int64     `json:"storage"`
	LastSeen time.Time `json:"last_seen"`
}

type PeerListResponse struct {
	Peers []PeerInfo `json:"peers"`
	Total int        `json:"total"`
}

// ListPeers lists all peers
func (c *Client) ListPeers(ctx context.Context) (*PeerListResponse, error) {
	resp, err := c.Get(ctx, "/api/v1/peers")
	if err != nil {
		return nil, err
	}

	var peers PeerListResponse
	err = c.ParseResponse(resp, &peers)
	return &peers, err
}

// AddPeer adds a new peer
func (c *Client) AddPeer(ctx context.Context, address string) (*PeerInfo, error) {
	body := strings.NewReader(fmt.Sprintf(`{"address": "%s"}`, address))
	resp, err := c.Post(ctx, "/api/v1/peers", body)
	if err != nil {
		return nil, err
	}

	var peer PeerInfo
	err = c.ParseResponse(resp, &peer)
	return &peer, err
}

// RemovePeer removes a peer
func (c *Client) RemovePeer(ctx context.Context, peerID string) error {
	resp, err := c.Delete(ctx, "/api/v1/peers/"+peerID)
	if err != nil {
		return err
	}

	return c.ParseResponse(resp, nil)
}

// System operations
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

type Metrics struct {
	FilesStored    int64   `json:"files_stored"`
	NetworkTraffic float64 `json:"network_traffic"`
	ActivePeers    int     `json:"active_peers"`
	StorageUsed    int64   `json:"storage_used"`
}

// GetHealth gets system health
func (c *Client) GetHealth(ctx context.Context) (*HealthStatus, error) {
	resp, err := c.Get(ctx, "/api/v1/health")
	if err != nil {
		return nil, err
	}

	var health HealthStatus
	err = c.ParseResponse(resp, &health)
	return &health, err
}

// GetMetrics gets system metrics
func (c *Client) GetMetrics(ctx context.Context) (*Metrics, error) {
	resp, err := c.Get(ctx, "/api/v1/metrics")
	if err != nil {
		return nil, err
	}

	var metrics Metrics
	err = c.ParseResponse(resp, &metrics)
	return &metrics, err
}
