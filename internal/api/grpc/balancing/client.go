package balancing

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoadBalancedClient provides a gRPC client with load balancing
type LoadBalancedClient struct {
	loadBalancer *LoadBalancer
	logger       *slog.Logger
	clients      map[string]grpc.ClientConnInterface
	clientsMux   sync.RWMutex
	config       *LoadBalancedClientConfig
}

// LoadBalancedClientConfig represents the configuration for a load-balanced client
type LoadBalancedClientConfig struct {
	MaxRetries        int
	RetryDelay        time.Duration
	FailoverTimeout   time.Duration
	ConnectionTimeout time.Duration
	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration
	MaxReceiveSize    int
	MaxSendSize       int
	UserAgent         string
}

// DefaultLoadBalancedClientConfig returns the default client configuration
func DefaultLoadBalancedClientConfig() *LoadBalancedClientConfig {
	return &LoadBalancedClientConfig{
		MaxRetries:        3,
		RetryDelay:        time.Second,
		FailoverTimeout:   10 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		KeepAliveInterval: 30 * time.Second,
		KeepAliveTimeout:  5 * time.Second,
		MaxReceiveSize:    4 * 1024 * 1024, // 4MB
		MaxSendSize:       4 * 1024 * 1024, // 4MB
		UserAgent:         "peervault-grpc-client/1.0",
	}
}

// NewLoadBalancedClient creates a new load-balanced gRPC client
func NewLoadBalancedClient(loadBalancer *LoadBalancer, config *LoadBalancedClientConfig, logger *slog.Logger) *LoadBalancedClient {
	if config == nil {
		config = DefaultLoadBalancedClientConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &LoadBalancedClient{
		loadBalancer: loadBalancer,
		logger:       logger,
		clients:      make(map[string]grpc.ClientConnInterface),
		config:       config,
	}
}

// Invoke performs a unary RPC with load balancing and failover
func (c *LoadBalancedClient) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	sessionID := c.getSessionID(ctx)

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		// Get server from load balancer
		server, err := c.loadBalancer.GetServer(sessionID)
		if err != nil {
			c.logger.Error("Failed to get server from load balancer", "error", err, "attempt", attempt+1)
			if attempt == c.config.MaxRetries-1 {
				return status.Error(codes.Unavailable, "no healthy servers available")
			}
			time.Sleep(c.config.RetryDelay * time.Duration(attempt+1))
			continue
		}

		// Get or create client connection
		client, err := c.getClient(server.ID, server.Conn)
		if err != nil {
			c.logger.Error("Failed to get client connection", "server_id", server.ID, "error", err)
			c.loadBalancer.DecrementConnections(server.ID)
			continue
		}

		// Increment connection count
		c.loadBalancer.IncrementConnections(server.ID)

		// Perform the RPC call
		err = client.Invoke(ctx, method, args, reply, opts...)

		// Decrement connection count
		c.loadBalancer.DecrementConnections(server.ID)

		if err != nil {
			c.logger.Error("RPC call failed", "server_id", server.ID, "method", method, "error", err, "attempt", attempt+1)

			// Check if error is retryable
			if c.isRetryableError(err) && attempt < c.config.MaxRetries-1 {
				// Mark server as unhealthy if it's a connection error
				if c.isConnectionError(err) {
					c.logger.Warn("Marking server as unhealthy due to connection error", "server_id", server.ID)
					// In a real implementation, you would update the server health status
				}

				time.Sleep(c.config.RetryDelay * time.Duration(attempt+1))
				continue
			}

			return err
		}

		// Success
		c.logger.Debug("RPC call successful", "server_id", server.ID, "method", method)
		return nil
	}

	return status.Error(codes.DeadlineExceeded, "max retries exceeded")
}

