package streaming

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamConfig represents configuration for streaming operations
type StreamConfig struct {
	BufferSize        int           // Buffer size for the stream
	MaxRetries        int           // Maximum number of retry attempts
	RetryDelay        time.Duration // Delay between retries
	BackpressureLimit int           // Maximum pending messages before backpressure
	HeartbeatInterval time.Duration // Interval for heartbeat messages
	Timeout           time.Duration // Stream timeout
}

// DefaultStreamConfig returns the default streaming configuration
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		BufferSize:        1000,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		BackpressureLimit: 100,
		HeartbeatInterval: 30 * time.Second,
		Timeout:           5 * time.Minute,
	}
}

// StreamManager manages streaming operations with advanced patterns
type StreamManager struct {
	config    *StreamConfig
	logger    *slog.Logger
	streams   map[string]*Stream
	streamMux sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// Stream represents a managed stream with error recovery and backpressure
type Stream struct {
	ID           string
	Config       *StreamConfig
	Logger       *slog.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	messageChan  chan interface{}
	errorChan    chan error
	retryCount   int
	lastActivity time.Time
	backpressure bool
	mutex        sync.RWMutex
}

// StreamMessage represents a message in the stream
type StreamMessage struct {
	ID        string
	Data      interface{}
	Timestamp time.Time
	Retry     int
}

// StreamError represents an error in the stream
type StreamError struct {
	Code    codes.Code
	Message string
	Retry   bool
	Stream  string
}

// NewStreamManager creates a new stream manager
func NewStreamManager(config *StreamConfig, logger *slog.Logger) *StreamManager {
	if config == nil {
		config = DefaultStreamConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &StreamManager{
		config:  config,
		logger:  logger,
		streams: make(map[string]*Stream),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// CreateStream creates a new managed stream
func (sm *StreamManager) CreateStream(id string) *Stream {
	sm.streamMux.Lock()
	defer sm.streamMux.Unlock()

	ctx, cancel := context.WithTimeout(sm.ctx, sm.config.Timeout)

	stream := &Stream{
		ID:           id,
		Config:       sm.config,
		Logger:       sm.logger.With("stream_id", id),
		ctx:          ctx,
		cancel:       cancel,
		messageChan:  make(chan interface{}, sm.config.BufferSize),
		errorChan:    make(chan error, sm.config.BufferSize),
		lastActivity: time.Now(),
	}

	sm.streams[id] = stream

	// Start stream management goroutines
	go stream.manageStream()
	go stream.handleBackpressure()

	sm.logger.Info("Created stream", "stream_id", id)
	return stream
}

// GetStream retrieves a stream by ID
func (sm *StreamManager) GetStream(id string) (*Stream, bool) {
	sm.streamMux.RLock()
	defer sm.streamMux.RUnlock()

	stream, exists := sm.streams[id]
	return stream, exists
}

// CloseStream closes a stream
func (sm *StreamManager) CloseStream(id string) error {
	sm.streamMux.Lock()
	defer sm.streamMux.Unlock()

	stream, exists := sm.streams[id]
	if !exists {
		return fmt.Errorf("stream not found: %s", id)
	}

	stream.cancel()
	delete(sm.streams, id)

	sm.logger.Info("Closed stream", "stream_id", id)
	return nil
}

// CloseAllStreams closes all managed streams
func (sm *StreamManager) CloseAllStreams() {
	sm.streamMux.Lock()
	defer sm.streamMux.Unlock()

	for id, stream := range sm.streams {
		stream.cancel()
		sm.logger.Info("Closed stream", "stream_id", id)
	}

	sm.streams = make(map[string]*Stream)
	sm.cancel()
}

// Stream Methods

// SendMessage sends a message to the stream with error recovery
func (s *Stream) SendMessage(message interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if stream is in backpressure mode
	if s.backpressure {
		return status.Error(codes.ResourceExhausted, "stream is in backpressure mode")
	}

	// Check if stream is closed
	select {
	case <-s.ctx.Done():
		return status.Error(codes.Canceled, "stream is closed")
	default:
	}

	streamMsg := &StreamMessage{
		ID:        fmt.Sprintf("%s-%d", s.ID, time.Now().UnixNano()),
		Data:      message,
		Timestamp: time.Now(),
		Retry:     0,
	}

	select {
	case s.messageChan <- streamMsg:
		s.lastActivity = time.Now()
		s.Logger.Debug("Message sent to stream", "message_id", streamMsg.ID)
		return nil
	default:
		// Channel is full, trigger backpressure
		s.backpressure = true
		s.Logger.Warn("Stream channel full, triggering backpressure", "stream_id", s.ID)
		return status.Error(codes.ResourceExhausted, "stream channel full")
	}
}

// ReceiveMessage receives a message from the stream
func (s *Stream) ReceiveMessage() (interface{}, error) {
	select {
	case message := <-s.messageChan:
		s.mutex.Lock()
		s.lastActivity = time.Now()
		s.mutex.Unlock()
		return message, nil
	case err := <-s.errorChan:
		return nil, err
	case <-s.ctx.Done():
		return nil, status.Error(codes.Canceled, "stream is closed")
	}
}

// SendError sends an error to the stream
func (s *Stream) SendError(err error) {
	select {
	case s.errorChan <- err:
		s.Logger.Error("Error sent to stream", "error", err)
	default:
		s.Logger.Error("Error channel full, dropping error", "error", err)
	}
}

// IsHealthy checks if the stream is healthy
func (s *Stream) IsHealthy() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if stream is closed
	select {
	case <-s.ctx.Done():
		return false
	default:
	}

	// Check if stream has been inactive for too long
	if time.Since(s.lastActivity) > s.Config.Timeout {
		return false
	}

	return true
}

// GetStats returns stream statistics
func (s *Stream) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"id":            s.ID,
		"buffer_size":   len(s.messageChan),
		"max_buffer":    cap(s.messageChan),
		"retry_count":   s.retryCount,
		"last_activity": s.lastActivity,
		"backpressure":  s.backpressure,
		"healthy":       s.IsHealthy(),
		"uptime":        time.Since(s.lastActivity),
	}
}

// manageStream manages the stream lifecycle
func (s *Stream) manageStream() {
	ticker := time.NewTicker(s.Config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send heartbeat
			heartbeat := &StreamMessage{
				ID:        fmt.Sprintf("%s-heartbeat-%d", s.ID, time.Now().UnixNano()),
				Data:      "heartbeat",
				Timestamp: time.Now(),
				Retry:     0,
			}

			select {
			case s.messageChan <- heartbeat:
				s.Logger.Debug("Heartbeat sent", "stream_id", s.ID)
			default:
				s.Logger.Warn("Failed to send heartbeat, channel full", "stream_id", s.ID)
			}

		case <-s.ctx.Done():
			s.Logger.Info("Stream management stopped", "stream_id", s.ID)
			return
		}
	}
}

