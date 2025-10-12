package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a GraphQL client
type Client struct {
	httpClient *http.Client
	baseURL    string
	authToken  string
}

// New creates a new GraphQL client
func New(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// SetAuthToken sets the authentication token
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
}

// Request represents a GraphQL request
type Request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Response represents a GraphQL response
type Response struct {
	Data   interface{} `json:"data"`
	Errors []Error     `json:"errors,omitempty"`
}

// Error represents a GraphQL error
type Error struct {
	Message    string                 `json:"message"`
	Locations  []Location             `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Location represents an error location
type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Execute executes a GraphQL query
func (c *Client) Execute(ctx context.Context, query string, variables map[string]interface{}) (*Response, error) {
	req := Request{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Errors) > 0 {
		return &response, fmt.Errorf("GraphQL errors: %v", response.Errors)
	}

	return &response, nil
}

// File queries
const (
	FilesQuery = `
		query GetFiles($limit: Int, $offset: Int) {
			files(limit: $limit, offset: $offset) {
				id
				key
				size
				hash
				createdAt
				owner
			}
		}
	`

	FileQuery = `
		query GetFile($id: ID!) {
			file(id: $id) {
				id
				key
				size
				hash
				createdAt
				owner
			}
		}
	`

	StoreFileMutation = `
		mutation StoreFile($input: FileInput!) {
			storeFile(input: $input) {
				id
				key
				size
				hash
				createdAt
				owner
			}
		}
	`

	DeleteFileMutation = `
		mutation DeleteFile($id: ID!) {
			deleteFile(id: $id) {
				success
				message
			}
		}
	`
)

// Peer queries
const (
	PeersQuery = `
		query GetPeers {
			peers {
				id
				address
				status
				latency
				storage
				lastSeen
			}
		}
	`

	AddPeerMutation = `
		mutation AddPeer($input: PeerInput!) {
			addPeer(input: $input) {
				id
				address
				status
				latency
				storage
				lastSeen
			}
		}
	`

	RemovePeerMutation = `
		mutation RemovePeer($id: ID!) {
			removePeer(id: $id) {
				success
				message
			}
		}
	`
)

// System queries
const (
	HealthQuery = `
		query GetHealth {
			health {
				status
				timestamp
				services {
					name
					status
				}
			}
		}
	`

	MetricsQuery = `
		query GetMetrics {
			metrics {
				filesStored
				networkTraffic
				activePeers
				storageUsed
			}
		}
	`
)

// IoT queries
const (
	DevicesQuery = `
		query GetDevices {
			devices {
				id
				name
				type
				status
				location
				capabilities
				lastSeen
			}
		}
	`

	DeviceDataQuery = `
		query GetDeviceData($deviceId: ID!, $limit: Int) {
			deviceData(deviceId: $deviceId, limit: $limit) {
				timestamp
				sensorData
				actuatorCommands
			}
		}
	`
)

// Blockchain queries
const (
	BlockchainStatusQuery = `
		query GetBlockchainStatus {
			blockchain {
				status
				latestBlock
				transactionCount
				networkHashRate
			}
		}
	`

	TransactionsQuery = `
		query GetTransactions($limit: Int, $offset: Int) {
			transactions(limit: $limit, offset: $offset) {
				id
				hash
				from
				to
				value
				timestamp
				status
			}
		}
	`
)

// Backup queries
const (
	BackupsQuery = `
		query GetBackups {
			backups {
				id
				type
				status
				createdAt
				size
				files
			}
		}
	`

	CreateBackupMutation = `
		mutation CreateBackup($input: BackupInput!) {
			createBackup(input: $input) {
				id
				type
				status
				createdAt
			}
		}
	`

	RestoreBackupMutation = `
		mutation RestoreBackup($id: ID!) {
			restoreBackup(id: $id) {
				success
				message
			}
		}
	`
)
