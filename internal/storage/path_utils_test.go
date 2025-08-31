package storage

import (
	"strings"
	"testing"
)

func TestWindowsPathSanitizer_SanitizePath(t *testing.T) {
	sanitizer := NewWindowsPathSanitizer()

	tests := []struct {
		input    string
		expected string
	}{
		// Test basic invalid characters
		{":3000_network", "_3000_network"},
		{"<invalid>", "_invalid_"},
		{">invalid<", "_invalid_"},
		{`"invalid"`, "_invalid_"},
		{"invalid|path", "invalid_path"},
		{"invalid?path", "invalid_path"},
		{"invalid*path", "invalid_path"},
		{"invalid\\path", "invalid_path"},
		{"invalid/path", "invalid_path"},

		// Test multiple invalid characters
		{":3000:network:", "_3000_network_"},
		{"<invalid:path>", "_invalid_path_"},

		// Test leading/trailing spaces and dots
		{"  path  ", "path"},
		{"...path...", "path"},
		{" . path . ", "path"},

		// Test empty or problematic inputs
		{"", "default"},
		{"   ", "default"},
		{"...", "default"},
		{":", "_"},
		{"<>", "__"},

		// Test valid paths (should remain unchanged)
		{"valid_path", "valid_path"},
		{"valid-path", "valid-path"},
		{"valid.path", "valid.path"},
		{"valid123", "valid123"},
		{"VALID_PATH", "VALID_PATH"},
	}

	for _, tt := range tests {
		result := sanitizer.SanitizePath(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizePath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestWindowsPathSanitizer_SanitizeStorageRoot(t *testing.T) {
	sanitizer := NewWindowsPathSanitizer()

	tests := []struct {
		input    string
		expected string
	}{
		{":3000", "node3000_network"},
		{":7000", "node7000_network"},
		{":5000", "node5000_network"},
		{"3000", "node3000_network"},
		{"7000", "node7000_network"},
		{"5000", "node5000_network"},
		{":8080", "node8080_network"},
		{":", "node_network"},
		{"", "node_network"},
	}

	for _, tt := range tests {
		result := sanitizer.SanitizeStorageRoot(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeStorageRoot(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestWindowsPathSanitizer_SanitizeStorageRootWithPrefix(t *testing.T) {
	sanitizer := NewWindowsPathSanitizer()

	tests := []struct {
		listenAddr string
		prefix     string
		expected   string
	}{
		{":3000", "peervault", "peervault_3000_network"},
		{":7000", "test", "test_7000_network"},
		{":5000", "demo", "demo_5000_network"},
		{"3000", "peervault", "peervault_3000_network"},
		{":", "default", "default__network"},
		{"", "default", "default__network"},
	}

	for _, tt := range tests {
		result := sanitizer.SanitizeStorageRootWithPrefix(tt.listenAddr, tt.prefix)
		if result != tt.expected {
			t.Errorf("SanitizeStorageRootWithPrefix(%q, %q) = %q, want %q",
				tt.listenAddr, tt.prefix, result, tt.expected)
		}
	}
}

func TestDefaultPathSanitizer(t *testing.T) {
	// Test that the default sanitizer works correctly
	result := DefaultPathSanitizer.SanitizePath(":3000_network")
	expected := "_3000_network"
	if result != expected {
		t.Errorf("DefaultPathSanitizer.SanitizePath(\":3000_network\") = %q, want %q", result, expected)
	}
}

func TestSanitizeStorageRootFromAddr(t *testing.T) {
	// Test the convenience function
	result := SanitizeStorageRootFromAddr(":3000")
	expected := "node3000_network"
	if result != expected {
		t.Errorf("SanitizeStorageRootFromAddr(\":3000\") = %q, want %q", result, expected)
	}
}

func TestSanitizeStorageRootFromAddrWithPrefix(t *testing.T) {
	// Test the convenience function with prefix
	result := SanitizeStorageRootFromAddrWithPrefix(":3000", "peervault")
	expected := "peervault_3000_network"
	if result != expected {
		t.Errorf("SanitizeStorageRootFromAddrWithPrefix(\":3000\", \"peervault\") = %q, want %q", result, expected)
	}
}

func TestPathSanitizationEdgeCases(t *testing.T) {
	sanitizer := NewWindowsPathSanitizer()

	// Test very long paths
	longPath := strings.Repeat("a", 100) + ":" + strings.Repeat("b", 100)
	result := sanitizer.SanitizePath(longPath)
	if strings.Contains(result, ":") {
		t.Errorf("Sanitized path should not contain colons: %q", result)
	}

	// Test paths with only invalid characters
	result = sanitizer.SanitizePath(":::::")
	expected := "_____"
	if result != expected {
		t.Errorf("SanitizePath(\":::::\") = %q, want %q", result, expected)
	}

	// Test paths with mixed valid and invalid characters
	result = sanitizer.SanitizePath("valid:invalid:valid")
	expected = "valid_invalid_valid"
	if result != expected {
		t.Errorf("SanitizePath(\"valid:invalid:valid\") = %q, want %q", result, expected)
	}
}