// NewStream creates a new stream with load balancing and failover
func (c *LoadBalancedClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	sessionID := c.getSessionID(ctx)

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		// Get server from load balancer
		server, err := c.loadBalancer.GetServer(sessionID)
		if err != nil {
			c.logger.Error("Failed to get server for stream", "error", err, "attempt", attempt+1)
			if attempt == c.config.MaxRetries-1 {
				return nil, status.Error(codes.Unavailable, "no healthy servers available")
			}
			time.Sleep(c.config.RetryDelay * time.Duration(attempt+1))
			continue
		}

		// Get or create client connection
		client, err := c.getClient(server.ID, server.Conn)
		if err != nil {
			c.logger.Error("Failed to get client connection for stream", "server_id", server.ID, "error", err)
			continue
		}

		// Increment connection count
		c.loadBalancer.IncrementConnections(server.ID)

		// Create stream
		stream, err := client.NewStream(ctx, desc, method, opts...)
		if err != nil {
			c.logger.Error("Failed to create stream", "server_id", server.ID, "method", method, "error", err, "attempt", attempt+1)
			c.loadBalancer.DecrementConnections(server.ID)

			// Check if error is retryable
			if c.isRetryableError(err) && attempt < c.config.MaxRetries-1 {
				time.Sleep(c.config.RetryDelay * time.Duration(attempt+1))
				continue
			}

			return nil, err
		}

		// Wrap stream to handle connection cleanup
		wrappedStream := &loadBalancedStream{
			ClientStream: stream,
			client:       c,
			serverID:     server.ID,
		}

		c.logger.Debug("Stream created successfully", "server_id", server.ID, "method", method)
		return wrappedStream, nil
	}

	return nil, status.Error(codes.DeadlineExceeded, "max retries exceeded")
}

// getClient gets or creates a client connection for a server
func (c *LoadBalancedClient) getClient(serverID string, conn *grpc.ClientConn) (grpc.ClientConnInterface, error) {
	c.clientsMux.RLock()
	client, exists := c.clients[serverID]
	c.clientsMux.RUnlock()

	if exists {
		return client, nil
	}

	// Create new client connection
	c.clientsMux.Lock()
	defer c.clientsMux.Unlock()

	// Double-check after acquiring write lock
	if client, exists := c.clients[serverID]; exists {
		return client, nil
	}

	// Use the provided connection
	c.clients[serverID] = conn
	return conn, nil
}

// getSessionID extracts session ID from context
func (c *LoadBalancedClient) getSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

// isRetryableError checks if an error is retryable
func (c *LoadBalancedClient) isRetryableError(err error) bool {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
			return true
		case codes.Internal, codes.Unknown:
			return true
		default:
			return false
		}
	}
	return false
}

// isConnectionError checks if an error is a connection error
func (c *LoadBalancedClient) isConnectionError(err error) bool {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable, codes.DeadlineExceeded:
			return true
		default:
			return false
		}
	}
	return false
}

// Close closes the load-balanced client
func (c *LoadBalancedClient) Close() error {
	c.clientsMux.Lock()
	defer c.clientsMux.Unlock()

	// Close all client connections
	for serverID, client := range c.clients {
		if conn, ok := client.(*grpc.ClientConn); ok {
			if err := conn.Close(); err != nil {
				c.logger.Error("Failed to close client connection", "server_id", serverID, "error", err)
			}
		}
	}

	c.clients = make(map[string]grpc.ClientConnInterface)
	c.logger.Info("Load-balanced client closed")
	return nil
}

// GetStats returns client statistics
func (c *LoadBalancedClient) GetStats() map[string]interface{} {
	c.clientsMux.RLock()
	defer c.clientsMux.RUnlock()

	stats := make(map[string]interface{})
	stats["total_connections"] = len(c.clients)
	stats["max_retries"] = c.config.MaxRetries
	stats["retry_delay"] = c.config.RetryDelay
	stats["failover_timeout"] = c.config.FailoverTimeout

	// Get load balancer stats
	stats["load_balancer"] = c.loadBalancer.GetServerStats()

	return stats
}

// loadBalancedStream wraps a gRPC stream to handle connection cleanup
type loadBalancedStream struct {
	grpc.ClientStream
	client   *LoadBalancedClient
	serverID string
}

// CloseSend closes the send direction of the stream
func (lbs *loadBalancedStream) CloseSend() error {
	err := lbs.ClientStream.CloseSend()
	lbs.client.loadBalancer.DecrementConnections(lbs.serverID)
	return err
}

// Context returns the context of the stream
func (lbs *loadBalancedStream) Context() context.Context {
	return lbs.ClientStream.Context()
}

// Header returns the header metadata
func (lbs *loadBalancedStream) Header() (metadata.MD, error) {
	return lbs.ClientStream.Header()
}

// Trailer returns the trailer metadata
func (lbs *loadBalancedStream) Trailer() metadata.MD {
	return lbs.ClientStream.Trailer()
}

// SendMsg sends a message
func (lbs *loadBalancedStream) SendMsg(m interface{}) error {
	return lbs.ClientStream.SendMsg(m)
}

// RecvMsg receives a message
func (lbs *loadBalancedStream) RecvMsg(m interface{}) error {
	return lbs.ClientStream.RecvMsg(m)
}
