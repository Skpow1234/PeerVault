package grpc

import (
	"time"
)

// Basic types for gRPC implementation
// These will be replaced with generated protobuf types once protoc is set up

type FileChunk struct {
	FileKey  string
	Data     []byte
	Offset   int64
	IsLast   bool
	Checksum string
}

type FileRequest struct {
	Key string
}

type FileResponse struct {
	Key         string
	Name        string
	Size        int64
	ContentType string
	Hash        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    map[string]string
	Replicas    []FileReplica
}

type FileReplica struct {
	PeerID    string
	Status    string
	CreatedAt time.Time
}

type ListFilesRequest struct {
	Page     int32
	PageSize int32
	Filter   string
}

type ListFilesResponse struct {
	Files    []FileResponse
	Total    int32
	Page     int32
	PageSize int32
}

type UpdateFileMetadataRequest struct {
	Key      string
	Metadata map[string]string
}

type DeleteFileResponse struct {
	Success bool
	Message string
}

type PeerRequest struct {
	ID string
}

type PeerResponse struct {
	ID        string
	Address   string
	Port      int32
	Status    string
	LastSeen  time.Time
	CreatedAt time.Time
	Metadata  map[string]string
}

type AddPeerRequest struct {
	Address  string
	Port     int32
	Metadata map[string]string
}

type ListPeersResponse struct {
	Peers []PeerResponse
	Total int32
}

type RemovePeerResponse struct {
	Success bool
	Message string
}

type PeerHealthResponse struct {
	PeerID        string
	Status        string
	LatencyMs     float64
	UptimeSeconds int32
	Metrics       map[string]string
}

type SystemInfoResponse struct {
	Version       string
	UptimeSeconds int64
	StartTime     time.Time
	StorageUsed   int64
	StorageTotal  int64
	PeerCount     int32
	FileCount     int32
}

type MetricsResponse struct {
	RequestsTotal       int64
	RequestsPerMinute   float64
	ActiveConnections   int32
	StorageUsagePercent float64
	LastUpdated         time.Time
}

type HealthResponse struct {
	Status    string
	Timestamp time.Time
	Version   string
}

type FileOperationEvent struct {
	EventType string
	FileKey   string
	PeerID    string
	Timestamp time.Time
	Metadata  map[string]string
}

type PeerEvent struct {
	EventType string
	PeerID    string
	Timestamp time.Time
	Metadata  map[string]string
}

type SystemEvent struct {
	EventType string
	Component string
	Timestamp time.Time
	Message   string
	Metadata  map[string]string
}

type Empty struct{}
