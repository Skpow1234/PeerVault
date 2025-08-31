package storage

import (
	"fmt"
	"regexp"
	"strings"
)

// WindowsPathSanitizer provides utilities for making paths Windows-safe
type WindowsPathSanitizer struct{}

// NewWindowsPathSanitizer creates a new Windows path sanitizer
func NewWindowsPathSanitizer() *WindowsPathSanitizer {
	return &WindowsPathSanitizer{}
}

// SanitizePath makes a path safe for Windows by replacing invalid characters
func (w *WindowsPathSanitizer) SanitizePath(path string) string {
	// Windows invalid characters: < > : " | ? * \ /
	// Replace them with safe alternatives
	invalidChars := regexp.MustCompile(`[<>:"|?*\\/]`)
	sanitized := invalidChars.ReplaceAllString(path, "_")

	// Remove leading/trailing spaces and dots (Windows doesn't allow these)
	sanitized = strings.Trim(sanitized, " .")

	// Ensure the path is not empty after sanitization
	if sanitized == "" {
		sanitized = "default"
	}

	return sanitized
}

// SanitizeStorageRoot creates a Windows-safe storage root from a listen address
func (w *WindowsPathSanitizer) SanitizeStorageRoot(listenAddr string) string {
	// Remove the colon prefix if present
	port := strings.TrimPrefix(listenAddr, ":")

	// Create a safe storage root name
	safeName := fmt.Sprintf("node%s_network", port)

	// Sanitize the final path
	return w.SanitizePath(safeName)
}

// SanitizeStorageRootWithPrefix creates a Windows-safe storage root with a custom prefix
func (w *WindowsPathSanitizer) SanitizeStorageRootWithPrefix(listenAddr, prefix string) string {
	// Remove the colon prefix if present
	port := strings.TrimPrefix(listenAddr, ":")

	// Create a safe storage root name with custom prefix
	safeName := fmt.Sprintf("%s_%s_network", prefix, port)

	// Sanitize the final path
	return w.SanitizePath(safeName)
}

// DefaultPathSanitizer is the default instance for easy access
var DefaultPathSanitizer = NewWindowsPathSanitizer()

// SanitizeStorageRootFromAddr is a convenience function to sanitize storage root from listen address
func SanitizeStorageRootFromAddr(listenAddr string) string {
	return DefaultPathSanitizer.SanitizeStorageRoot(listenAddr)
}

// SanitizeStorageRootFromAddrWithPrefix is a convenience function to sanitize storage root with prefix
func SanitizeStorageRootFromAddrWithPrefix(listenAddr, prefix string) string {
	return DefaultPathSanitizer.SanitizeStorageRootWithPrefix(listenAddr, prefix)
}
