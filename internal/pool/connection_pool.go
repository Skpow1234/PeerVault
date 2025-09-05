package pool

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Connection represents a pooled connection
type Connection interface {
	net.Conn
	IsHealthy() bool
	LastUsed() time.Time
	Reset() error
}

// ConnectionFactory creates new connections
type ConnectionFactory func(ctx context.Context, address string) (Connection, error)

// ConnectionPool manages a pool of connections
type ConnectionPool struct {
	factory     ConnectionFactory
	address     string
	connections chan Connection
	maxSize     int
	minSize     int
	currentSize int32
	mu          sync.RWMutex
	closed      int32
	cleanup     *time.Ticker
	ctx         context.Context
	cancel      context.CancelFunc
}

// PoolConfig holds configuration for the connection pool
type PoolConfig struct {
	MaxSize        int           `yaml:"max_size"`
	MinSize        int           `yaml:"min_size"`
	MaxIdleTime    time.Duration `yaml:"max_idle_time"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxSize:         10,
		MinSize:         2,
		MaxIdleTime:     5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		ConnectTimeout:  10 * time.Second,
	}
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(factory ConnectionFactory, address string, config *PoolConfig) *ConnectionPool {
	if config == nil {
		config = DefaultPoolConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &ConnectionPool{
		factory:     factory,
		address:     address,
		connections: make(chan Connection, config.MaxSize),
		maxSize:     config.MaxSize,
		minSize:     config.MinSize,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start cleanup routine
	pool.cleanup = time.NewTicker(config.CleanupInterval)
	go pool.cleanupRoutine(config.MaxIdleTime)

	// Pre-populate with minimum connections
	go pool.prePopulate()

	return pool
}

// Get retrieves a connection from the pool
func (cp *ConnectionPool) Get(ctx context.Context) (Connection, error) {
	if atomic.LoadInt32(&cp.closed) == 1 {
		return nil, fmt.Errorf("connection pool is closed")
	}

	select {
	case conn := <-cp.connections:
		// Check if connection is still healthy
		if conn.IsHealthy() {
			return conn, nil
		}
		// Connection is unhealthy, create a new one
		conn.Close()
		atomic.AddInt32(&cp.currentSize, -1)
		return cp.createConnection(ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// No available connections, create a new one if under limit
		return cp.createConnection(ctx)
	}
}

// Put returns a connection to the pool
func (cp *ConnectionPool) Put(conn Connection) {
	if atomic.LoadInt32(&cp.closed) == 1 {
		conn.Close()
		return
	}

	// Reset connection state
	if err := conn.Reset(); err != nil {
		conn.Close()
		atomic.AddInt32(&cp.currentSize, -1)
		return
	}

	select {
	case cp.connections <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection
		conn.Close()
		atomic.AddInt32(&cp.currentSize, -1)
	}
}

// Close closes the connection pool
func (cp *ConnectionPool) Close() error {
	if !atomic.CompareAndSwapInt32(&cp.closed, 0, 1) {
		return nil // Already closed
	}

	cp.cancel()
	cp.cleanup.Stop()

	// Close all connections in the pool
	close(cp.connections)
	for conn := range cp.connections {
		conn.Close()
	}

	return nil
}

// Stats returns pool statistics
func (cp *ConnectionPool) Stats() PoolStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	return PoolStats{
		CurrentSize: int(atomic.LoadInt32(&cp.currentSize)),
		MaxSize:     cp.maxSize,
		MinSize:     cp.minSize,
		Available:   len(cp.connections),
		InUse:       int(atomic.LoadInt32(&cp.currentSize)) - len(cp.connections),
	}
}

// PoolStats holds pool statistics
type PoolStats struct {
	CurrentSize int `json:"current_size"`
	MaxSize     int `json:"max_size"`
	MinSize     int `json:"min_size"`
	Available   int `json:"available"`
	InUse       int `json:"in_use"`
}

// createConnection creates a new connection
func (cp *ConnectionPool) createConnection(ctx context.Context) (Connection, error) {
	// Check if we're at the limit
	if int(atomic.LoadInt32(&cp.currentSize)) >= cp.maxSize {
		return nil, fmt.Errorf("connection pool is at maximum capacity")
	}

	conn, err := cp.factory(ctx, cp.address)
	if err != nil {
		return nil, err
	}

	atomic.AddInt32(&cp.currentSize, 1)
	return conn, nil
}

// prePopulate creates initial connections
func (cp *ConnectionPool) prePopulate() {
	for i := 0; i < cp.minSize; i++ {
		conn, err := cp.createConnection(cp.ctx)
		if err != nil {
			continue // Skip failed connections
		}
		
		select {
		case cp.connections <- conn:
		default:
			conn.Close()
			atomic.AddInt32(&cp.currentSize, -1)
		}
	}
}

// cleanupRoutine periodically cleans up idle connections
func (cp *ConnectionPool) cleanupRoutine(maxIdleTime time.Duration) {
	for {
		select {
		case <-cp.cleanup.C:
			cp.cleanupIdleConnections(maxIdleTime)
		case <-cp.ctx.Done():
			return
		}
	}
}

// cleanupIdleConnections removes idle connections
func (cp *ConnectionPool) cleanupIdleConnections(maxIdleTime time.Duration) {
	now := time.Now()
	connections := make([]Connection, 0, len(cp.connections))
	
	// Collect all connections
	for {
		select {
		case conn := <-cp.connections:
			connections = append(connections, conn)
		default:
			goto process
		}
	}

process:
	// Process connections
	for _, conn := range connections {
		if now.Sub(conn.LastUsed()) > maxIdleTime && int(atomic.LoadInt32(&cp.currentSize)) > cp.minSize {
			// Connection is idle and we're above minimum size
			conn.Close()
			atomic.AddInt32(&cp.currentSize, -1)
		} else {
			// Return connection to pool
			select {
			case cp.connections <- conn:
			default:
				conn.Close()
				atomic.AddInt32(&cp.currentSize, -1)
			}
		}
	}
}
