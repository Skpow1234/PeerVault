package types

import (
	"time"

	"github.com/Skpow1234/Peervault/internal/api/rest/types/requests"
	"github.com/Skpow1234/Peervault/internal/api/rest/types/responses"
)

// FileToResponse converts a File entity to FileResponse
func FileToResponse(file *File) *responses.FileResponse {
	if file == nil {
		return nil
	}

	replicas := make([]responses.FileReplicaResponse, len(file.Replicas))
	for i, replica := range file.Replicas {
		replicas[i] = responses.FileReplicaResponse{
			PeerID:    replica.PeerID,
			Status:    replica.Status,
			CreatedAt: replica.CreatedAt,
		}
	}

	return &responses.FileResponse{
		Key:         file.Key,
		Name:        file.Name,
		Size:        file.Size,
		ContentType: file.ContentType,
		Hash:        file.Hash,
		CreatedAt:   file.CreatedAt,
		UpdatedAt:   file.UpdatedAt,
		Metadata:    file.Metadata,
		Replicas:    replicas,
	}
}

// FilesToResponse converts a slice of File entities to FileListResponse
func FilesToResponse(files []File) *responses.FileListResponse {
	fileResponses := make([]responses.FileResponse, len(files))
	for i, file := range files {
		response := FileToResponse(&file)
		if response != nil {
			fileResponses[i] = *response
		}
	}

	return &responses.FileListResponse{
		Files: fileResponses,
		Total: len(fileResponses),
	}
}

// ResponseToFile converts a FileResponse to File entity
func ResponseToFile(response *responses.FileResponse) *File {
	if response == nil {
		return nil
	}

	replicas := make([]FileReplica, len(response.Replicas))
	for i, replica := range response.Replicas {
		replicas[i] = FileReplica{
			PeerID:    replica.PeerID,
			Status:    replica.Status,
			CreatedAt: replica.CreatedAt,
		}
	}

	return &File{
		Key:         response.Key,
		Name:        response.Name,
		Size:        response.Size,
		ContentType: response.ContentType,
		Hash:        response.Hash,
		CreatedAt:   response.CreatedAt,
		UpdatedAt:   response.UpdatedAt,
		Metadata:    response.Metadata,
		Replicas:    replicas,
	}
}

// PeerToResponse converts a Peer entity to PeerResponse
func PeerToResponse(peer *Peer) *responses.PeerResponse {
	if peer == nil {
		return nil
	}

	return &responses.PeerResponse{
		ID:        peer.ID,
		Address:   peer.Address,
		Port:      peer.Port,
		Status:    peer.Status,
		LastSeen:  peer.LastSeen,
		CreatedAt: peer.CreatedAt,
		Metadata:  peer.Metadata,
	}
}

// PeersToResponse converts a slice of Peer entities to PeerListResponse
func PeersToResponse(peers []Peer) *responses.PeerListResponse {
	peerResponses := make([]responses.PeerResponse, len(peers))
	for i, peer := range peers {
		response := PeerToResponse(&peer)
		if response != nil {
			peerResponses[i] = *response
		}
	}

	return &responses.PeerListResponse{
		Peers: peerResponses,
		Total: len(peerResponses),
	}
}

// AddRequestToPeer converts a PeerAddRequest to Peer entity
func AddRequestToPeer(request *requests.PeerAddRequest) *Peer {
	if request == nil {
		return nil
	}

	return &Peer{
		Address:   request.Address,
		Port:      request.Port,
		Status:    "active",
		LastSeen:  time.Now(),
		CreatedAt: time.Now(),
		Metadata:  request.Metadata,
	}
}

// SystemInfoToResponse converts a SystemInfo entity to SystemInfoResponse
func SystemInfoToResponse(info *SystemInfo) *responses.SystemInfoResponse {
	if info == nil {
		return nil
	}

	return &responses.SystemInfoResponse{
		Version:      info.Version,
		Uptime:       info.Uptime,
		StartTime:    info.StartTime,
		StorageUsed:  info.StorageUsed,
		StorageTotal: info.StorageTotal,
		PeerCount:    info.PeerCount,
		FileCount:    info.FileCount,
	}
}

// MetricsToResponse converts a Metrics entity to MetricsResponse
func MetricsToResponse(metrics *Metrics) *responses.MetricsResponse {
	if metrics == nil {
		return nil
	}

	return &responses.MetricsResponse{
		RequestsTotal:     metrics.RequestsTotal,
		RequestsPerMinute: metrics.RequestsPerMinute,
		ActiveConnections: metrics.ActiveConnections,
		StorageUsage:      metrics.StorageUsage,
		LastUpdated:       metrics.LastUpdated,
	}
}
