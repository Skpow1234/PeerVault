package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Client represents a gRPC client
type Client struct {
	conn   *grpc.ClientConn
	server string
}

// New creates a new gRPC client
func New(server string) *Client {
	return &Client{
		server: server,
	}
}

// Connect connects to the gRPC server
func (c *Client) Connect(ctx context.Context) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.NewClient(c.server, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	c.conn = conn
	return nil
}

// Disconnect disconnects from the gRPC server
func (c *Client) Disconnect() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.conn != nil
}

// GetConnection returns the gRPC connection
func (c *Client) GetConnection() *grpc.ClientConn {
	return c.conn
}

// SetServer sets the server address
func (c *Client) SetServer(server string) {
	c.server = server
}

// GetServer returns the server address
func (c *Client) GetServer() string {
	return c.server
}

// HealthCheck performs a health check
func (c *Client) HealthCheck(ctx context.Context) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to gRPC server")
	}

	// Simple health check by getting connection state
	state := c.conn.GetState()
	if state.String() == "READY" {
		return nil
	}

	return fmt.Errorf("gRPC connection not ready: %s", state.String())
}

// StreamOptions represents streaming options
type StreamOptions struct {
	BufferSize int
	Timeout    time.Duration
}

// DefaultStreamOptions returns default streaming options
func DefaultStreamOptions() *StreamOptions {
	return &StreamOptions{
		BufferSize: 1024,
		Timeout:    30 * time.Second,
	}
}

// ContextWithTimeout creates a context with timeout
func (c *Client) ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// ContextWithDeadline creates a context with deadline
func (c *Client) ContextWithDeadline(deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), deadline)
}
