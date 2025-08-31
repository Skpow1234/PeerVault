package utils

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	fs "github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/logging"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

// TestServerManager manages multiple test servers for integration tests
type TestServerManager struct {
	servers map[string]*fs.Server
	mu      sync.RWMutex
}

// NewTestServerManager creates a new test server manager
func NewTestServerManager() *TestServerManager {
	return &TestServerManager{
		servers: make(map[string]*fs.Server),
	}
}

// CreateServer creates a test server with the given configuration
func (tsm *TestServerManager) CreateServer(name, listenAddr string, bootstrapNodes []string) *fs.Server {
	// Generate unique node ID
	nodeID := crypto.GenerateID()

	// Create transport
	tcptransportOpts := netp2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: netp2p.AuthenticatedHandshakeFunc(nodeID),
		Decoder:       netp2p.LengthPrefixedDecoder{},
	}
	tcpTransport := netp2p.NewTCPTransport(tcptransportOpts)

	// Create storage root with unique name to avoid conflicts
	storageRoot := storage.SanitizeStorageRootFromAddr(listenAddr) + "_" + name

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

	// Store server reference
	tsm.mu.Lock()
	tsm.servers[name] = s
	tsm.mu.Unlock()

	return s
}

// GetServer retrieves a server by name
func (tsm *TestServerManager) GetServer(name string) *fs.Server {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()
	return tsm.servers[name]
}

// StartAll starts all servers in parallel
func (tsm *TestServerManager) StartAll(t *testing.T) {
	tsm.mu.RLock()
	servers := make([]*fs.Server, 0, len(tsm.servers))
	for _, s := range tsm.servers {
		servers = append(servers, s)
	}
	tsm.mu.RUnlock()

	var wg sync.WaitGroup
	for _, server := range servers {
		wg.Add(1)
		go func(s *fs.Server) {
			defer wg.Done()
			if err := s.Start(); err != nil {
				t.Errorf("Failed to start server: %v", err)
			}
		}(server)
	}

	// Wait for all servers to start
	wg.Wait()

	// Give servers time to connect
	time.Sleep(2 * time.Second)
}

// StopAll stops all servers
func (tsm *TestServerManager) StopAll() {
	tsm.mu.RLock()
	servers := make([]*fs.Server, 0, len(tsm.servers))
	for _, s := range tsm.servers {
		servers = append(servers, s)
	}
	tsm.mu.RUnlock()

	for _, server := range servers {
		server.Stop()
	}
}

// CreateTestNetwork creates a network of test servers
func CreateTestNetwork(t *testing.T, config NetworkConfig) *TestServerManager {
	// Configure logging for tests
	logging.ConfigureLogger("info")

	manager := NewTestServerManager()

	// Create bootstrap nodes
	for _, node := range config.BootstrapNodes {
		manager.CreateServer(node.Name, node.ListenAddr, []string{})
	}

	// Create client nodes
	for _, node := range config.ClientNodes {
		manager.CreateServer(node.Name, node.ListenAddr, node.BootstrapNodes)
	}

	// Start all servers
	manager.StartAll(t)

	return manager
}

// NetworkConfig defines the configuration for a test network
type NetworkConfig struct {
	BootstrapNodes []NodeConfig
	ClientNodes    []NodeConfig
}

// NodeConfig defines the configuration for a single node
type NodeConfig struct {
	Name           string
	ListenAddr     string
	BootstrapNodes []string
}

// TestDataGenerator generates test data for various scenarios
type TestDataGenerator struct{}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{}
}

// GenerateSmallFile generates a small test file (1KB)
func (tdg *TestDataGenerator) GenerateSmallFile() []byte {
	return tdg.GenerateFile(1024)
}

// GenerateMediumFile generates a medium test file (100KB)
func (tdg *TestDataGenerator) GenerateMediumFile() []byte {
	return tdg.GenerateFile(100 * 1024)
}

// GenerateLargeFile generates a large test file (1MB)
func (tdg *TestDataGenerator) GenerateLargeFile() []byte {
	return tdg.GenerateFile(1024 * 1024)
}

// GenerateFile generates a file of the specified size
func (tdg *TestDataGenerator) GenerateFile(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// GenerateTextFile generates a text file with the given content
func (tdg *TestDataGenerator) GenerateTextFile(content string) []byte {
	return []byte(content)
}

// TestScenario defines a test scenario with setup and teardown
type TestScenario struct {
	Name     string
	Setup    func(*testing.T) *TestServerManager
	Teardown func(*testing.T, *TestServerManager)
	Test     func(*testing.T, *TestServerManager)
	Timeout  time.Duration
}

// RunTestScenario runs a test scenario with proper setup and teardown
func RunTestScenario(t *testing.T, scenario TestScenario) {
	t.Run(scenario.Name, func(t *testing.T) {
		// Set timeout if specified
		if scenario.Timeout > 0 {
			_, cancel := context.WithTimeout(context.Background(), scenario.Timeout)
			defer cancel()
			t.Logf("Running scenario with timeout: %v", scenario.Timeout)
		}

		// Setup
		var manager *TestServerManager
		if scenario.Setup != nil {
			manager = scenario.Setup(t)
		}

		// Ensure teardown runs
		defer func() {
			if scenario.Teardown != nil && manager != nil {
				scenario.Teardown(t, manager)
			}
		}()

		// Run test
		if scenario.Test != nil {
			scenario.Test(t, manager)
		}
	})
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, description string) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Errorf("Condition not met within timeout: %s", description)
	return false
}

// AssertNoError asserts that no error occurred
func AssertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}, message string) {
	if value == nil {
		t.Errorf("%s: value is nil", message)
	}
}

// LogTestEvent logs a test event with structured logging
func LogTestEvent(t *testing.T, event string, fields ...interface{}) {
	args := append([]interface{}{"event", event}, fields...)
	slog.Info("test event", args...)
	t.Logf("Test event: %s", event)
}
