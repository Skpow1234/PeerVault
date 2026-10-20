package implementations

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Skpow1234/Peervault/internal/api/rest/services"
	"github.com/Skpow1234/Peervault/internal/api/rest/types"
	"github.com/Skpow1234/Peervault/internal/app/fileserver"
)

type PeerServiceImpl struct {
	server *fileserver.Server
}

func NewPeerService(server *fileserver.Server) services.PeerService {
	return &PeerServiceImpl{server: server}
}

func (s *PeerServiceImpl) ListPeers(ctx context.Context) ([]types.Peer, error) {
	_ = ctx
	if s.server == nil {
		return []types.Peer{}, nil
	}
	infos := s.server.ListPeers()
	out := make([]types.Peer, 0, len(infos))
	for _, info := range infos {
		out = append(out, peerInfoToREST(info))
	}
	return out, nil
}

func (s *PeerServiceImpl) GetPeer(ctx context.Context, peerID string) (*types.Peer, error) {
	_ = ctx
	if s.server == nil {
		return nil, fmt.Errorf("fileserver not configured")
	}
	for _, info := range s.server.ListPeers() {
		if info.ID == peerID || info.Address == peerID {
			p := peerInfoToREST(info)
			return &p, nil
		}
	}
	return nil, fmt.Errorf("peer not found: %s", peerID)
}

func (s *PeerServiceImpl) AddPeer(ctx context.Context, peer types.Peer) (*types.Peer, error) {
	if s.server == nil {
		return nil, fmt.Errorf("fileserver not configured")
	}
	if peer.Address == "" {
		return nil, fmt.Errorf("peer address is required")
	}
	info, err := s.server.AddPeer(ctx, peer.Address, peer.Port)
	if err != nil {
		return nil, err
	}
	p := peerInfoToREST(*info)
	return &p, nil
}

func (s *PeerServiceImpl) RemovePeer(ctx context.Context, peerID string) error {
	_ = ctx
	if s.server == nil {
		return fmt.Errorf("fileserver not configured")
	}
	return s.server.RemovePeer(peerID)
}

func peerInfoToREST(info fileserver.PeerInfo) types.Peer {
	host, portStr, err := net.SplitHostPort(normalizeListen(info.Address))
	port := 0
	if err == nil {
		port, _ = strconv.Atoi(portStr)
	} else {
		host = info.Address
	}
	status := info.Status
	if status == "" {
		status = "active"
	}
	return types.Peer{
		ID:        info.ID,
		Address:   host,
		Port:      port,
		Status:    status,
		LastSeen:  info.LastSeen,
		CreatedAt: info.LastSeen,
		Metadata:  map[string]string{},
	}
}

func normalizeListen(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	if addr == "" {
		return "127.0.0.1:0"
	}
	return addr
}
