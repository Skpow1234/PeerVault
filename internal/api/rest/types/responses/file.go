package responses

import "time"

// FileResponse represents a file response
type FileResponse struct {
	Key         string                `json:"key"`
	Name        string                `json:"name"`
	Size        int64                 `json:"size"`
	ContentType string                `json:"content_type"`
	Hash        string                `json:"hash"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Metadata    map[string]string     `json:"metadata,omitempty"`
	Replicas    []FileReplicaResponse `json:"replicas,omitempty"`
}

// FileReplicaResponse represents a file replica response
type FileReplicaResponse struct {
	PeerID    string    `json:"peer_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// FileListResponse represents a list of files response
type FileListResponse struct {
	Files []FileResponse `json:"files"`
	Total int            `json:"total"`
}
