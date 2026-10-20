package resolvers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/graphql/types"
	"github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
)

// Resolver is the main resolver interface for GraphQL operations
type Resolver interface {
	File(ctx context.Context, key string) (*types.File, error)
	Files(ctx context.Context, limit *int, offset *int, filter *types.FileFilter) ([]*types.File, error)
	FileMetadata(ctx context.Context, key string) (*types.FileMetadata, error)
	Node(ctx context.Context, id string) (*types.Node, error)
	Nodes(ctx context.Context) ([]*types.Node, error)
	PeerNetwork(ctx context.Context) (*types.PeerNetwork, error)
	SystemMetrics(ctx context.Context) (*types.SystemMetrics, error)
	PerformanceStats(ctx context.Context) (*types.PerformanceMetrics, error)
	StorageStats(ctx context.Context) (*types.StorageMetrics, error)
	Health(ctx context.Context) (*types.HealthStatus, error)

	UploadFile(ctx context.Context, file interface{}, key *string, metadata *types.FileMetadataInput) (*types.FileUpload, error)
	DeleteFile(ctx context.Context, key string) (bool, error)
	UpdateFileMetadata(ctx context.Context, key string, metadata *types.FileMetadataInput) (*types.FileMetadata, error)
	AddPeer(ctx context.Context, address string, port int) (*types.Node, error)
	RemovePeer(ctx context.Context, id string) (bool, error)
	UpdateConfiguration(ctx context.Context, config *types.ConfigurationInput) (bool, error)

	FileUploaded(ctx context.Context) (<-chan *types.File, error)
	FileDeleted(ctx context.Context) (<-chan string, error)
	FileUpdated(ctx context.Context) (<-chan *types.File, error)
	PeerConnected(ctx context.Context) (<-chan *types.Node, error)
	PeerDisconnected(ctx context.Context) (<-chan *types.Node, error)
	PeerHealthChanged(ctx context.Context) (<-chan *types.NodeHealth, error)
	SystemMetricsUpdated(ctx context.Context) (<-chan *types.SystemMetrics, error)
	PerformanceAlert(ctx context.Context) (<-chan *types.PerformanceAlert, error)
}

// BaseResolver provides fileserver-backed GraphQL resolvers.
type BaseResolver struct {
	server *fileserver.Server
}

// NewResolver creates a new GraphQL resolver
func NewResolver(server *fileserver.Server) Resolver {
	return &BaseResolver{server: server}
}

func (r *BaseResolver) requireServer() error {
	if r.server == nil {
		return fmt.Errorf("fileserver not configured")
	}
	return nil
}

