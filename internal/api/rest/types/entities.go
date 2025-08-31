package types

import "time"

// File represents a file in the PeerVault system
type File struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	Hash        string            `json:"hash"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Replicas    []FileReplica     `json:"replicas,omitempty"`
}

// FileReplica represents a replica of a file on a peer
type FileReplica struct {
	PeerID    string    `json:"peer_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Peer represents a peer node in the PeerVault network
type Peer struct {
	ID        string            `json:"id"`
	Address   string            `json:"address"`
	Port      int               `json:"port"`
	Status    string            `json:"status"`
	LastSeen  time.Time         `json:"last_seen"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// SystemInfo represents system information and status
type SystemInfo struct {
	Version      string        `json:"version"`
	Uptime       time.Duration `json:"uptime"`
	StartTime    time.Time     `json:"start_time"`
	StorageUsed  int64         `json:"storage_used"`
	StorageTotal int64         `json:"storage_total"`
	PeerCount    int           `json:"peer_count"`
	FileCount    int           `json:"file_count"`
}

// Metrics represents system metrics and performance data
type Metrics struct {
	RequestsTotal     int64     `json:"requests_total"`
	RequestsPerMinute float64   `json:"requests_per_minute"`
	ActiveConnections int       `json:"active_connections"`
	StorageUsage      float64   `json:"storage_usage_percent"`
	LastUpdated       time.Time `json:"last_updated"`
}