// handleBackpressure manages backpressure for the stream
func (s *Stream) handleBackpressure() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mutex.Lock()
			if s.backpressure && len(s.messageChan) < s.Config.BackpressureLimit/2 {
				s.backpressure = false
				s.Logger.Info("Backpressure cleared", "stream_id", s.ID)
			}
			s.mutex.Unlock()

		case <-s.ctx.Done():
			return
		}
	}
}

// BidirectionalStream represents a bidirectional stream with flow control
type BidirectionalStream struct {
	*Stream
	SendChan    chan interface{}
	ReceiveChan chan interface{}
	FlowControl *FlowController
}

// FlowController manages flow control for bidirectional streams
type FlowController struct {
	WindowSize    int
	CurrentWindow int
	Mutex         sync.RWMutex
}

// NewBidirectionalStream creates a new bidirectional stream
func (sm *StreamManager) NewBidirectionalStream(id string) *BidirectionalStream {
	stream := sm.CreateStream(id)

	return &BidirectionalStream{
		Stream:      stream,
		SendChan:    make(chan interface{}, sm.config.BufferSize),
		ReceiveChan: make(chan interface{}, sm.config.BufferSize),
		FlowControl: &FlowController{
			WindowSize:    sm.config.BufferSize,
			CurrentWindow: sm.config.BufferSize,
		},
	}
}

