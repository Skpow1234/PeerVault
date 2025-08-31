package end_to_end

import (
	"bytes"
	"context"
	"io"
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

// TestBasicStoreGet tests the basic store and get operations
func TestBasicStoreGet(t *testing.T) {
	// Configure logging for tests
	logging.ConfigureLogger("info")

	// Create test data
	testData := []byte("Hello, PeerVault! This is a test file for end-to-end testing.")
	testKey := "test_file_001.txt"

	// Create server 1 (bootstrap node)
	server1 := createTestServer(":3001", []string{})
	defer server1.Stop()

	// Create server 2 (connects to server 1)
	server2 := createTestServer(":3002", []string{":3001"})
	defer server2.Stop()

	// Start both servers
	go func() {
		if err := server1.Start(); err != nil {
			t.Errorf("Failed to start server1: %v", err)
		}
	}()

	go func() {
		if err := server2.Start(); err != nil {
			t.Errorf("Failed to start server2: %v", err)
		}
	}()

	// Wait for servers to start and connect
	time.Sleep(2 * time.Second)

	// Add timeout to prevent infinite hanging
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()

	// Use the test context with timeout
	ctx, cancel := context.WithTimeout(testCtx, 30*time.Second)
	defer cancel()

	// Test 1: Store file on server1
	t.Run("Store on server1", func(t *testing.T) {
		slog.Info("Storing file on server1", "key", testKey)

		err := server1.Store(ctx, testKey, bytes.NewReader(testData))
		if err != nil {
			t.Fatalf("Failed to store file on server1: %v", err)
		}

		slog.Info("File stored successfully on server1")
	})

	// Wait for replication
	time.Sleep(1 * time.Second)

	// Test 2: Get file from server2 (should be replicated)
	t.Run("Get from server2", func(t *testing.T) {
		slog.Info("Retrieving file from server2", "key", testKey)

		reader, err := server2.Get(ctx, testKey)
		if err != nil {
			t.Fatalf("Failed to get file from server2: %v", err)
		}
		if closer, ok := reader.(io.Closer); ok {
			defer func() {
				if err := closer.Close(); err != nil {
					t.Logf("Failed to close reader: %v", err)
				}
			}()
		}

		// Read the data
		retrievedData, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read retrieved data: %v", err)
		}

		// Verify the data matches
		if !bytes.Equal(testData, retrievedData) {
			t.Errorf("Data mismatch: expected %q, got %q", string(testData), string(retrievedData))
		}

		slog.Info("File retrieved successfully from server2",
			"expected_size", len(testData),
			"actual_size", len(retrievedData))
	})

	// Test 3: Get file from server1 (should be local)
	t.Run("Get from server1", func(t *testing.T) {
		slog.Info("Retrieving file from server1", "key", testKey)

		reader, err := server1.Get(ctx, testKey)
		if err != nil {
			t.Fatalf("Failed to get file from server1: %v", err)
		}
		if closer, ok := reader.(io.Closer); ok {
			defer func() {
				if err := closer.Close(); err != nil {
					t.Logf("Failed to close reader: %v", err)
				}
			}()
		}

		// Read the data
		retrievedData, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read retrieved data: %v", err)
		}

		// Verify the data matches
		if !bytes.Equal(testData, retrievedData) {
			t.Errorf("Data mismatch: expected %q, got %q", string(testData), string(retrievedData))
		}

		slog.Info("File retrieved successfully from server1")
	})
}

// TestLargeFileStreaming tests streaming of large files
func TestLargeFileStreaming(t *testing.T) {
	// Configure logging for tests
	logging.ConfigureLogger("info")

	// Create large test data (1MB)
	testData := make([]byte, 1024*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	testKey := "large_test_file.bin"

	// Create servers
	server1 := createTestServer(":3003", []string{})
	defer server1.Stop()

	server2 := createTestServer(":3004", []string{":3003"})
	defer server2.Stop()

	// Start servers
	go func() {
		if err := server1.Start(); err != nil {
			t.Errorf("Failed to start server1: %v", err)
		}
	}()

	go func() {
		if err := server2.Start(); err != nil {
			t.Errorf("Failed to start server2: %v", err)
		}
	}()

	// Wait for servers to start
	time.Sleep(2 * time.Second)

	// Add timeout to prevent infinite hanging
	testCtx, testCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer testCancel()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(testCtx, 60*time.Second)
	defer cancel()

	// Store large file
	t.Run("Store large file", func(t *testing.T) {
		slog.Info("Storing large file", "key", testKey, "size", len(testData))

		start := time.Now()
		err := server1.Store(ctx, testKey, bytes.NewReader(testData))
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to store large file: %v", err)
		}

		slog.Info("Large file stored successfully", "duration", duration)
	})

	// Wait for replication
	time.Sleep(2 * time.Second)

	// Retrieve large file
	t.Run("Get large file", func(t *testing.T) {
		slog.Info("Retrieving large file", "key", testKey)

		start := time.Now()
		reader, err := server2.Get(ctx, testKey)
		if err != nil {
			t.Fatalf("Failed to get large file: %v", err)
		}
		if closer, ok := reader.(io.Closer); ok {
			defer closer.Close()
		}

		retrievedData, err := io.ReadAll(reader)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to read large file: %v", err)
		}

		if !bytes.Equal(testData, retrievedData) {
			t.Errorf("Large file data mismatch: expected %d bytes, got %d bytes",
				len(testData), len(retrievedData))
		}

		slog.Info("Large file retrieved successfully",
			"duration", duration,
			"size", len(retrievedData))
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
