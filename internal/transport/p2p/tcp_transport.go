package p2p

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	// The underlying connection of the peer. Which in this case
	// is a TCP connection.
	net.Conn
	// if we dial and retrieve a conn => outbound == true
	// if we accept and retrieve a conn => outbound == false
	outbound bool

	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) CloseStream() { p.wg.Done() }

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Write(b)
	return err
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
	OnStream      func(Peer, io.Reader) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
	stopCh   chan struct{}
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	// Use LengthPrefixedDecoder by default if no decoder is specified
	if opts.Decoder == nil {
		opts.Decoder = LengthPrefixedDecoder{}
	}
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
		stopCh:           make(chan struct{}),
	}
}

// Addr implements the Transport interface return the address
// the transport is accepting connections.
func (t *TCPTransport) Addr() string { return t.ListenAddr }

// Consume implements the Tranport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC { return t.rpcch }

// Close implements the Transport interface.
func (t *TCPTransport) Close() error {
	select {
	case <-t.stopCh:
		// Already closed, just close the listener
	default:
		close(t.stopCh)
	}

	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}

// Dial implements the Transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	slog.Info("TCP transport listening on port", slog.String("port", t.ListenAddr))
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		select {
		case <-t.stopCh:
			return
		default:
			conn, err := t.listener.Accept()
			if err == io.EOF {
				return
			}
			if err != nil {
				// Only log if not shutting down
				select {
				case <-t.stopCh:
					return
				default:
					slog.Error("TCP accept error", slog.String("error", err.Error()))
				}
				continue
			}
			go t.handleConn(conn, false)
		}
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		slog.Error("dropping peer connection", slog.String("error", err.Error()))
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("failed to close connection", slog.String("error", closeErr.Error()))
		}
	}()

	peer := NewTCPPeer(conn, outbound)
	if err = t.HandshakeFunc(peer); err != nil {
		return
	}
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr().String()
		if rpc.Stream {
			peer.wg.Add(1)
			slog.Info("incoming stream", slog.String("peer", conn.RemoteAddr().String()))
			// Handle the stream if callback is provided
			if t.OnStream != nil {
				go func() {
					defer func() {
						peer.wg.Done()
					}()
					if err := t.OnStream(peer, conn); err != nil {
						slog.Error("failed to handle stream", slog.String("error", err.Error()))
					}
				}()
			} else {
				peer.wg.Done()
			}
			continue
		}
		t.rpcch <- rpc
	}
}

// handleIncomingStream handles incoming file streams
func (t *TCPTransport) handleIncomingStream(conn net.Conn, peer *TCPPeer) error {
	// Read the key length and key first
	var keyLen uint32
	if err := binary.Read(conn, binary.LittleEndian, &keyLen); err != nil {
		return fmt.Errorf("failed to read key length: %w", err)
	}

	keyBytes := make([]byte, keyLen)
	if _, err := io.ReadFull(conn, keyBytes); err != nil {
		return fmt.Errorf("failed to read key: %w", err)
	}
	key := string(keyBytes)

	slog.Info("receiving file stream", 
		slog.String("key", key),
		slog.String("peer", conn.RemoteAddr().String()))

	// For now, just read and discard the stream data
	// In a real implementation, this would store the file with the given key
	bytesRead, err := io.Copy(io.Discard, conn)
	if err != nil {
		return fmt.Errorf("failed to read stream data: %w", err)
	}

	slog.Info("received file stream", 
		slog.String("key", key),
		slog.Int64("bytesRead", bytesRead),
		slog.String("peer", conn.RemoteAddr().String()))
	
	return nil
}
