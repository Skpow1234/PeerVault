package types

import (
	"time"
)

// File represents a file in the distributed storage system
type File struct {
	ID        string         `json:"id"`
	Key       string         `json:"key"`
	HashedKey string         `json:"hashedKey"`
	Size      int64          `json:"size"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Owner     *Node          `json:"owner"`
	Replicas  []*FileReplica `json:"replicas"`
	Metadata  *FileMetadata  `json:"metadata"`
}

// FileMetadata contains additional information about a file
type FileMetadata struct {
	ContentType string                 `json:"contentType"`
	Checksum    string                 `json:"checksum"`
	Tags        []string               `json:"tags"`
	CustomFields map[string]interface{} `json:"customFields"`
}

// FileReplica represents a replica of a file on a specific node
type FileReplica struct {
	Node     *Node         `json:"node"`
	Status   ReplicaStatus `json:"status"`
	LastSync *time.Time    `json:"lastSync"`
	Size     *int64        `json:"size"`
}

// ReplicaStatus represents the status of a file replica
type ReplicaStatus string

const (
	ReplicaStatusSynced   ReplicaStatus = "SYNCED"
	ReplicaStatusSyncing  ReplicaStatus = "SYNCING"
	ReplicaStatusFailed   ReplicaStatus = "FAILED"
	ReplicaStatusPending  ReplicaStatus = "PENDING"
)

// Node represents a node in the peer network
type Node struct {
	ID           string      `json:"id"`
	Address      string      `json:"address"`
	Port         int         `json:"port"`
	Status       NodeStatus  `json:"status"`
	LastSeen     *time.Time  `json:"lastSeen"`
	Health       *NodeHealth `json:"health"`
	Capabilities []string    `json:"capabilities"`
}

// NodeHealth represents the health status of a node
type NodeHealth struct {
	IsHealthy      bool      `json:"isHealthy"`
	LastHeartbeat  *time.Time `json:"lastHeartbeat"`
	ResponseTime   *float64   `json:"responseTime"`
	Uptime         *float64   `json:"uptime"`
	Errors         []string   `json:"errors"`
}

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodeStatusOnline    NodeStatus = "ONLINE"
	NodeStatusOffline   NodeStatus = "OFFLINE"
	NodeStatusDegraded  NodeStatus = "DEGRADED"
	NodeStatusUnknown   NodeStatus = "UNKNOWN"
)

// PeerNetwork represents the network of peers
type PeerNetwork struct {
	Nodes       []*Node           `json:"nodes"`
	Connections []*Connection     `json:"connections"`
	Topology    *NetworkTopology  `json:"topology"`
}

// Connection represents a connection between two nodes
type Connection struct {
	From         *Node            `json:"from"`
	To           *Node            `json:"to"`
	Status       ConnectionStatus `json:"status"`
	Latency      *float64         `json:"latency"`
	Bandwidth    *float64         `json:"bandwidth"`
	LastActivity *time.Time       `json:"lastActivity"`
}

// ConnectionStatus represents the status of a connection
type ConnectionStatus string

const (
	ConnectionStatusActive   ConnectionStatus = "ACTIVE"
	ConnectionStatusInactive ConnectionStatus = "INACTIVE"
	ConnectionStatusFailed   ConnectionStatus = "FAILED"
	ConnectionStatusPending  ConnectionStatus = "PENDING"
)

// NetworkTopology represents the topology of the network
type NetworkTopology struct {
	TotalNodes      int        `json:"totalNodes"`
	ConnectedNodes  int        `json:"connectedNodes"`
	AverageLatency  *float64   `json:"averageLatency"`
	NetworkDiameter *int       `json:"networkDiameter"`
	Clusters        []*Cluster `json:"clusters"`
}

// Cluster represents a cluster of nodes
type Cluster struct {
	ID     string  `json:"id"`
	Nodes  []*Node `json:"nodes"`
	Leader *Node   `json:"leader"`
	Size   int     `json:"size"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	Storage     *StorageMetrics     `json:"storage"`
	Network     *NetworkMetrics     `json:"network"`
	Performance *PerformanceMetrics `json:"performance"`
	Uptime      float64             `json:"uptime"`
}

