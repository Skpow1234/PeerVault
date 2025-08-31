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
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	// Use LengthPrefixedDecoder by default if no decoder is specified
	if opts.Decoder == nil {
		opts.Decoder = LengthPrefixedDecoder{}
	}
	return &TCPTransport{TCPTransportOpts: opts, rpcch: make(chan RPC, 1024)}
}

// Addr implements the Transport interface return the address
// the transport is accepting connections.
func (t *TCPTransport) Addr() string { return t.ListenAddr }

// Consume implements the Tranport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC { return t.rpcch }

// Close implements the Transport interface.
func (t *TCPTransport) Close() error { return t.listener.Close() }

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
		conn, err := t.listener.Accept()
		if err == io.EOF {
			return
		}
		if err != nil {
			slog.Error("TCP accept error", slog.String("error", err.Error()))
		}
		go t.handleConn(conn, false)
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
			peer.wg.Wait()
			slog.Info("stream closed", slog.String("peer", conn.RemoteAddr().String()))
			continue
		}
		t.rpcch <- rpc
	}
}
