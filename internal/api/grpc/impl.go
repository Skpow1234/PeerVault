package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Skpow1234/Shugo/internal/api/grpc/services"
	"github.com/Skpow1234/Shugo/proto/shugo"
)

// ShugoServiceImpl implements the ShugoService gRPC interface
type ShugoServiceImpl struct {
	shugo.UnimplementedShugoServiceServer
	fileService   *services.FileService
	peerService   *services.PeerService
	systemService *services.SystemService
	logger        *slog.Logger
}

// NewShugoServiceImpl creates a new service implementation
func NewShugoServiceImpl(logger *slog.Logger) *ShugoServiceImpl {
	if logger == nil {
		logger = slog.Default()
	}

	return &ShugoServiceImpl{
		fileService:   services.NewFileService(nil),
		peerService:   services.NewPeerService(),
		systemService: services.NewSystemService(),
		logger:        logger,
	}
}

// File Operations

// UploadFile implements streaming file upload
func (s *ShugoServiceImpl) UploadFile(stream shugo.ShugoService_UploadFileServer) error {
	var fileKey string
	var fileData []byte
	var totalSize int64

	s.logger.Info("Starting file upload stream")

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			s.logger.Error("Error receiving file chunk", "error", err)
			return status.Error(codes.Internal, "failed to receive file chunk")
		}

		if fileKey == "" {
			fileKey = chunk.FileKey
		}

		fileData = append(fileData, chunk.Data...)
		totalSize += int64(len(chunk.Data))

		s.logger.Debug("Received file chunk",
			"file_key", fileKey,
			"chunk_size", len(chunk.Data),
			"total_size", totalSize)
	}

	// Process the uploaded file
	response, err := s.fileService.UploadFile(fileKey, fileData)
	if err != nil {
		s.logger.Error("Error uploading file", "file_key", fileKey, "error", err)
		return status.Error(codes.Internal, "failed to upload file")
	}

	s.logger.Info("File upload completed", "file_key", fileKey, "size", totalSize)
	return stream.SendAndClose(response)
}

// DownloadFile implements streaming file download
func (s *ShugoServiceImpl) DownloadFile(req *shugo.FileRequest, stream shugo.ShugoService_DownloadFileServer) error {
	s.logger.Info("Starting file download", "file_key", req.Key)

	fileData, err := s.fileService.DownloadFile(req.Key)
	if err != nil {
		s.logger.Error("Error downloading file", "file_key", req.Key, "error", err)
		return status.Error(codes.NotFound, "file not found")
	}

	// Send file in chunks
	chunkSize := 64 * 1024 // 64KB chunks
	offset := int64(0)

	for offset < int64(len(fileData)) {
		end := offset + int64(chunkSize)
		if end > int64(len(fileData)) {
			end = int64(len(fileData))
		}

		chunk := &shugo.FileChunk{
			FileKey:  req.Key,
			Data:     fileData[offset:end],
			Offset:   offset,
			IsLast:   end == int64(len(fileData)),
			Checksum: fmt.Sprintf("%x", fileData[offset:end]), // Simple checksum
		}

		if err := stream.Send(chunk); err != nil {
			s.logger.Error("Error sending file chunk", "file_key", req.Key, "error", err)
			return status.Error(codes.Internal, "failed to send file chunk")
		}

		offset = end
	}

	s.logger.Info("File download completed", "file_key", req.Key, "size", len(fileData))
	return nil
}

// ListFiles implements file listing with pagination
func (s *ShugoServiceImpl) ListFiles(ctx context.Context, req *shugo.ListFilesRequest) (*shugo.ListFilesResponse, error) {
	s.logger.Info("Listing files", "page", req.Page, "page_size", req.PageSize, "filter", req.Filter)

	response, err := s.fileService.ListFiles(int(req.Page), int(req.PageSize), req.Filter)
	if err != nil {
		s.logger.Error("Error listing files", "error", err)
		return nil, status.Error(codes.Internal, "failed to list files")
	}

	return response, nil
}

// GetFile implements file metadata retrieval
func (s *ShugoServiceImpl) GetFile(ctx context.Context, req *shugo.FileRequest) (*shugo.FileResponse, error) {
	s.logger.Info("Getting file", "file_key", req.Key)

	response, err := s.fileService.GetFile(req.Key)
	if err != nil {
		s.logger.Error("Error getting file", "file_key", req.Key, "error", err)
		return nil, status.Error(codes.NotFound, "file not found")
	}

	return response, nil
}

