package versioning

import (
	"fmt"
	"net/http"
)

// VersionedFeature demonstrates how to implement version-specific features
type VersionedFeature struct {
	config *VersionConfig
}

// NewVersionedFeature creates a new versioned feature handler
func NewVersionedFeature(config *VersionConfig) *VersionedFeature {
	return &VersionedFeature{config: config}
}

// HandleWebhookFeature demonstrates version-specific webhook handling
func (vf *VersionedFeature) HandleWebhookFeature(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	version, ok := GetVersionFromContext(ctx)
	if !ok {
		http.Error(w, "API version not found", http.StatusInternalServerError)
		return
	}

	// Webhooks are only available in v1.1.0+
	if version.Compare(Version_1_1_0) < 0 {
		http.Error(w, "Webhooks require API version v1.1.0 or higher", http.StatusBadRequest)
		return
	}

	// Handle webhook-specific logic for v1.1.0+
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, `{"message": "Webhook processed with API version %s"}`, version.String()); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// HandleLegacyFeature demonstrates backward compatibility
func (vf *VersionedFeature) HandleLegacyFeature(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	version, ok := GetVersionFromContext(ctx)
	if !ok {
		http.Error(w, "API version not found", http.StatusInternalServerError)
		return
	}

	// Simple response struct for example
	response := struct {
		Name        string            `json:"name"`
		Version     string            `json:"version"`
		Description string            `json:"description"`
		Endpoints   map[string]string `json:"endpoints"`
	}{
		Name:    "PeerVault REST API",
		Version: version.String(),
	}

	// Add version-specific fields
	if version.Compare(Version_1_1_0) >= 0 {
		response.Description = "A distributed file storage system with P2P capabilities (v1.1.0+ features)"
		response.Endpoints = map[string]string{
			"files":   "/api/v1/files",
			"peers":   "/api/v1/peers",
			"health":  "/health",
			"metrics": "/metrics",
			"system":  "/api/v1/system/info",
			"webhook": "/api/v1/webhook",
			"docs":    "/docs",
			"swagger": "/swagger.json",
		}
	} else {
		response.Description = "A distributed file storage system with P2P capabilities"
		response.Endpoints = map[string]string{
			"files":   "/api/v1/files",
			"peers":   "/api/v1/peers",
			"health":  "/health",
			"metrics": "/metrics",
			"system":  "/api/v1/system/info",
			"docs":    "/docs",
			"swagger": "/swagger.json",
		}
	}

	// Send JSON response (simplified for example)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, `{
		"name": "%s",
		"version": "%s",
		"description": "%s",
		"endpoints": %s
	}`, response.Name, response.Version, response.Description, formatEndpoints(response.Endpoints)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// formatEndpoints is a helper to format the endpoints map as JSON
func formatEndpoints(endpoints map[string]string) string {
	result := "{"
	for k, v := range endpoints {
		result += fmt.Sprintf(`"%s": "%s",`, k, v)
	}
	if len(result) > 1 {
		result = result[:len(result)-1] // Remove trailing comma
	}
	result += "}"
	return result
}

// VersionedRateLimit demonstrates version-aware rate limiting
func (vf *VersionedFeature) VersionedRateLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	version, ok := GetVersionFromContext(ctx)
	if !ok {
		http.Error(w, "API version not found", http.StatusInternalServerError)
		return
	}

	// Different rate limits for different versions
	var limit int
	switch {
	case version.Compare(Version_1_1_0) >= 0:
		limit = 200 // Higher limit for newer versions
	default:
		limit = 100 // Standard limit for older versions
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, `{"rate_limit": %d, "version": "%s"}`, limit, version.String()); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
