package streaming

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Skpow1234/Peervault/proto/peervault"
)

// StreamingService provides advanced streaming operations
type StreamingService struct {
	streamManager    *StreamManager
	multiplexer      *StreamMultiplexer
	logger           *slog.Logger
	activeStreams    map[string]*BidirectionalStream
	streamMux        sync.RWMutex
	eventBroadcaster *EventBroadcaster
}

// EventBroadcaster manages event broadcasting to multiple streams
type EventBroadcaster struct {
	subscribers map[string]chan *peervault.FileOperationEvent
	mutex       sync.RWMutex
	logger      *slog.Logger
}

// NewStreamingService creates a new streaming service
func NewStreamingService(logger *slog.Logger) *StreamingService {
	if logger == nil {
		logger = slog.Default()
	}

	config := DefaultStreamConfig()
	streamManager := NewStreamManager(config, logger)
	multiplexer := NewStreamMultiplexer(logger)
	eventBroadcaster := &EventBroadcaster{
		subscribers: make(map[string]chan *peervault.FileOperationEvent),
		logger:      logger,
	}

	return &StreamingService{
		streamManager:    streamManager,
		multiplexer:      multiplexer,
		logger:           logger,
		activeStreams:    make(map[string]*BidirectionalStream),
		eventBroadcaster: eventBroadcaster,
	}
}

// CreateBidirectionalStream creates a new bidirectional stream
func (s *StreamingService) CreateBidirectionalStream(ctx context.Context, streamID string) (*BidirectionalStream, error) {
	s.streamMux.Lock()
	defer s.streamMux.Unlock()

	// Check if stream already exists
	if _, exists := s.activeStreams[streamID]; exists {
		return nil, status.Error(codes.AlreadyExists, "stream already exists")
	}

	// Create new bidirectional stream
	stream := s.streamManager.NewBidirectionalStream(streamID)
	s.activeStreams[streamID] = stream

	// Add to multiplexer
	s.multiplexer.AddStream(streamID, stream.Stream)

	s.logger.Info("Created bidirectional stream", "stream_id", streamID)
	return stream, nil
}

// CloseBidirectionalStream closes a bidirectional stream
func (s *StreamingService) CloseBidirectionalStream(streamID string) error {
	s.streamMux.Lock()
	defer s.streamMux.Unlock()

	stream, exists := s.activeStreams[streamID]
	if !exists {
		return status.Error(codes.NotFound, "stream not found")
	}

	// Close the stream
	stream.cancel()
	delete(s.activeStreams, streamID)
	s.multiplexer.RemoveStream(streamID)

	s.logger.Info("Closed bidirectional stream", "stream_id", streamID)
	return nil
}

// StreamFileOperationsWithRecovery streams file operations with error recovery
func (s *StreamingService) StreamFileOperationsWithRecovery(ctx context.Context, streamID string) (<-chan *peervault.FileOperationEvent, error) {
	stream, err := s.CreateBidirectionalStream(ctx, streamID)
	if err != nil {
		return nil, err
	}

	eventChan := make(chan *peervault.FileOperationEvent, 100)

	// Start event generation goroutine
	go s.generateFileOperationEvents(stream, eventChan)

	return eventChan, nil
}

// StreamPeerEventsWithRecovery streams peer events with error recovery
func (s *StreamingService) StreamPeerEventsWithRecovery(ctx context.Context, streamID string) (<-chan *peervault.PeerEvent, error) {
	stream, err := s.CreateBidirectionalStream(ctx, streamID)
	if err != nil {
		return nil, err
	}

	eventChan := make(chan *peervault.PeerEvent, 100)

	// Start event generation goroutine
	go s.generatePeerEvents(stream, eventChan)

	return eventChan, nil
}

// StreamSystemEventsWithRecovery streams system events with error recovery
func (s *StreamingService) StreamSystemEventsWithRecovery(ctx context.Context, streamID string) (<-chan *peervault.SystemEvent, error) {
	stream, err := s.CreateBidirectionalStream(ctx, streamID)
	if err != nil {
		return nil, err
	}

	eventChan := make(chan *peervault.SystemEvent, 100)

	// Start event generation goroutine
	go s.generateSystemEvents(stream, eventChan)

	return eventChan, nil
}

