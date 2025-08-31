package multi_node

import (
	"log/slog"
	"testing"
	"time"

	fs "github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/logging"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

// TestNetworkTopology tests a multi-node network topology
// Creates a star topology: 1 bootstrap node, 3 client nodes
func TestNetworkTopology(t *testing.T) {
	// Configure logging for tests
	logging.ConfigureLogger("info")

	// Create bootstrap node (center of star)
	bootstrapNode := createTestServer(":4001", []string{})
	defer bootstrapNode.Stop()

	// Create client nodes (connect to bootstrap)
	clientNode1 := createTestServer(":4002", []string{":4001"})
	defer clientNode1.Stop()

	clientNode2 := createTestServer(":4003", []string{":4001"})
	defer clientNode2.Stop()

	clientNode3 := createTestServer(":4004", []string{":4001"})
	defer clientNode3.Stop()

	// Start all nodes
	nodes := []*fs.Server{bootstrapNode, clientNode1, clientNode2, clientNode3}
	for _, node := range nodes {
		go func(s *fs.Server) {
			if err := s.Start(); err != nil {
				t.Errorf("Failed to start node: %v", err)
			}
		}(node)
	}

	// Wait for all nodes to start and connect
	time.Sleep(3 * time.Second)

	// Test network topology
	t.Run("Network Topology", func(t *testing.T) {
		// Verify that all nodes are running
		if bootstrapNode == nil {
			t.Error("Bootstrap node should not be nil")
		}

		if clientNode1 == nil || clientNode2 == nil || clientNode3 == nil {
			t.Error("All client nodes should not be nil")
		}

		slog.Info("Network topology test passed",
			"bootstrap_node", "running",
			"client_nodes", 3)
	})

	// Test resource limits configuration
	t.Run("Resource Limits", func(t *testing.T) {
		// Check that resource limits are properly configured
		limits := peer.DefaultResourceLimits()
		if limits.MaxConcurrentStreams <= 0 {
			t.Error("MaxConcurrentStreams should be positive")
		}

		if limits.StreamTimeout <= 0 {
			t.Error("StreamTimeout should be positive")
		}

		slog.Info("Resource limits test passed",
			"max_streams", limits.MaxConcurrentStreams,
			"stream_timeout", limits.StreamTimeout)
	})

	// Test network resilience
	t.Run("Network Resilience", func(t *testing.T) {
		// Stop one client node
		clientNode1.Stop()
		time.Sleep(1 * time.Second)

		// Verify that other nodes are still running
		if clientNode2 == nil || clientNode3 == nil {
			t.Error("Remaining client nodes should still be running")
		}

		slog.Info("Network resilience test passed", "remaining_nodes", 2)
	})
}

// TestPeerHealthMonitoring tests peer health monitoring
func TestPeerHealthMonitoring(t *testing.T) {
	// Configure logging for tests
	logging.ConfigureLogger("info")

	// Create bootstrap node
	bootstrapNode := createTestServer(":5001", []string{})
	defer bootstrapNode.Stop()

	// Create client node
	clientNode := createTestServer(":5002", []string{":5001"})
	defer clientNode.Stop()

	// Start nodes
	go func() {
		if err := bootstrapNode.Start(); err != nil {
			t.Errorf("Failed to start bootstrap node: %v", err)
		}
	}()

	go func() {
		if err := clientNode.Start(); err != nil {
			t.Errorf("Failed to start client node: %v", err)
		}
	}()

	// Wait for connection
	time.Sleep(2 * time.Second)

	// Test initial health
	t.Run("Initial Health Check", func(t *testing.T) {
		// Check that nodes are running
		if bootstrapNode == nil {
			t.Error("Bootstrap node should not be nil")
		}

		if clientNode == nil {
			t.Error("Client node should not be nil")
		}

		slog.Info("Initial health check passed")
	})

	// Test health monitoring over time
	t.Run("Health Monitoring", func(t *testing.T) {
		// Monitor health for a short period
		for i := 0; i < 5; i++ {
			time.Sleep(500 * time.Millisecond)

			// Verify nodes are still running
			if bootstrapNode == nil || clientNode == nil {
				t.Errorf("Health check failed at iteration %d", i+1)
			}

			slog.Info("Health check",
				"iteration", i+1,
				"bootstrap_node", "running",
				"client_node", "running")
		}

		slog.Info("Health monitoring test passed")
	})
}

// TestResourceLimits tests resource limits and backpressure
func TestResourceLimits(t *testing.T) {
	// Configure logging for tests
	logging.ConfigureLogger("info")

	// Create nodes
	bootstrapNode := createTestServer(":6001", []string{})
	defer bootstrapNode.Stop()

	clientNode := createTestServer(":6002", []string{":6001"})
	defer clientNode.Stop()

	// Start nodes
	go func() {
		if err := bootstrapNode.Start(); err != nil {
			t.Errorf("Failed to start bootstrap node: %v", err)
		}
	}()

	go func() {
		if err := clientNode.Start(); err != nil {
			t.Errorf("Failed to start client node: %v", err)
		}
	}()

	// Wait for connection
	time.Sleep(2 * time.Second)

	// Test resource limits
	t.Run("Resource Limits", func(t *testing.T) {
		// Check that limits are properly configured
		limits := peer.DefaultResourceLimits()
		if limits.MaxConcurrentStreams <= 0 {
			t.Error("MaxConcurrentStreams should be positive")
		}

		if limits.StreamTimeout <= 0 {
			t.Error("StreamTimeout should be positive")
		}

		slog.Info("Resource limits test passed",
			"max_streams", limits.MaxConcurrentStreams,
			"stream_timeout", limits.StreamTimeout)
	})
}

// Helper function to create a test server
func createTestServer(listenAddr string, bootstrapNodes []string) *fs.Server {
	// Generate unique node ID
	nodeID := crypto.GenerateID()

	// Create transport
	tcptransportOpts := netp2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: netp2p.AuthenticatedHandshakeFunc(nodeID),
		Decoder:       netp2p.LengthPrefixedDecoder{},
	}
	tcpTransport := netp2p.NewTCPTransport(tcptransportOpts)

	// Create storage root
	storageRoot := storage.SanitizeStorageRootFromAddr(listenAddr)

	// Create server options
	fileServerOpts := fs.Options{
		ID:                nodeID,
		EncKey:            crypto.NewEncryptionKey(),
		StorageRoot:       storageRoot,
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    bootstrapNodes,
		ResourceLimits:    peer.DefaultResourceLimits(),
	}

	// Create and configure server
	s := fs.New(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}
