package main

import (
	"testing"

	"github.com/Skpow1234/Peervault/internal/storage"
)

func TestWindowsPortability(t *testing.T) {
	// Test that storage roots are properly sanitized for Windows
	tests := []struct {
		listenAddr string
		expected   string
	}{
		{":3000", "node3000_network"},
		{":7000", "node7000_network"},
		{":5000", "node5000_network"},
		{":8080", "node8080_network"},
		{":", "node_network"},
		{"", "node_network"},
	}

	for _, tt := range tests {
		result := storage.SanitizeStorageRootFromAddr(tt.listenAddr)
		if result != tt.expected {
			t.Errorf("SanitizeStorageRootFromAddr(%q) = %q, want %q", tt.listenAddr, result, tt.expected)
		}
	}
}

func TestWindowsPathSanitization(t *testing.T) {
	// Test that invalid Windows characters are properly handled
	sanitizer := storage.NewWindowsPathSanitizer()

	// Test that colons are replaced
	result := sanitizer.SanitizePath(":3000_network")
	if result == ":3000_network" {
		t.Errorf("Path should not contain colons: %q", result)
	}

	// Test that other invalid characters are replaced
	result = sanitizer.SanitizePath("invalid<path>with:chars")
	if result == "invalid<path>with:chars" {
		t.Errorf("Path should not contain invalid characters: %q", result)
	}

	// Test that valid paths remain unchanged
	validPath := "valid_path_123"
	result = sanitizer.SanitizePath(validPath)
	if result != validPath {
		t.Errorf("Valid path should remain unchanged: %q != %q", result, validPath)
	}
}

func TestStorageRootGeneration(t *testing.T) {
	// Test the actual storage root generation used in main.go
	listenAddr := ":3000"
	expected := "node3000_network"

	result := storage.SanitizeStorageRootFromAddr(listenAddr)
	if result != expected {
		t.Errorf("Storage root generation failed: %q != %q", result, expected)
	}
}