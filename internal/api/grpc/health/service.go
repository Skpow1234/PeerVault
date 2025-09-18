package health

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Skpow1234/Peervault/proto/peervault"
)

// HealthService implements the gRPC health service
type HealthService struct {
	peervault.UnimplementedPeerVaultServiceServer
	healthChecker *AdvancedHealthChecker
	logger        *slog.Logger
}

// NewHealthService creates a new health service
func NewHealthService(healthChecker *AdvancedHealthChecker, logger *slog.Logger) *HealthService {
	if logger == nil {
		logger = slog.Default()
	}

	return &HealthService{
		healthChecker: healthChecker,
		logger:        logger,
	}
}

// HealthCheck implements the basic health check endpoint
func (hs *HealthService) HealthCheck(ctx context.Context, req *emptypb.Empty) (*peervault.HealthResponse, error) {
	hs.logger.Debug("Health check requested")

	response := hs.healthChecker.GetHealthStatus()

	hs.logger.Debug("Health check completed", "status", response.Status)
	return response, nil
}

// GetDetailedHealth returns detailed health information
func (hs *HealthService) GetDetailedHealth(ctx context.Context, req *emptypb.Empty) (*peervault.HealthResponse, error) {
	hs.logger.Debug("Detailed health check requested")

	detailedStatus := hs.healthChecker.GetDetailedHealthStatus()

	// Convert detailed status to HealthResponse
	response := &peervault.HealthResponse{
		Status:    detailedStatus["overall_status"].(string),
		Timestamp: timestamppb.Now(),
		Version:   "1.0.0",
	}

	hs.logger.Debug("Detailed health check completed", "status", response.Status)
	return response, nil
}

// GetComponentHealth returns health information for a specific component
func (hs *HealthService) GetComponentHealth(ctx context.Context, req *peervault.ComponentHealthRequest) (*peervault.ComponentHealthResponse, error) {
	hs.logger.Debug("Component health check requested", "component", req.Component)

	result, err := hs.healthChecker.GetComponentHealth(req.Component)
	if err != nil {
		hs.logger.Error("Component health check failed", "component", req.Component, "error", err)
		return nil, status.Error(codes.NotFound, "component not found")
	}

	response := &peervault.ComponentHealthResponse{
		Component: result.Component,
		Status:    string(result.Status),
		Message:   result.Message,
		Timestamp: result.Timestamp,
		Duration:  int64(result.Duration.Nanoseconds()),
		Metrics:   make(map[string]string),
		Details:   make(map[string]string),
	}

	// Convert metrics to string map
	for key, value := range result.Metrics {
		response.Metrics[key] = fmt.Sprintf("%v", value)
	}

	// Convert details to string map
	for key, value := range result.Details {
		response.Details[key] = fmt.Sprintf("%v", value)
	}

	hs.logger.Debug("Component health check completed", "component", req.Component, "status", result.Status)
	return response, nil
}

// ForceHealthCheck forces a health check for a specific component
func (hs *HealthService) ForceHealthCheck(ctx context.Context, req *peervault.ForceHealthCheckRequest) (*peervault.ForceHealthCheckResponse, error) {
	hs.logger.Info("Force health check requested", "component", req.Component)

	var err error
	if req.Component == "" {
		// Force health check for all components
		hs.healthChecker.ForceHealthCheckAll()
	} else {
		// Force health check for specific component
		err = hs.healthChecker.ForceHealthCheck(req.Component)
	}

	if err != nil {
		hs.logger.Error("Force health check failed", "component", req.Component, "error", err)
		return nil, status.Error(codes.Internal, "failed to force health check")
	}

	response := &peervault.ForceHealthCheckResponse{
		Success: true,
		Message: "Health check completed successfully",
	}

	hs.logger.Info("Force health check completed", "component", req.Component)
	return response, nil
}

// GetHealthMetrics returns health metrics
func (hs *HealthService) GetHealthMetrics(ctx context.Context, req *emptypb.Empty) (*peervault.HealthMetricsResponse, error) {
	hs.logger.Debug("Health metrics requested")

	metrics := hs.healthChecker.GetHealthMetrics()

	response := &peervault.HealthMetricsResponse{
		Metrics: make(map[string]string),
	}

	// Convert metrics to string map
	for key, value := range metrics {
		response.Metrics[key] = fmt.Sprintf("%v", value)
	}

	hs.logger.Debug("Health metrics completed")
	return response, nil
}

