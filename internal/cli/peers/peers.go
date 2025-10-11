package peers

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// Manager manages peer operations and analytics
type Manager struct {
	client    *client.Client
	formatter *formatter.Formatter
	peers     map[string]*PeerInfo
	mu        sync.RWMutex
}

// PeerInfo represents extended peer information
type PeerInfo struct {
	*client.PeerInfo
	LatencyHistory []time.Duration
	Uptime         time.Duration
	LastError      error
	Performance    *PerformanceMetrics
}

// PerformanceMetrics represents peer performance metrics
type PerformanceMetrics struct {
	AvgLatency   time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
	SuccessRate  float64
	RequestCount int64
	ErrorCount   int64
	LastUpdated  time.Time
}

// NetworkTopology represents the network topology
type NetworkTopology struct {
	Nodes    []*TopologyNode
	Edges    []*TopologyEdge
	Clusters []*Cluster
}

// TopologyNode represents a node in the topology
type TopologyNode struct {
	ID       string
	Address  string
	Status   string
	Position Position
	Metrics  *PerformanceMetrics
}

// TopologyEdge represents a connection between nodes
type TopologyEdge struct {
	From     string
	To       string
	Latency  time.Duration
	Strength float64
}

// Cluster represents a cluster of peers
type Cluster struct {
	ID      string
	Peers   []string
	Center  Position
	Radius  float64
	Metrics *ClusterMetrics
}

// ClusterMetrics represents cluster-level metrics
type ClusterMetrics struct {
	AvgLatency   time.Duration
	TotalStorage int64
	ActivePeers  int
	HealthScore  float64
}

// Position represents a 2D position
type Position struct {
	X float64
	Y float64
}

// New creates a new peer manager
func New(client *client.Client, formatter *formatter.Formatter) *Manager {
	return &Manager{
		client:    client,
		formatter: formatter,
		peers:     make(map[string]*PeerInfo),
	}
}

// DiscoverPeers discovers peers in the network
func (m *Manager) DiscoverPeers(ctx context.Context) error {
	m.formatter.PrintInfo("Discovering peers in the network...")

	// Get current peers
	peers, err := m.client.ListPeers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list peers: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update peer information
	for _, peer := range peers.Peers {
		if existing, exists := m.peers[peer.ID]; exists {
			// Update existing peer
			existing.PeerInfo = &peer
			existing.LatencyHistory = append(existing.LatencyHistory, time.Duration(peer.Latency)*time.Microsecond)

			// Keep only last 100 latency measurements
			if len(existing.LatencyHistory) > 100 {
				existing.LatencyHistory = existing.LatencyHistory[len(existing.LatencyHistory)-100:]
			}

			existing.Performance = m.calculatePerformanceMetrics(existing.LatencyHistory)
		} else {
			// Add new peer
			m.peers[peer.ID] = &PeerInfo{
				PeerInfo:       &peer,
				LatencyHistory: []time.Duration{time.Duration(peer.Latency) * time.Microsecond},
				Performance:    m.calculatePerformanceMetrics([]time.Duration{time.Duration(peer.Latency) * time.Microsecond}),
			}
		}
	}

	m.formatter.PrintSuccess(fmt.Sprintf("Discovered %d peers", len(peers.Peers)))
	return nil
}

// GetTopology returns the network topology
func (m *Manager) GetTopology(ctx context.Context) (*NetworkTopology, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	topology := &NetworkTopology{
		Nodes:    make([]*TopologyNode, 0, len(m.peers)),
		Edges:    make([]*TopologyEdge, 0),
		Clusters: make([]*Cluster, 0),
	}

	// Create nodes
	for _, peer := range m.peers {
		node := &TopologyNode{
			ID:       peer.ID,
			Address:  peer.Address,
			Status:   peer.Status,
			Position: m.calculatePosition(peer),
			Metrics:  peer.Performance,
		}
		topology.Nodes = append(topology.Nodes, node)
	}

	// Create edges (simplified - connect all peers)
	for i, node1 := range topology.Nodes {
		for j, node2 := range topology.Nodes {
			if i != j {
				edge := &TopologyEdge{
					From:     node1.ID,
					To:       node2.ID,
					Latency:  m.calculateLatency(node1.ID, node2.ID),
					Strength: m.calculateConnectionStrength(node1.ID, node2.ID),
				}
				topology.Edges = append(topology.Edges, edge)
			}
		}
	}

	// Create clusters
	topology.Clusters = m.identifyClusters(topology.Nodes)

	return topology, nil
}

