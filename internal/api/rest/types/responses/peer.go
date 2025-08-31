package responses

import "time"

// PeerResponse represents a peer response
type PeerResponse struct {
	ID        string            `json:"id"`
	Address   string            `json:"address"`
	Port      int               `json:"port"`
	Status    string            `json:"status"`
	LastSeen  time.Time         `json:"last_seen"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// PeerListResponse represents a list of peers response
type PeerListResponse struct {
	Peers []PeerResponse `json:"peers"`
	Total int            `json:"total"`
}
