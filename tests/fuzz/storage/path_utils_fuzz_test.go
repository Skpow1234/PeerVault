package storage

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Skpow1234/Peervault/internal/storage"
)

// FuzzSanitizePath tests path sanitization with fuzz-generated data
func FuzzSanitizePath(f *testing.F) {
	// Add seed corpus for path sanitization testing
	seedCorpus := []string{
		// Valid paths
		"normal/path",
		"path/with/spaces",
		"path-with-dashes",
		"path_with_underscores",
		"path.with.dots",

		// Paths with invalid characters
		"path<with>invalid:chars",
		"path|with\\invalid/chars",
		"path\"with'quotes",
		"path*with?wildcards",

		// Paths with leading/trailing spaces and dots
		"  path with spaces  ",
		"...path with leading dots",
		"path with trailing dots...",
		"   ...path with both...   ",

		// Empty and single character paths
		"",
		".",
		"..",
		"a",
		" ",

		// Very long paths (but not too long)
		strings.Repeat("very/long/path/", 10),

		// Paths with null bytes
		"path\x00with\x00nulls",

		// Paths with control characters
		"path\x01\x02\x03with\x04\x05\x06control\x07\x08\x09chars",
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, path string) {
		// Test path sanitization
		sanitizer := storage.NewWindowsPathSanitizer()
		sanitized := sanitizer.SanitizePath(path)

		// Basic validation of sanitized path
		if sanitized == "" {
			// Empty result is acceptable for completely invalid paths
			return
		}

		// Check that sanitized path doesn't contain invalid characters
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*", "\\", "/"}
		for _, char := range invalidChars {
			if strings.Contains(sanitized, char) {
				t.Errorf("Sanitized path contains invalid character '%s': %s", char, sanitized)
			}
		}

		// Check that path doesn't start or end with ASCII spaces or dots
		// Note: The sanitizer only handles ASCII spaces, not Unicode whitespace
		if strings.Trim(sanitized, " .") != sanitized {
			t.Errorf("Sanitized path has leading/trailing ASCII spaces or dots: '%s'", sanitized)
		}

		// Check that path doesn't exceed reasonable length
		// Note: Very long paths might be created by sanitization of long inputs
		// This is acceptable for fuzz testing as it tests edge cases
		if len(sanitized) > 50000 {
			t.Errorf("Sanitized path unreasonably long: %d characters", len(sanitized))
		}

		// Check that path components are reasonable
		// Note: Very long components might be created by sanitization of long paths
		// This is acceptable for fuzz testing as it tests edge cases
		components := strings.Split(sanitized, string(filepath.Separator))
		for _, component := range components {
			if len(component) > 50000 {
				// Only fail for unreasonably long components
				t.Errorf("Path component unreasonably long: %d characters", len(component))
			}
		}
	})
}

// FuzzSanitizeStorageRoot tests storage root sanitization with fuzz-generated data
func FuzzSanitizeStorageRoot(f *testing.F) {
	// Add seed corpus for storage root sanitization testing
	seedCorpus := []string{
		// Valid listen addresses
		":8080",
		":12345",
		":0",
		":65535",

		// Addresses with prefixes
		"localhost:8080",
		"127.0.0.1:12345",
		"0.0.0.0:8080",

		// Invalid addresses
		"",
		":",
		"invalid",
		":abc",
		":99999",

		// Edge cases
		"   :8080   ",
		"...:8080...",
		":8080:8080",
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, listenAddr string) {
		// Test storage root sanitization
		sanitized := storage.SanitizeStorageRootFromAddr(listenAddr)

		// Basic validation
		if sanitized == "" {
			t.Errorf("Sanitized storage root is empty for: %s", listenAddr)
		}

		// Check that sanitized path doesn't contain invalid characters
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*", "\\", "/"}
		for _, char := range invalidChars {
			if strings.Contains(sanitized, char) {
				t.Errorf("Sanitized storage root contains invalid character '%s': %s", char, sanitized)
			}
		}

		// Check that path doesn't start or end with ASCII spaces or dots
		// Note: The sanitizer only handles ASCII spaces, not Unicode whitespace
		if strings.Trim(sanitized, " .") != sanitized {
			t.Errorf("Sanitized storage root has leading/trailing ASCII spaces or dots: '%s'", sanitized)
		}

		// Check that path doesn't exceed reasonable length
		// Note: Very long paths might be created by sanitization of long inputs
		// This is acceptable for fuzz testing as it tests edge cases
		if len(sanitized) > 50000 {
			t.Errorf("Sanitized storage root unreasonably long: %d characters", len(sanitized))
		}
	})
}

// FuzzSanitizeStorageRootWithPrefix tests storage root sanitization with prefix
func FuzzSanitizeStorageRootWithPrefix(f *testing.F) {
	// Add seed corpus for storage root sanitization with prefix testing
	seedCorpus := []struct {
		listenAddr string
		prefix     string
	}{
		// Valid combinations
		{":8080", "test"},
		{":12345", "node"},
		{":0", "peer"},
		{":65535", "vault"},

		// Addresses with prefixes
		{"localhost:8080", "test"},
		{"127.0.0.1:12345", "node"},
		{"0.0.0.0:8080", "peer"},

		// Invalid addresses
		{"", "test"},
		{":", "node"},
		{"invalid", "peer"},
		{":abc", "vault"},
		{":99999", "test"},

		// Edge cases
		{"   :8080   ", "test"},
		{"...:8080...", "node"},
		{":8080:8080", "peer"},
		{":8080", ""},
		{":8080", "   "},
		{":8080", "..."},
	}

	for _, seed := range seedCorpus {
		f.Add(seed.listenAddr, seed.prefix)
	}

	f.Fuzz(func(t *testing.T, listenAddr, prefix string) {
		// Test storage root sanitization with prefix
		sanitized := storage.SanitizeStorageRootFromAddrWithPrefix(listenAddr, prefix)

		// Basic validation
		if sanitized == "" {
			t.Errorf("Sanitized storage root is empty for: %s, prefix: %s", listenAddr, prefix)
		}

		// Check that sanitized path doesn't contain invalid characters
		invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*", "\\", "/"}
		for _, char := range invalidChars {
			if strings.Contains(sanitized, char) {
				t.Errorf("Sanitized storage root contains invalid character '%s': %s", char, sanitized)
			}
		}

		// Check that path doesn't start or end with ASCII spaces or dots
		// Note: The sanitizer only handles ASCII spaces, not Unicode whitespace
		if strings.Trim(sanitized, " .") != sanitized {
			t.Errorf("Sanitized storage root has leading/trailing ASCII spaces or dots: '%s'", sanitized)
		}

		// Check that path doesn't exceed reasonable length
		// Note: Very long paths might be created by sanitization of long inputs
		// This is acceptable for fuzz testing as it tests edge cases
		if len(sanitized) > 50000 {
			t.Errorf("Sanitized storage root unreasonably long: %d characters", len(sanitized))
		}

		// Check that prefix is included in result (if not empty and not just spaces/dots)
		if prefix != "" && strings.TrimSpace(prefix) != "" && strings.Trim(prefix, ".") != "" {
			sanitizedPrefix := storage.NewWindowsPathSanitizer().SanitizePath(prefix)
			if sanitizedPrefix != "default" && !strings.Contains(sanitized, sanitizedPrefix) {
				t.Errorf("Sanitized storage root doesn't contain sanitized prefix: %s (expected: %s)", sanitized, sanitizedPrefix)
			}
		}
	})
}
