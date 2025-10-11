package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

// New creates a new client instance
func New(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   cfg.ServerURL,
		authToken: cfg.AuthToken,
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

// makeRequest makes an HTTP request to the API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
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
	defer resp.Body.Close()

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

// StoreFile stores a file
func (c *Client) StoreFile(ctx context.Context, filePath string) (*FileInfo, error) {
	// This would implement file upload logic
	// For now, return a mock response
	return &FileInfo{
		ID:        "mock-file-id",
		Key:       filePath,
		Size:      1024,
		Hash:      "mock-hash",
		CreatedAt: time.Now(),
		Owner:     "current-user",
	}, nil
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
