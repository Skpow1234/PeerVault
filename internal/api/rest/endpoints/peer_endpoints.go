package endpoints

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Skpow1234/Peervault/internal/api/rest/services"
	"github.com/Skpow1234/Peervault/internal/api/rest/types"
	"github.com/Skpow1234/Peervault/internal/api/rest/types/requests"
)

type PeerEndpoints struct {
	peerService services.PeerService
	logger      *slog.Logger
}

func NewPeerEndpoints(peerService services.PeerService, logger *slog.Logger) *PeerEndpoints {
	return &PeerEndpoints{
		peerService: peerService,
		logger:      logger,
	}
}

func (e *PeerEndpoints) HandleListPeers(w http.ResponseWriter, r *http.Request) {
	peers, err := e.peerService.ListPeers(r.Context())
	if err != nil {
		e.logger.Error("Failed to list peers", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := types.PeersToResponse(peers)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (e *PeerEndpoints) HandleGetPeer(w http.ResponseWriter, r *http.Request) {
	peerID := r.URL.Query().Get("id")
	if peerID == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	peer, err := e.peerService.GetPeer(r.Context(), peerID)
	if err != nil {
		e.logger.Error("Failed to get peer", "id", peerID, "error", err)
		http.Error(w, "Peer not found", http.StatusNotFound)
		return
	}

	response := types.PeerToResponse(peer)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (e *PeerEndpoints) HandleAddPeer(w http.ResponseWriter, r *http.Request) {
	var request requests.PeerAddRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	peer := types.AddRequestToPeer(&request)
	addedPeer, err := e.peerService.AddPeer(r.Context(), *peer)
	if err != nil {
		e.logger.Error("Failed to add peer", "error", err)
		http.Error(w, "Failed to add peer", http.StatusInternalServerError)
		return
	}

	response := types.PeerToResponse(addedPeer)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (e *PeerEndpoints) HandleRemovePeer(w http.ResponseWriter, r *http.Request) {
	peerID := r.URL.Query().Get("id")
	if peerID == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	err := e.peerService.RemovePeer(r.Context(), peerID)
	if err != nil {
		e.logger.Error("Failed to remove peer", "id", peerID, "error", err)
		http.Error(w, "Failed to remove peer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