// GetHealthTraces returns health traces
func (hs *HealthService) GetHealthTraces(ctx context.Context, req *emptypb.Empty) (*peervault.HealthTracesResponse, error) {
	hs.logger.Debug("Health traces requested")

	traces := hs.healthChecker.GetHealthTraces()

	response := &peervault.HealthTracesResponse{
		Traces: make([]*peervault.HealthTrace, 0, len(traces)),
	}

	// Convert traces to protobuf format
	for _, trace := range traces {
		pbTrace := &peervault.HealthTrace{
			Id:        trace.ID,
			Component: trace.Component,
			StartTime: trace.StartTime,
			EndTime:   trace.EndTime,
			Duration:  int64(trace.Duration.Nanoseconds()),
			Status:    string(trace.Status),
			Details:   make(map[string]string),
		}

		// Convert details to string map
		for key, value := range trace.Details {
			pbTrace.Details[key] = fmt.Sprintf("%v", value)
		}

		response.Traces = append(response.Traces, pbTrace)
	}

	hs.logger.Debug("Health traces completed", "trace_count", len(traces))
	return response, nil
}

// GetHealthProfiles returns health profiles
func (hs *HealthService) GetHealthProfiles(ctx context.Context, req *emptypb.Empty) (*peervault.HealthProfilesResponse, error) {
	hs.logger.Debug("Health profiles requested")

	profiles := hs.healthChecker.GetHealthProfiles()

	response := &peervault.HealthProfilesResponse{
		Profiles: make([]*peervault.HealthProfile, 0, len(profiles)),
	}

	// Convert profiles to protobuf format
	for _, profile := range profiles {
		pbProfile := &peervault.HealthProfile{
			Component:     profile.Component,
			CheckCount:    profile.CheckCount,
			TotalDuration: int64(profile.TotalDuration.Nanoseconds()),
			AvgDuration:   int64(profile.AvgDuration.Nanoseconds()),
			MinDuration:   int64(profile.MinDuration.Nanoseconds()),
			MaxDuration:   int64(profile.MaxDuration.Nanoseconds()),
			LastUpdated:   profile.LastUpdated,
		}

		response.Profiles = append(response.Profiles, pbProfile)
	}

	hs.logger.Debug("Health profiles completed", "profile_count", len(profiles))
	return response, nil
}

// StreamHealthEvents streams health events
func (hs *HealthService) StreamHealthEvents(req *emptypb.Empty, stream peervault.PeerVaultService_StreamHealthEventsServer) error {
	hs.logger.Info("Health events stream started")

	// Create a channel for health events
	eventChan := make(chan *peervault.HealthEvent, 100)
	defer close(eventChan)

	// Start a goroutine to generate health events
	go hs.generateHealthEvents(eventChan, stream.Context())

	// Send events to the client
	for {
		select {
		case event := <-eventChan:
			if err := stream.Send(event); err != nil {
				hs.logger.Error("Error sending health event", "error", err)
				return status.Error(codes.Internal, "failed to send health event")
			}
		case <-stream.Context().Done():
			hs.logger.Info("Health events stream ended")
			return nil
		}
	}
}

// generateHealthEvents generates health events
func (hs *HealthService) generateHealthEvents(eventChan chan<- *peervault.HealthEvent, ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get current health status
			status := hs.healthChecker.GetHealthStatus()

			event := &peervault.HealthEvent{
				EventType: "health_check",
				Component: "system",
				Timestamp: time.Now(),
				Status:    status.Status,
				Message:   "Periodic health check",
				Metadata: map[string]string{
					"version": status.Version,
				},
			}

			// Add metadata from health status (if available)
			// Note: HealthResponse doesn't have Metadata field in current proto definition

			select {
			case eventChan <- event:
			default:
				// Channel is full, skip this event
				hs.logger.Warn("Health event channel full, dropping event")
			}

		case <-ctx.Done():
			return
		}
	}
}