// DeleteFile implements file deletion
func (s *ShugoServiceImpl) DeleteFile(ctx context.Context, req *shugo.FileRequest) (*shugo.DeleteFileResponse, error) {
	s.logger.Info("Deleting file", "file_key", req.Key)

	success, err := s.fileService.DeleteFile(req.Key)
	if err != nil {
		s.logger.Error("Error deleting file", "file_key", req.Key, "error", err)
		return nil, status.Error(codes.Internal, "failed to delete file")
	}

	return &shugo.DeleteFileResponse{
		Success: success,
		Message: "File deleted successfully",
	}, nil
}

// UpdateFileMetadata implements file metadata updates
func (s *ShugoServiceImpl) UpdateFileMetadata(ctx context.Context, req *shugo.UpdateFileMetadataRequest) (*shugo.FileResponse, error) {
	s.logger.Info("Updating file metadata", "file_key", req.Key)

	response, err := s.fileService.UpdateFileMetadata(req.Key, req.Metadata)
	if err != nil {
		s.logger.Error("Error updating file metadata", "file_key", req.Key, "error", err)
		return nil, status.Error(codes.NotFound, "file not found")
	}

	return response, nil
}

// Peer Operations

// ListPeers implements peer listing
func (s *ShugoServiceImpl) ListPeers(ctx context.Context, req *emptypb.Empty) (*shugo.ListPeersResponse, error) {
	s.logger.Info("Listing peers")

	response, err := s.peerService.ListPeers()
	if err != nil {
		s.logger.Error("Error listing peers", "error", err)
		return nil, status.Error(codes.Internal, "failed to list peers")
	}

	return response, nil
}

// GetPeer implements peer retrieval
func (s *ShugoServiceImpl) GetPeer(ctx context.Context, req *shugo.PeerRequest) (*shugo.PeerResponse, error) {
	s.logger.Info("Getting peer", "peer_id", req.Id)

	response, err := s.peerService.GetPeer(req.Id)
	if err != nil {
		s.logger.Error("Error getting peer", "peer_id", req.Id, "error", err)
		return nil, status.Error(codes.NotFound, "peer not found")
	}

	return response, nil
}

// AddPeer implements peer addition
func (s *ShugoServiceImpl) AddPeer(ctx context.Context, req *shugo.AddPeerRequest) (*shugo.PeerResponse, error) {
	s.logger.Info("Adding peer", "address", req.Address, "port", req.Port)

	response, err := s.peerService.AddPeer(req.Address, int(req.Port), req.Metadata)
	if err != nil {
		s.logger.Error("Error adding peer", "address", req.Address, "port", req.Port, "error", err)
		return nil, status.Error(codes.Internal, "failed to add peer")
	}

	return response, nil
}