// generateFileOperationEvents generates file operation events with error recovery
func (s *StreamingService) generateFileOperationEvents(stream *BidirectionalStream, eventChan chan<- *peervault.FileOperationEvent) {
	defer close(eventChan)
	defer func() {
		if err := s.CloseBidirectionalStream(stream.ID); err != nil {
			s.logger.Warn("Failed to close bidirectional stream", "stream_id", stream.ID, "error", err)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	retryCount := 0
	maxRetries := stream.Config.MaxRetries

	for {
		select {
		case <-ticker.C:
			event := &peervault.FileOperationEvent{
				EventType: "periodic_check",
				FileKey:   fmt.Sprintf("file_%d", time.Now().Unix()),
				PeerId:    "local",
				Timestamp: timestamppb.Now(),
				Metadata: map[string]string{
					"type":        "periodic",
					"stream_id":   stream.ID,
					"retry_count": fmt.Sprintf("%d", retryCount),
				},
			}

			// Try to send event with flow control
			if err := stream.Send(event); err != nil {
				s.logger.Error("Failed to send file operation event", "stream_id", stream.ID, "error", err)

				// Implement retry logic
				if retryCount < maxRetries {
					retryCount++
					s.logger.Info("Retrying file operation event", "stream_id", stream.ID, "retry", retryCount)

					// Wait before retry
					time.Sleep(stream.Config.RetryDelay * time.Duration(retryCount))
					continue
				} else {
					s.logger.Error("Max retries exceeded for file operation event", "stream_id", stream.ID)
					return
				}
			}

			// Reset retry count on successful send
			retryCount = 0

			// Send to event channel
			select {
			case eventChan <- event:
			default:
				s.logger.Warn("Event channel full, dropping file operation event", "stream_id", stream.ID)
			}

		case <-stream.ctx.Done():
			s.logger.Info("File operation event generation stopped", "stream_id", stream.ID)
			return
		}
	}
}

// generatePeerEvents generates peer events with error recovery
func (s *StreamingService) generatePeerEvents(stream *BidirectionalStream, eventChan chan<- *peervault.PeerEvent) {
	defer close(eventChan)
	defer func() {
		if err := s.CloseBidirectionalStream(stream.ID); err != nil {
			s.logger.Warn("Failed to close bidirectional stream", "stream_id", stream.ID, "error", err)
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	retryCount := 0
	maxRetries := stream.Config.MaxRetries

	for {
		select {
		case <-ticker.C:
			event := &peervault.PeerEvent{
				EventType: "health_check",
				PeerId:    fmt.Sprintf("peer_%d", time.Now().Unix()),
				Timestamp: timestamppb.Now(),
				Metadata: map[string]string{
					"type":        "periodic",
					"stream_id":   stream.ID,
					"retry_count": fmt.Sprintf("%d", retryCount),
				},
			}

			// Try to send event with flow control
			if err := stream.Send(event); err != nil {
				s.logger.Error("Failed to send peer event", "stream_id", stream.ID, "error", err)

				// Implement retry logic
				if retryCount < maxRetries {
					retryCount++
					s.logger.Info("Retrying peer event", "stream_id", stream.ID, "retry", retryCount)

					// Wait before retry
					time.Sleep(stream.Config.RetryDelay * time.Duration(retryCount))
					continue
				} else {
					s.logger.Error("Max retries exceeded for peer event", "stream_id", stream.ID)
					return
				}
			}

			// Reset retry count on successful send
			retryCount = 0

			// Send to event channel
			select {
			case eventChan <- event:
			default:
				s.logger.Warn("Event channel full, dropping peer event", "stream_id", stream.ID)
			}

		case <-stream.ctx.Done():
			s.logger.Info("Peer event generation stopped", "stream_id", stream.ID)
			return
		}
	}
}

// generateSystemEvents generates system events with error recovery
func (s *StreamingService) generateSystemEvents(stream *BidirectionalStream, eventChan chan<- *peervault.SystemEvent) {
	defer close(eventChan)
	defer func() {
		if err := s.CloseBidirectionalStream(stream.ID); err != nil {
			s.logger.Warn("Failed to close bidirectional stream", "stream_id", stream.ID, "error", err)
		}
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	retryCount := 0
	maxRetries := stream.Config.MaxRetries

	for {
		select {
		case <-ticker.C:
			event := &peervault.SystemEvent{
				EventType: "status_update",
				Component: "streaming_service",
				Timestamp: timestamppb.Now(),
				Message:   "System running normally",
				Metadata: map[string]string{
					"uptime":      time.Since(time.Now()).String(),
					"stream_id":   stream.ID,
					"retry_count": fmt.Sprintf("%d", retryCount),
				},
			}

			// Try to send event with flow control
			if err := stream.Send(event); err != nil {
				s.logger.Error("Failed to send system event", "stream_id", stream.ID, "error", err)

				// Implement retry logic
				if retryCount < maxRetries {
					retryCount++
					s.logger.Info("Retrying system event", "stream_id", stream.ID, "retry", retryCount)

					// Wait before retry
					time.Sleep(stream.Config.RetryDelay * time.Duration(retryCount))
					continue
				} else {
					s.logger.Error("Max retries exceeded for system event", "stream_id", stream.ID)
					return
				}
			}

			// Reset retry count on successful send
			retryCount = 0

			// Send to event channel
			select {
			case eventChan <- event:
			default:
				s.logger.Warn("Event channel full, dropping system event", "stream_id", stream.ID)
			}

		case <-stream.ctx.Done():
			s.logger.Info("System event generation stopped", "stream_id", stream.ID)
			return
		}
	}
}

// BroadcastFileOperationEvent broadcasts a file operation event to all subscribers
func (s *StreamingService) BroadcastFileOperationEvent(event *peervault.FileOperationEvent) {
	s.eventBroadcaster.mutex.RLock()
	defer s.eventBroadcaster.mutex.RUnlock()

	for subscriberID, eventChan := range s.eventBroadcaster.subscribers {
		select {
		case eventChan <- event:
			s.logger.Debug("File operation event broadcasted", "subscriber_id", subscriberID)
		default:
			s.logger.Warn("Subscriber channel full, dropping event", "subscriber_id", subscriberID)
		}
	}
}

// SubscribeToFileOperationEvents subscribes to file operation events
func (s *StreamingService) SubscribeToFileOperationEvents(subscriberID string) (<-chan *peervault.FileOperationEvent, error) {
	s.eventBroadcaster.mutex.Lock()
	defer s.eventBroadcaster.mutex.Unlock()

	if _, exists := s.eventBroadcaster.subscribers[subscriberID]; exists {
		return nil, status.Error(codes.AlreadyExists, "subscriber already exists")
	}

	eventChan := make(chan *peervault.FileOperationEvent, 100)
	s.eventBroadcaster.subscribers[subscriberID] = eventChan

	s.logger.Info("Subscribed to file operation events", "subscriber_id", subscriberID)
	return eventChan, nil
}

// UnsubscribeFromFileOperationEvents unsubscribes from file operation events
func (s *StreamingService) UnsubscribeFromFileOperationEvents(subscriberID string) error {
	s.eventBroadcaster.mutex.Lock()
	defer s.eventBroadcaster.mutex.Unlock()

	eventChan, exists := s.eventBroadcaster.subscribers[subscriberID]
	if !exists {
		return status.Error(codes.NotFound, "subscriber not found")
	}

	close(eventChan)
	delete(s.eventBroadcaster.subscribers, subscriberID)

	s.logger.Info("Unsubscribed from file operation events", "subscriber_id", subscriberID)
	return nil
}

// GetStreamStats returns statistics for all active streams
func (s *StreamingService) GetStreamStats() map[string]interface{} {
	s.streamMux.RLock()
	defer s.streamMux.RUnlock()

	stats := make(map[string]interface{})
	stats["active_streams"] = len(s.activeStreams)
	stats["multiplexer_stats"] = s.multiplexer.GetStats()
	stats["subscribers"] = len(s.eventBroadcaster.subscribers)

	return stats
}

// HealthCheck performs a health check on the streaming service
func (s *StreamingService) HealthCheck() map[string]interface{} {
	s.streamMux.RLock()
	defer s.streamMux.RUnlock()

	healthyStreams := 0
	totalStreams := len(s.activeStreams)

	for _, stream := range s.activeStreams {
		if stream.IsHealthy() {
			healthyStreams++
		}
	}

	return map[string]interface{}{
		"status":          "healthy",
		"total_streams":   totalStreams,
		"healthy_streams": healthyStreams,
		"subscribers":     len(s.eventBroadcaster.subscribers),
		"timestamp":       time.Now(),
	}
}

// Close closes the streaming service and all its resources
func (s *StreamingService) Close() error {
	s.logger.Info("Closing streaming service")

	// Close all active streams
	s.streamMux.Lock()
	for streamID, stream := range s.activeStreams {
		stream.cancel()
		s.logger.Info("Closed stream", "stream_id", streamID)
	}
	s.activeStreams = make(map[string]*BidirectionalStream)
	s.streamMux.Unlock()

	// Close stream manager
	s.streamManager.CloseAllStreams()

	// Close all subscribers
	s.eventBroadcaster.mutex.Lock()
	for subscriberID, eventChan := range s.eventBroadcaster.subscribers {
		close(eventChan)
		s.logger.Info("Closed subscriber", "subscriber_id", subscriberID)
	}
	s.eventBroadcaster.subscribers = make(map[string]chan *peervault.FileOperationEvent)
	s.eventBroadcaster.mutex.Unlock()

	s.logger.Info("Streaming service closed")
	return nil
}
