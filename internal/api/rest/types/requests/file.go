package requests

// FileUploadRequest represents a file upload request
type FileUploadRequest struct {
	Name        string            `json:"name"`
	ContentType string            `json:"content_type"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FileMetadataUpdateRequest represents a file metadata update request
type FileMetadataUpdateRequest struct {
	Metadata map[string]string `json:"metadata"`
}
