package services

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Skpow1234/Peervault/proto/peervault"
)

// PeerService provides peer-related operations
type PeerService struct {
	// Mock storage for demonstration
	peers map[string]*peervault.PeerResponse
}

// NewPeerService creates a new peer service instance
func NewPeerService() *PeerService {
	return &PeerService{
		peers: make(map[string]*peervault.PeerResponse),
	}
}

// ListPeers lists all peers
func (s *PeerService) ListPeers() (*peervault.ListPeersResponse, error) {
	// Mock implementation
	peers := []*peervault.PeerResponse{
		{
			Id:        "peer1",
			Address:   "192.168.1.100",
			Port:      50051,
			Status:    "active",
			LastSeen:  timestamppb.Now(),
			CreatedAt: timestamppb.Now(),
			Metadata:  map[string]string{},
		},
		{
			Id:        "peer2",
			Address:   "192.168.1.101",
			Port:      50051,
			Status:    "active",
			LastSeen:  timestamppb.Now(),
			CreatedAt: timestamppb.Now(),
			Metadata:  map[string]string{},
		},
	}

	return &peervault.ListPeersResponse{
		Peers: peers,
		Total: int32(len(peers)),
	}, nil
}

// GetPeer retrieves peer information by ID
func (s *PeerService) GetPeer(id string) (*peervault.PeerResponse, error) {
	peer, exists := s.peers[id]
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", id)
	}
	return peer, nil
}

// AddPeer adds a new peer
func (s *PeerService) AddPeer(address string, port int, metadata map[string]string) (*peervault.PeerResponse, error) {
	peerID := fmt.Sprintf("peer_%d", time.Now().Unix())

	peer := &peervault.PeerResponse{
		Id:        peerID,
		Address:   address,
		Port:      int32(port),
		Status:    "active",
		LastSeen:  timestamppb.Now(),
		CreatedAt: timestamppb.Now(),
		Metadata:  metadata,
	}

	s.peers[peerID] = peer
	return peer, nil
}

// RemovePeer removes a peer by ID
func (s *PeerService) RemovePeer(id string) (bool, error) {
	if _, exists := s.peers[id]; !exists {
		return false, fmt.Errorf("peer not found: %s", id)
	}

	delete(s.peers, id)
	return true, nil
}

// GetPeerHealth retrieves peer health information
func (s *PeerService) GetPeerHealth(id string) (*peervault.PeerHealthResponse, error) {
	_, exists := s.peers[id]
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", id)
	}

	return &peervault.PeerHealthResponse{
		PeerId:        id,
		Status:        "healthy",
		LatencyMs:     15.5,
		UptimeSeconds: 3600,
		Metrics: map[string]string{
			"cpu_usage":    "25%",
			"memory_usage": "60%",
			"disk_usage":   "45%",
		},
	}, nil
}

// StreamPeerEvents streams peer events
func (s *PeerService) StreamPeerEvents() (<-chan *peervault.PeerEvent, error) {
	// Mock implementation - return a channel that will be closed
	ch := make(chan *peervault.PeerEvent)
	close(ch)
	return ch, nil
}
