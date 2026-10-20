package end_to_end

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	fs "github.com/Skpow1234/Shugo/internal/app/fileserver"
	"github.com/Skpow1234/Shugo/internal/crypto"
	"github.com/Skpow1234/Shugo/internal/logging"
	"github.com/Skpow1234/Shugo/internal/peer"
	"github.com/Skpow1234/Shugo/internal/storage"
	netp2p "github.com/Skpow1234/Shugo/internal/transport/p2p"
)

// freePort asks the OS for an available TCP port.
func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func startNode(t *testing.T, listen string, bootstrap []string, km *crypto.KeyManager) *fs.Server {
	t.Helper()
	nodeID := crypto.GenerateID()
	tr := netp2p.NewTCPTransport(netp2p.TCPTransportOpts{
		ListenAddr:    listen,
		HandshakeFunc: netp2p.AuthenticatedHandshakeFunc(nodeID),
		Decoder:       netp2p.LengthPrefixedDecoder{},
	})
	s := fs.New(fs.Options{
		ID:                nodeID,
		KeyManager:        km,
		StorageRoot:       t.TempDir(),
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tr,
		BootstrapNodes:    bootstrap,
		ResourceLimits:    peer.DefaultResourceLimits(),
	})
	tr.OnPeer = s.OnPeer
	tr.OnStream = s.OnStream
	if err := s.Start(); err != nil {
		t.Fatalf("start %s: %v", listen, err)
	}
	t.Cleanup(s.Stop)
	return s
}

// TestReplicationStoreAGetB is the critical multi-node acceptance test:
// store on node A, fetch from node B after replication.
func TestReplicationStoreAGetB(t *testing.T) {
	logging.ConfigureLogger("error")

	km, err := crypto.NewKeyManagerFromClusterKey([]byte("shugo-test-cluster-key-32byt"))
	if err != nil {
		t.Fatal(err)
	}

	addrA := freePort(t)
	addrB := freePort(t)

	nodeA := startNode(t, addrA, nil, km)
	nodeB := startNode(t, addrB, []string{addrA}, km)

	// Wait until B has at least one peer (bootstrap connected).
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if nodeB.Has("__probe__") { // just force a sync point; peer count via store path
			break
		}
		// Dial is async; small poll for connection establishment.
		time.Sleep(50 * time.Millisecond)
		if time.Now().After(deadline.Add(-4 * time.Second)) {
			break
		}
	}
	time.Sleep(300 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	key := fmt.Sprintf("repl_%d.txt", time.Now().UnixNano())
	payload := []byte("replication-payload-across-peers")

	if err := nodeA.Store(ctx, key, bytes.NewReader(payload)); err != nil {
		t.Fatalf("store on A: %v", err)
	}

	// Allow stream replication to complete.
	deadline = time.Now().Add(10 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if nodeB.Has(key) {
			r, err := nodeB.Get(ctx, key)
			if err != nil {
				lastErr = err
				time.Sleep(100 * time.Millisecond)
				continue
			}
			got, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("read from B: %v", err)
			}
			if !bytes.Equal(got, payload) {
				t.Fatalf("payload mismatch: got %q want %q", got, payload)
			}
			return
		}
		// Also try Get (pending fetch) if replicate announcement raced
		r, err := nodeB.Get(ctx, key)
		if err == nil {
			got, readErr := io.ReadAll(r)
			if readErr == nil && bytes.Equal(got, payload) {
				return
			}
			lastErr = readErr
		} else {
			lastErr = err
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("file not available on B in time: last error: %v", lastErr)
}