// GetPerformanceAnalytics returns performance analytics for peers
func (m *Manager) GetPerformanceAnalytics(ctx context.Context) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	analytics := map[string]interface{}{
		"total_peers":      len(m.peers),
		"healthy_peers":    0,
		"degraded_peers":   0,
		"unhealthy_peers":  0,
		"avg_latency":      time.Duration(0),
		"min_latency":      time.Duration(math.MaxInt64),
		"max_latency":      time.Duration(0),
		"top_performers":   make([]*PeerInfo, 0),
		"worst_performers": make([]*PeerInfo, 0),
	}

	var totalLatency time.Duration
	peerList := make([]*PeerInfo, 0, len(m.peers))

	for _, peer := range m.peers {
		peerList = append(peerList, peer)

		// Count by status
		switch peer.Status {
		case "healthy":
			analytics["healthy_peers"] = analytics["healthy_peers"].(int) + 1
		case "degraded":
			analytics["degraded_peers"] = analytics["degraded_peers"].(int) + 1
		default:
			analytics["unhealthy_peers"] = analytics["unhealthy_peers"].(int) + 1
		}

		// Calculate latency statistics
		if peer.Performance != nil {
			totalLatency += peer.Performance.AvgLatency

			if peer.Performance.MinLatency < analytics["min_latency"].(time.Duration) {
				analytics["min_latency"] = peer.Performance.MinLatency
			}

			if peer.Performance.MaxLatency > analytics["max_latency"].(time.Duration) {
				analytics["max_latency"] = peer.Performance.MaxLatency
			}
		}
	}

	if len(peerList) > 0 {
		analytics["avg_latency"] = totalLatency / time.Duration(len(peerList))
	}

	// Sort peers by performance
	sort.Slice(peerList, func(i, j int) bool {
		if peerList[i].Performance == nil || peerList[j].Performance == nil {
			return false
		}
		return peerList[i].Performance.SuccessRate > peerList[j].Performance.SuccessRate
	})

	// Get top and worst performers
	if len(peerList) >= 3 {
		analytics["top_performers"] = peerList[:3]
		analytics["worst_performers"] = peerList[len(peerList)-3:]
	}

	return analytics, nil
}

// MonitorPeers continuously monitors peer health
func (m *Manager) MonitorPeers(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.formatter.PrintInfo("Monitoring peer health...")

			if err := m.DiscoverPeers(ctx); err != nil {
				m.formatter.PrintError(fmt.Errorf("peer discovery failed: %w", err))
			}

			// Check for unhealthy peers
			m.mu.RLock()
			for _, peer := range m.peers {
				if peer.Status != "healthy" {
					m.formatter.PrintWarning(fmt.Sprintf("Peer %s is %s", peer.Address, peer.Status))
				}
			}
			m.mu.RUnlock()
		}
	}
}

// calculatePerformanceMetrics calculates performance metrics from latency history
func (m *Manager) calculatePerformanceMetrics(latencyHistory []time.Duration) *PerformanceMetrics {
	if len(latencyHistory) == 0 {
		return &PerformanceMetrics{}
	}

	var total time.Duration
	min := latencyHistory[0]
	max := latencyHistory[0]

	for _, latency := range latencyHistory {
		total += latency
		if latency < min {
			min = latency
		}
		if latency > max {
			max = latency
		}
	}

	return &PerformanceMetrics{
		AvgLatency:  total / time.Duration(len(latencyHistory)),
		MinLatency:  min,
		MaxLatency:  max,
		SuccessRate: 0.95, // Simplified - would be calculated from actual success/failure data
		LastUpdated: time.Now(),
	}
}

// calculatePosition calculates a position for a peer (simplified)
func (m *Manager) calculatePosition(peer *PeerInfo) Position {
	// Simplified positioning based on peer ID hash
	hash := 0
	for _, char := range peer.ID {
		hash += int(char)
	}

	return Position{
		X: float64(hash % 100),
		Y: float64((hash / 100) % 100),
	}
}

// calculateLatency calculates latency between two peers
func (m *Manager) calculateLatency(peer1ID, peer2ID string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peer1, exists1 := m.peers[peer1ID]
	peer2, exists2 := m.peers[peer2ID]

	if !exists1 || !exists2 || peer1.Performance == nil || peer2.Performance == nil {
		return time.Duration(0)
	}

	// Simplified calculation - average of both peers' average latency
	return (peer1.Performance.AvgLatency + peer2.Performance.AvgLatency) / 2
}

// calculateConnectionStrength calculates connection strength between peers
func (m *Manager) calculateConnectionStrength(peer1ID, peer2ID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	peer1, exists1 := m.peers[peer1ID]
	peer2, exists2 := m.peers[peer2ID]

	if !exists1 || !exists2 {
		return 0.0
	}

	// Simplified calculation based on status and performance
	strength := 1.0

	if peer1.Status != "healthy" || peer2.Status != "healthy" {
		strength *= 0.5
	}

	if peer1.Performance != nil && peer2.Performance != nil {
		avgSuccessRate := (peer1.Performance.SuccessRate + peer2.Performance.SuccessRate) / 2
		strength *= avgSuccessRate
	}

	return strength
}

// identifyClusters identifies clusters of peers
func (m *Manager) identifyClusters(nodes []*TopologyNode) []*Cluster {
	// Simplified clustering - group peers by proximity
	clusters := make([]*Cluster, 0)
	visited := make(map[string]bool)

	for _, node := range nodes {
		if visited[node.ID] {
			continue
		}

		cluster := &Cluster{
			ID:      fmt.Sprintf("cluster-%d", len(clusters)),
			Peers:   []string{node.ID},
			Center:  node.Position,
			Radius:  50.0, // Simplified
			Metrics: &ClusterMetrics{},
		}

		// Find nearby peers
		for _, otherNode := range nodes {
			if otherNode.ID == node.ID || visited[otherNode.ID] {
				continue
			}

			dx := node.Position.X - otherNode.Position.X
			dy := node.Position.Y - otherNode.Position.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance <= cluster.Radius {
				cluster.Peers = append(cluster.Peers, otherNode.ID)
				visited[otherNode.ID] = true
			}
		}

		visited[node.ID] = true
		clusters = append(clusters, cluster)
	}

	return clusters
}