// RemovePeer implements peer removal
func (s *ShugoServiceImpl) RemovePeer(ctx context.Context, req *shugo.PeerRequest) (*shugo.RemovePeerResponse, error) {
	s.logger.Info("Removing peer", "peer_id", req.Id)

	success, err := s.peerService.RemovePeer(req.Id)
	if err != nil {
		s.logger.Error("Error removing peer", "peer_id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "failed to remove peer")
	}

	return &shugo.RemovePeerResponse{
		Success: success,
		Message: "Peer removed successfully",
	}, nil
}

// GetPeerHealth implements peer health checking
func (s *ShugoServiceImpl) GetPeerHealth(ctx context.Context, req *shugo.PeerRequest) (*shugo.PeerHealthResponse, error) {
	s.logger.Info("Getting peer health", "peer_id", req.Id)

	response, err := s.peerService.GetPeerHealth(req.Id)
	if err != nil {
		s.logger.Error("Error getting peer health", "peer_id", req.Id, "error", err)
		return nil, status.Error(codes.NotFound, "peer not found")
	}

	return response, nil
}

// System Operations

// GetSystemInfo implements system information retrieval
func (s *ShugoServiceImpl) GetSystemInfo(ctx context.Context, req *emptypb.Empty) (*shugo.SystemInfoResponse, error) {
	s.logger.Info("Getting system info")

	response, err := s.systemService.GetSystemInfo()
	if err != nil {
		s.logger.Error("Error getting system info", "error", err)
		return nil, status.Error(codes.Internal, "failed to get system info")
	}

	return response, nil
}

// GetMetrics implements metrics retrieval
func (s *ShugoServiceImpl) GetMetrics(ctx context.Context, req *emptypb.Empty) (*shugo.MetricsResponse, error) {
	s.logger.Info("Getting metrics")

	response, err := s.systemService.GetMetrics()
	if err != nil {
		s.logger.Error("Error getting metrics", "error", err)
		return nil, status.Error(codes.Internal, "failed to get metrics")
	}

	return response, nil
}

// HealthCheck implements health checking
func (s *ShugoServiceImpl) HealthCheck(ctx context.Context, req *emptypb.Empty) (*shugo.HealthResponse, error) {
	s.logger.Info("Health check requested")

	response, err := s.systemService.HealthCheck()
	if err != nil {
		s.logger.Error("Error in health check", "error", err)
		return nil, status.Error(codes.Internal, "health check failed")
	}

	return response, nil
}

// Streaming Operations

// StreamFileOperations implements file operation event streaming
func (s *ShugoServiceImpl) StreamFileOperations(req *emptypb.Empty, stream shugo.ShugoService_StreamFileOperationsServer) error {
	s.logger.Info("Starting file operations stream")

	// Create a channel for file operation events
	eventChan := make(chan *shugo.FileOperationEvent, 100)
	defer close(eventChan)

	// Start a goroutine to generate mock events
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				event := &shugo.FileOperationEvent{
					EventType: "periodic_check",
					FileKey:   "system",
					PeerId:    "local",
					Timestamp: timestamppb.Now(),
					Metadata: map[string]string{
						"type": "periodic",
					},
				}
				select {
				case eventChan <- event:
				default:
					// Channel is full, skip this event
				}
			case <-stream.Context().Done():
				return
			}
		}
	}()

	// Send events to the client
	for {
		select {
		case event := <-eventChan:
			if err := stream.Send(event); err != nil {
				s.logger.Error("Error sending file operation event", "error", err)
				return status.Error(codes.Internal, "failed to send event")
			}
		case <-stream.Context().Done():
			s.logger.Info("File operations stream ended")
			return nil
		}
	}
}

// StreamPeerEvents implements peer event streaming
func (s *ShugoServiceImpl) StreamPeerEvents(req *emptypb.Empty, stream shugo.ShugoService_StreamPeerEventsServer) error {
	s.logger.Info("Starting peer events stream")

	// Create a channel for peer events
	eventChan := make(chan *shugo.PeerEvent, 100)
	defer close(eventChan)

	// Start a goroutine to generate mock events
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				event := &shugo.PeerEvent{
					EventType: "health_check",
					PeerId:    "system",
					Timestamp: timestamppb.Now(),
					Metadata: map[string]string{
						"type": "periodic",
					},
				}
				select {
				case eventChan <- event:
				default:
					// Channel is full, skip this event
				}
			case <-stream.Context().Done():
				return
			}
		}
	}()

	// Send events to the client
	for {
		select {
		case event := <-eventChan:
			if err := stream.Send(event); err != nil {
				s.logger.Error("Error sending peer event", "error", err)
				return status.Error(codes.Internal, "failed to send event")
			}
		case <-stream.Context().Done():
			s.logger.Info("Peer events stream ended")
			return nil
		}
	}
}

// StreamSystemEvents implements system event streaming
func (s *ShugoServiceImpl) StreamSystemEvents(req *emptypb.Empty, stream shugo.ShugoService_StreamSystemEventsServer) error {
	s.logger.Info("Starting system events stream")

	// Create a channel for system events
	eventChan := make(chan *shugo.SystemEvent, 100)
	defer close(eventChan)

	// Start a goroutine to generate mock events
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				event := &shugo.SystemEvent{
					EventType: "status_update",
					Component: "grpc_server",
					Timestamp: timestamppb.Now(),
					Message:   "System running normally",
					Metadata: map[string]string{
						"uptime": time.Since(time.Now()).String(),
					},
				}
				select {
				case eventChan <- event:
				default:
					// Channel is full, skip this event
				}
			case <-stream.Context().Done():
				return
			}
		}
	}()

	// Send events to the client
	for {
		select {
		case event := <-eventChan:
			if err := stream.Send(event); err != nil {
				s.logger.Error("Error sending system event", "error", err)
				return status.Error(codes.Internal, "failed to send event")
			}
		case <-stream.Context().Done():
			s.logger.Info("System events stream ended")
			return nil
		}
	}
}
