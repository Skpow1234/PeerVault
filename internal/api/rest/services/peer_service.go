package services

import (
	"context"

	"github.com/Skpow1234/Peervault/internal/api/rest/types"
)

type PeerService interface {
	ListPeers(ctx context.Context) ([]types.Peer, error)
	GetPeer(ctx context.Context, peerID string) (*types.Peer, error)
	AddPeer(ctx context.Context, peer types.Peer) (*types.Peer, error)
	RemovePeer(ctx context.Context, peerID string) error
}
