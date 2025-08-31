package endpoints

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/Skpow1234/Peervault/internal/api/rest/services"
	"github.com/Skpow1234/Peervault/internal/api/rest/types"
	"github.com/Skpow1234/Peervault/internal/api/rest/types/requests"
)

type FileEndpoints struct {
	fileService services.FileService
	logger      *slog.Logger
}

func NewFileEndpoints(fileService services.FileService, logger *slog.Logger) *FileEndpoints {
	return &FileEndpoints{
		fileService: fileService,
		logger:      logger,
	}
}

func (e *FileEndpoints) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := e.fileService.ListFiles(r.Context())
	if err != nil {
		e.logger.Error("Failed to list files", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := types.FilesToResponse(files)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (e *FileEndpoints) HandleGetFile(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	file, err := e.fileService.GetFile(r.Context(), key)
	if err != nil {
		e.logger.Error("Failed to get file", "key", key, "error", err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	response := types.FileToResponse(file)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (e *FileEndpoints) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			e.logger.Error("Failed to close file", "error", err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		e.logger.Error("Failed to read file data", "error", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	metadataStr := r.FormValue("metadata")
	var metadata map[string]string
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			http.Error(w, "Invalid metadata format", http.StatusBadRequest)
			return
		}
	}

	uploadedFile, err := e.fileService.UploadFile(r.Context(), header.Filename, data, header.Header.Get("Content-Type"), metadata)
	if err != nil {
		e.logger.Error("Failed to upload file", "error", err)
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	response := types.FileToResponse(uploadedFile)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (e *FileEndpoints) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	err := e.fileService.DeleteFile(r.Context(), key)
	if err != nil {
		e.logger.Error("Failed to delete file", "key", key, "error", err)
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (e *FileEndpoints) HandleUpdateFileMetadata(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	var request requests.FileMetadataUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	file, err := e.fileService.UpdateFileMetadata(r.Context(), key, request.Metadata)
	if err != nil {
		e.logger.Error("Failed to update file metadata", "key", key, "error", err)
		http.Error(w, "Failed to update file metadata", http.StatusInternalServerError)
		return
	}

	response := types.FileToResponse(file)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
