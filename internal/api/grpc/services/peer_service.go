package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/grpc/types"
)

type PeerService struct {
	logger *slog.Logger
	// TODO: Integrate with actual peer management
	peers map[string]*MockPeer
}

type MockPeer struct {
	ID        string
	Address   string
	Port      int32
	Status    string
	LastSeen  time.Time
	CreatedAt time.Time
	Metadata  map[string]string
}

func NewPeerService(logger *slog.Logger) *PeerService {
	// Initialize with mock data
	peers := map[string]*MockPeer{
		"peer1": {
			ID:        "peer1",
			Address:   "192.168.1.100",
			Port:      8080,
			Status:    "active",
			LastSeen:  time.Now(),
			CreatedAt: time.Now().Add(-time.Hour),
			Metadata:  map[string]string{"region": "us-east"},
		},
	}

	return &PeerService{
		logger: logger,
		peers:  peers,
	}
}

func (s *PeerService) ListPeers(ctx context.Context) (*types.ListPeersResponse, error) {
	var peerResponses []types.PeerResponse

	for _, peer := range s.peers {
		response := types.PeerResponse{
			ID:        peer.ID,
			Address:   peer.Address,
			Port:      peer.Port,
			Status:    peer.Status,
			LastSeen:  peer.LastSeen,
			CreatedAt: peer.CreatedAt,
			Metadata:  peer.Metadata,
		}
		peerResponses = append(peerResponses, response)
	}

	return &types.ListPeersResponse{
		Peers: peerResponses,
		Total: int32(len(peerResponses)),
	}, nil
}

func (s *PeerService) GetPeer(ctx context.Context, req *types.PeerRequest) (*types.PeerResponse, error) {
	peer, exists := s.peers[req.ID]
	if !exists {
		return nil, nil // Return nil for not found
	}

	return &types.PeerResponse{
		ID:        peer.ID,
		Address:   peer.Address,
		Port:      peer.Port,
		Status:    peer.Status,
		LastSeen:  peer.LastSeen,
		CreatedAt: peer.CreatedAt,
		Metadata:  peer.Metadata,
	}, nil
}

func (s *PeerService) AddPeer(ctx context.Context, req *types.AddPeerRequest) (*types.PeerResponse, error) {
	peerID := "peer" + string(rune(len(s.peers)+1))

	peer := &MockPeer{
		ID:        peerID,
		Address:   req.Address,
		Port:      req.Port,
		Status:    "active",
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
		Metadata:  req.Metadata,
	}

	s.peers[peerID] = peer

	s.logger.Info("Peer added via gRPC", "id", peerID, "address", req.Address)

	return &types.PeerResponse{
		ID:        peer.ID,
		Address:   peer.Address,
		Port:      peer.Port,
		Status:    peer.Status,
		LastSeen:  peer.LastSeen,
		CreatedAt: peer.CreatedAt,
		Metadata:  peer.Metadata,
	}, nil
}

func (s *PeerService) RemovePeer(ctx context.Context, req *types.PeerRequest) (*types.RemovePeerResponse, error) {
	_, exists := s.peers[req.ID]
	if !exists {
		return &types.RemovePeerResponse{
			Success: false,
			Message: "Peer not found",
		}, nil
	}

	delete(s.peers, req.ID)

	s.logger.Info("Peer removed via gRPC", "id", req.ID)

	return &types.RemovePeerResponse{
		Success: true,
		Message: "Peer removed successfully",
	}, nil
}

func (s *PeerService) GetPeerHealth(ctx context.Context, req *types.PeerRequest) (*types.PeerHealthResponse, error) {
	peer, exists := s.peers[req.ID]
	if !exists {
		return nil, nil // Return nil for not found
	}

	return &types.PeerHealthResponse{
		PeerID:        peer.ID,
		Status:        peer.Status,
		LatencyMs:     10.5,
		UptimeSeconds: int32(time.Since(peer.CreatedAt).Seconds()),
		Metrics:       map[string]string{"connections": "5"},
	}, nil
}

func (s *PeerService) StreamPeerEvents(stream interface{}) error {
	// TODO: Implement real-time peer event streaming
	s.logger.Info("Peer event streaming not yet implemented")
	return nil
}
