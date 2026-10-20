package p2p

import (
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
	_, err := t.DialPeer(addr)
	return err
}

// DialPeer dials, handshakes, registers OnPeer, starts the read loop, and returns the peer.
func (t *TCPTransport) DialPeer(addr string) (Peer, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	peer, err := t.setupConn(conn, true)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return peer, nil
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
	peer, err := t.setupConn(conn, outbound)
	if err != nil {
		slog.Error("dropping peer connection", slog.String("error", err.Error()))
		_ = conn.Close()
		return
	}
	_ = peer // read loop already started in setupConn
}

// setupConn performs handshake, OnPeer, and starts the read loop.
func (t *TCPTransport) setupConn(conn net.Conn, outbound bool) (Peer, error) {
	peer := NewTCPPeer(conn, outbound)

	if t.HandshakeFunc != nil {
		if err := t.HandshakeFunc(peer); err != nil {
			return nil, err
		}
	}
	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			return nil, err
		}
	}

	go t.readLoop(peer)
	return peer, nil
}

func (t *TCPTransport) readLoop(peer *TCPPeer) {
	defer func() {
		if closeErr := peer.Close(); closeErr != nil {
			slog.Error("failed to close connection", slog.String("error", closeErr.Error()))
		}
	}()

	for {
		select {
		case <-t.stopCh:
			return
		default:
		}

		rpc := RPC{}
		err := t.Decoder.Decode(peer, &rpc)
		if err != nil {
			slog.Debug("peer read ended", slog.String("peer", peer.RemoteAddr().String()), slog.String("error", err.Error()))
			return
		}
		rpc.From = peer.RemoteAddr().String()

		if rpc.Stream {
			// Handle streams synchronously so the next frame is not decoded mid-stream.
			if t.OnStream != nil {
				if err := t.OnStream(peer, peer); err != nil {
					slog.Error("failed to handle stream", slog.String("error", err.Error()))
					return
				}
			}
			continue
		}

		select {
		case t.rpcch <- rpc:
		case <-t.stopCh:
			return
		}
	}
}