// Send sends a message with flow control
func (bs *BidirectionalStream) Send(message interface{}) error {
	bs.FlowControl.Mutex.Lock()
	defer bs.FlowControl.Mutex.Unlock()

	if bs.FlowControl.CurrentWindow <= 0 {
		return status.Error(codes.ResourceExhausted, "flow control window exhausted")
	}

	select {
	case bs.SendChan <- message:
		bs.FlowControl.CurrentWindow--
		bs.Logger.Debug("Message sent with flow control", "window", bs.FlowControl.CurrentWindow)
		return nil
	default:
		return status.Error(codes.ResourceExhausted, "send channel full")
	}
}

// Receive receives a message with flow control
func (bs *BidirectionalStream) Receive() (interface{}, error) {
	select {
	case message := <-bs.ReceiveChan:
		// Acknowledge receipt to increase window
		bs.FlowControl.Mutex.Lock()
		bs.FlowControl.CurrentWindow++
		bs.FlowControl.Mutex.Unlock()
		return message, nil
	case <-bs.ctx.Done():
		return nil, status.Error(codes.Canceled, "stream is closed")
	}
}

// Acknowledge acknowledges receipt of a message (increases flow control window)
func (bs *BidirectionalStream) Acknowledge() {
	bs.FlowControl.Mutex.Lock()
	defer bs.FlowControl.Mutex.Unlock()

	if bs.FlowControl.CurrentWindow < bs.FlowControl.WindowSize {
		bs.FlowControl.CurrentWindow++
		bs.Logger.Debug("Message acknowledged", "window", bs.FlowControl.CurrentWindow)
	}
}

// StreamMultiplexer allows multiple streams to be multiplexed over a single connection
type StreamMultiplexer struct {
	Streams map[string]*Stream
	Mutex   sync.RWMutex
	Logger  *slog.Logger
}

// NewStreamMultiplexer creates a new stream multiplexer
func NewStreamMultiplexer(logger *slog.Logger) *StreamMultiplexer {
	if logger == nil {
		logger = slog.Default()
	}

	return &StreamMultiplexer{
		Streams: make(map[string]*Stream),
		Logger:  logger,
	}
}

// AddStream adds a stream to the multiplexer
func (sm *StreamMultiplexer) AddStream(id string, stream *Stream) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	sm.Streams[id] = stream
	sm.Logger.Info("Stream added to multiplexer", "stream_id", id)
}

// RemoveStream removes a stream from the multiplexer
func (sm *StreamMultiplexer) RemoveStream(id string) {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	if stream, exists := sm.Streams[id]; exists {
		stream.cancel()
		delete(sm.Streams, id)
		sm.Logger.Info("Stream removed from multiplexer", "stream_id", id)
	}
}

// GetStream retrieves a stream from the multiplexer
func (sm *StreamMultiplexer) GetStream(id string) (*Stream, bool) {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()

	stream, exists := sm.Streams[id]
	return stream, exists
}

// BroadcastMessage broadcasts a message to all streams
func (sm *StreamMultiplexer) BroadcastMessage(message interface{}) {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()

	for id, stream := range sm.Streams {
		if err := stream.SendMessage(message); err != nil {
			sm.Logger.Error("Failed to broadcast message to stream", "stream_id", id, "error", err)
		}
	}
}

// GetStats returns statistics for all streams
func (sm *StreamMultiplexer) GetStats() map[string]interface{} {
	sm.Mutex.RLock()
	defer sm.Mutex.RUnlock()

	stats := make(map[string]interface{})
	for id, stream := range sm.Streams {
		stats[id] = stream.GetStats()
	}

	return stats
}