// StorageMetrics represents storage-related metrics
type StorageMetrics struct {
	TotalSpace        int64   `json:"totalSpace"`
	UsedSpace         int64   `json:"usedSpace"`
	AvailableSpace    int64   `json:"availableSpace"`
	FileCount         int     `json:"fileCount"`
	ReplicationFactor *float64 `json:"replicationFactor"`
}

// NetworkMetrics represents network-related metrics
type NetworkMetrics struct {
	ActiveConnections    int     `json:"activeConnections"`
	TotalBytesTransferred int64  `json:"totalBytesTransferred"`
	AverageBandwidth     *float64 `json:"averageBandwidth"`
	ErrorRate            *float64 `json:"errorRate"`
}

// PerformanceMetrics represents performance-related metrics
type PerformanceMetrics struct {
	AverageResponseTime *float64 `json:"averageResponseTime"`
	RequestsPerSecond   *float64 `json:"requestsPerSecond"`
	ErrorRate           *float64 `json:"errorRate"`
	MemoryUsage         *float64 `json:"memoryUsage"`
	CPUUsage            *float64 `json:"cpuUsage"`
}

// FileUpload represents a file upload operation
type FileUpload struct {
	ID        string        `json:"id"`
	Key       string        `json:"key"`
	Size      int64         `json:"size"`
	Status    UploadStatus  `json:"status"`
	Progress  *float64      `json:"progress"`
	UploadedAt *time.Time   `json:"uploadedAt"`
	Replicas  []*FileReplica `json:"replicas"`
}

// UploadStatus represents the status of a file upload
type UploadStatus string

const (
	UploadStatusPending    UploadStatus = "PENDING"
	UploadStatusUploading  UploadStatus = "UPLOADING"
	UploadStatusReplicating UploadStatus = "REPLICATING"
	UploadStatusCompleted  UploadStatus = "COMPLETED"
	UploadStatusFailed     UploadStatus = "FAILED"
)

// FileDownload represents a file download operation
type FileDownload struct {
	ID           string         `json:"id"`
	Key          string         `json:"key"`
	Size         int64          `json:"size"`
	Status       DownloadStatus `json:"status"`
	Progress     *float64       `json:"progress"`
	DownloadedAt *time.Time     `json:"downloadedAt"`
	Source       *Node          `json:"source"`
}

// DownloadStatus represents the status of a file download
type DownloadStatus string

const (
	DownloadStatusPending    DownloadStatus = "PENDING"
	DownloadStatusDownloading DownloadStatus = "DOWNLOADING"
	DownloadStatusCompleted  DownloadStatus = "COMPLETED"
	DownloadStatusFailed     DownloadStatus = "FAILED"
)

// FileFilter represents filters for file queries
type FileFilter struct {
	Owner           *string    `json:"owner"`
	SizeMin         *int64     `json:"sizeMin"`
	SizeMax         *int64     `json:"sizeMax"`
	CreatedAtAfter  *time.Time `json:"createdAtAfter"`
	CreatedAtBefore *time.Time `json:"createdAtBefore"`
	Tags            []string   `json:"tags"`
}

// FileMetadataInput represents input for file metadata
type FileMetadataInput struct {
	ContentType  *string                 `json:"contentType"`
	Tags         []string                `json:"tags"`
	CustomFields map[string]interface{}  `json:"customFields"`
}

// ConfigurationInput represents input for configuration updates
type ConfigurationInput struct {
	StorageRoot       *string `json:"storageRoot"`
	ReplicationFactor *int    `json:"replicationFactor"`
	MaxFileSize       *int64  `json:"maxFileSize"`
	EncryptionEnabled *bool   `json:"encryptionEnabled"`
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	Type      AlertType           `json:"type"`
	Message   string              `json:"message"`
	Severity  AlertSeverity       `json:"severity"`
	Timestamp time.Time           `json:"timestamp"`
	Metrics   *PerformanceMetrics `json:"metrics"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeHighCPU       AlertType = "HIGH_CPU"
	AlertTypeHighMemory    AlertType = "HIGH_MEMORY"
	AlertTypeSlowResponse  AlertType = "SLOW_RESPONSE"
	AlertTypeNetworkError  AlertType = "NETWORK_ERROR"
	AlertTypeStorageFull   AlertType = "STORAGE_FULL"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "LOW"
	AlertSeverityMedium   AlertSeverity = "MEDIUM"
	AlertSeverityHigh     AlertSeverity = "HIGH"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)