func (r *BaseResolver) File(ctx context.Context, key string) (*types.File, error) {
	_ = ctx
	if err := r.requireServer(); err != nil {
		return nil, err
	}
	info, ok := r.server.GetFileInfo(key)
	if !ok || !r.server.Has(key) {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	return fileInfoToGraphQL(info, r.server), nil
}

func (r *BaseResolver) Files(ctx context.Context, limit *int, offset *int, filter *types.FileFilter) ([]*types.File, error) {
	_ = ctx
	if err := r.requireServer(); err != nil {
		return nil, err
	}
	all := r.server.ListFiles()
	out := make([]*types.File, 0, len(all))
	for i := range all {
		info := all[i]
		if filter != nil {
			if filter.SizeMin != nil && info.Size < *filter.SizeMin {
				continue
			}
			if filter.SizeMax != nil && info.Size > *filter.SizeMax {
				continue
			}
			if filter.CreatedAtAfter != nil && info.CreatedAt.Before(*filter.CreatedAtAfter) {
				continue
			}
			if filter.CreatedAtBefore != nil && info.CreatedAt.After(*filter.CreatedAtBefore) {
				continue
			}
		}
		out = append(out, fileInfoToGraphQL(&info, r.server))
	}
	start := 0
	if offset != nil && *offset > 0 {
		start = *offset
	}
	if start > len(out) {
		return []*types.File{}, nil
	}
	out = out[start:]
	if limit != nil && *limit >= 0 && *limit < len(out) {
		out = out[:*limit]
	}
	return out, nil
}

func (r *BaseResolver) FileMetadata(ctx context.Context, key string) (*types.FileMetadata, error) {
	f, err := r.File(ctx, key)
	if err != nil {
		return nil, err
	}
	return f.Metadata, nil
}

func (r *BaseResolver) Node(ctx context.Context, id string) (*types.Node, error) {
	_ = ctx
	if err := r.requireServer(); err != nil {
		return nil, err
	}
	for _, p := range r.server.ListPeers() {
		if p.ID == id || p.Address == id {
			return peerInfoToNode(p), nil
		}
	}
	if id == r.server.NodeID() {
		return selfNode(r.server), nil
	}
	return nil, fmt.Errorf("node not found: %s", id)
}

func (r *BaseResolver) Nodes(ctx context.Context) ([]*types.Node, error) {
	_ = ctx
	if err := r.requireServer(); err != nil {
		return nil, err
	}
	peers := r.server.ListPeers()
	out := make([]*types.Node, 0, len(peers))
	for _, p := range peers {
		out = append(out, peerInfoToNode(p))
	}
	return out, nil
}

func (r *BaseResolver) PeerNetwork(ctx context.Context) (*types.PeerNetwork, error) {
	nodes, err := r.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	connected := len(nodes)
	if connected > 0 {
		connected-- // exclude self from "connected remotes" display in topology
	}
	return &types.PeerNetwork{
		Nodes:       nodes,
		Connections: []*types.Connection{},
		Topology: &types.NetworkTopology{
			TotalNodes:     len(nodes),
			ConnectedNodes: connected,
		},
	}, nil
}

func (r *BaseResolver) SystemMetrics(ctx context.Context) (*types.SystemMetrics, error) {
	storage, err := r.StorageStats(ctx)
	if err != nil {
		return nil, err
	}
	perf, err := r.PerformanceStats(ctx)
	if err != nil {
		return nil, err
	}
	peers := 0
	uptime := 0.0
	if r.server != nil {
		peers = r.server.PeerCount()
		uptime = r.server.Uptime().Seconds()
	}
	return &types.SystemMetrics{
		Storage: storage,
		Network: &types.NetworkMetrics{
			ActiveConnections: peers,
			AverageBandwidth:  nil,
			ErrorRate:         floatPtr(0),
		},
		Performance: perf,
		Uptime:      uptime,
	}, nil
}

func (r *BaseResolver) PerformanceStats(ctx context.Context) (*types.PerformanceMetrics, error) {
	_ = ctx
	return &types.PerformanceMetrics{
		AverageResponseTime: floatPtr(0),
		RequestsPerSecond:   floatPtr(0),
		ErrorRate:           floatPtr(0),
	}, nil
}

func (r *BaseResolver) StorageStats(ctx context.Context) (*types.StorageMetrics, error) {
	_ = ctx
	if err := r.requireServer(); err != nil {
		return nil, err
	}
	files := r.server.ListFiles()
	var used int64
	for _, f := range files {
		used += f.Size
	}
	rf := 1.0 + float64(r.server.PeerCount())
	return &types.StorageMetrics{
		TotalSpace:        used,
		UsedSpace:         used,
		AvailableSpace:    0,
		FileCount:         len(files),
		ReplicationFactor: &rf,
	}, nil
}

func (r *BaseResolver) Health(ctx context.Context) (*types.HealthStatus, error) {
	_ = ctx
	if err := r.requireServer(); err != nil {
		return &types.HealthStatus{Status: "unhealthy", Timestamp: time.Now().UTC()}, err
	}
	return &types.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Details: map[string]interface{}{
			"fileserver": "up",
			"peers":      r.server.PeerCount(),
			"files":      r.server.FileCount(),
			"uptime":     r.server.Uptime().String(),
		},
	}, nil
}

func (r *BaseResolver) UploadFile(ctx context.Context, file interface{}, key *string, metadata *types.FileMetadataInput) (*types.FileUpload, error) {
	if err := r.requireServer(); err != nil {
		return nil, err
	}
	data, err := coerceUploadBytes(file)
	if err != nil {
		return nil, err
	}
	logicalKey := crypto.GenerateID()
	if key != nil && *key != "" {
		logicalKey = *key
	}
	contentType := "application/octet-stream"
	meta := map[string]string{}
	if metadata != nil {
		if metadata.ContentType != nil {
			contentType = *metadata.ContentType
		}
		if len(metadata.Tags) > 0 {
			meta["tags"] = strings.Join(metadata.Tags, ",")
		}
	}
	if err := r.server.StoreWithMeta(ctx, logicalKey, bytes.NewReader(data), contentType, meta); err != nil {
		return nil, err
	}
	info, _ := r.server.GetFileInfo(logicalKey)
	now := time.Now().UTC()
	progress := 1.0
	upload := &types.FileUpload{
		ID:         logicalKey,
		Key:        logicalKey,
		Size:       int64(len(data)),
		Status:     types.UploadStatusCompleted,
		Progress:   &progress,
		UploadedAt: &now,
	}
	if info != nil {
		upload.Size = info.Size
	}
	return upload, nil
}

func (r *BaseResolver) DeleteFile(ctx context.Context, key string) (bool, error) {
	if err := r.requireServer(); err != nil {
		return false, err
	}
	if err := r.server.Delete(ctx, key); err != nil {
		return false, err
	}
	return true, nil
}

func (r *BaseResolver) UpdateFileMetadata(ctx context.Context, key string, metadata *types.FileMetadataInput) (*types.FileMetadata, error) {
	_ = ctx
	if err := r.requireServer(); err != nil {
		return nil, err
	}
	meta := map[string]string{}
	contentType := ""
	if metadata != nil {
		if metadata.ContentType != nil {
			contentType = *metadata.ContentType
			meta["content_type"] = contentType
		}
		if len(metadata.Tags) > 0 {
			meta["tags"] = strings.Join(metadata.Tags, ",")
		}
	}
	info, err := r.server.UpdateFileMetadata(key, meta)
	if err != nil {
		return nil, err
	}
	if contentType == "" {
		contentType = info.ContentType
	}
	return &types.FileMetadata{
		ContentType:  contentType,
		Checksum:     info.Hash,
		Tags:         strings.Split(meta["tags"], ","),
		CustomFields: map[string]interface{}{},
	}, nil
}

