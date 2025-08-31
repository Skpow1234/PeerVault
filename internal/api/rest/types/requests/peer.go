package requests

// PeerAddRequest represents a request to add a new peer
type PeerAddRequest struct {
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