func (r *BaseResolver) AddPeer(ctx context.Context, address string, port int) (*types.Node, error) {
	if err := r.requireServer(); err != nil {
		return nil, err
	}
	info, err := r.server.AddPeer(ctx, address, port)
	if err != nil {
		return nil, err
	}
	return peerInfoToNode(*info), nil
}

func (r *BaseResolver) RemovePeer(ctx context.Context, id string) (bool, error) {
	_ = ctx
	if err := r.requireServer(); err != nil {
		return false, err
	}
	if err := r.server.RemovePeer(id); err != nil {
		return false, err
	}
	return true, nil
}

func (r *BaseResolver) UpdateConfiguration(ctx context.Context, config *types.ConfigurationInput) (bool, error) {
	_ = ctx
	_ = config
	// Runtime reconfiguration is not supported yet; acknowledge without error for API compatibility.
	return true, nil
}

func (r *BaseResolver) FileUploaded(ctx context.Context) (<-chan *types.File, error) {
	return nil, fmt.Errorf("subscription resolver not configured")
}
func (r *BaseResolver) FileDeleted(ctx context.Context) (<-chan string, error) {
	return nil, fmt.Errorf("subscription resolver not configured")
}
func (r *BaseResolver) FileUpdated(ctx context.Context) (<-chan *types.File, error) {
	return nil, fmt.Errorf("subscription resolver not configured")
}
func (r *BaseResolver) PeerConnected(ctx context.Context) (<-chan *types.Node, error) {
	return nil, fmt.Errorf("subscription resolver not configured")
}
func (r *BaseResolver) PeerDisconnected(ctx context.Context) (<-chan *types.Node, error) {
	return nil, fmt.Errorf("subscription resolver not configured")
}
func (r *BaseResolver) PeerHealthChanged(ctx context.Context) (<-chan *types.NodeHealth, error) {
	return nil, fmt.Errorf("subscription resolver not configured")
}
func (r *BaseResolver) SystemMetricsUpdated(ctx context.Context) (<-chan *types.SystemMetrics, error) {
	return nil, fmt.Errorf("subscription resolver not configured")
}
func (r *BaseResolver) PerformanceAlert(ctx context.Context) (<-chan *types.PerformanceAlert, error) {
	return nil, fmt.Errorf("subscription resolver not configured")
}

func fileInfoToGraphQL(info *fileserver.FileInfo, server *fileserver.Server) *types.File {
	owner := selfNode(server)
	return &types.File{
		ID:        info.Key,
		Key:       info.Key,
		HashedKey: crypto.HashKey(info.Key),
		Size:      info.Size,
		CreatedAt: info.CreatedAt,
		UpdatedAt: info.UpdatedAt,
		Owner:     owner,
		Replicas: []*types.FileReplica{{
			Node:   owner,
			Status: types.ReplicaStatusSynced,
			Size:   &info.Size,
		}},
		Metadata: &types.FileMetadata{
			ContentType:  info.ContentType,
			Checksum:     info.Hash,
			Tags:         nil,
			CustomFields: map[string]interface{}{},
		},
	}
}

func selfNode(server *fileserver.Server) *types.Node {
	now := time.Now().UTC()
	host, portStr, _ := net.SplitHostPort(normalizeAddr(server.ListenAddr()))
	port, _ := strconv.Atoi(portStr)
	return &types.Node{
		ID:       server.NodeID(),
		Address:  host,
		Port:     port,
		Status:   types.NodeStatusOnline,
		LastSeen: &now,
		Health: &types.NodeHealth{
			IsHealthy:     true,
			LastHeartbeat: &now,
		},
		Capabilities: []string{"store", "get", "replicate"},
	}
}

func peerInfoToNode(p fileserver.PeerInfo) *types.Node {
	host, portStr, err := net.SplitHostPort(normalizeAddr(p.Address))
	port := 0
	if err == nil {
		port, _ = strconv.Atoi(portStr)
	} else {
		host = p.Address
	}
	last := p.LastSeen
	status := types.NodeStatusOnline
	if p.Status == "self" {
		status = types.NodeStatusOnline
	}
	return &types.Node{
		ID:       p.ID,
		Address:  host,
		Port:     port,
		Status:   status,
		LastSeen: &last,
		Health: &types.NodeHealth{
			IsHealthy:     true,
			LastHeartbeat: &last,
		},
	}
}

func normalizeAddr(addr string) string {
	if addr == "" {
		return "127.0.0.1:0"
	}
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr
}

func coerceUploadBytes(file interface{}) ([]byte, error) {
	switch v := file.(type) {
	case nil:
		return nil, fmt.Errorf("file is required")
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case io.Reader:
		return io.ReadAll(v)
	default:
		return nil, fmt.Errorf("unsupported upload type %T", file)
	}
}

func floatPtr(v float64) *float64 { return &v }
