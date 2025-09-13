package versioning

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected APIVersion
		hasError bool
	}{
		{"v1.0.0", APIVersion{1, 0, 0}, false},
		{"1.1.0", APIVersion{1, 1, 0}, false},
		{"v2.0.0", APIVersion{2, 0, 0}, false},
		{"1.0", APIVersion{}, true},
		{"invalid", APIVersion{}, true},
		{"", APIVersion{}, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseVersion(test.input)
			if test.hasError {
				if err == nil {
					t.Errorf("expected error for input %s, got none", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %s: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("expected %v, got %v", test.expected, result)
				}
			}
		})
	}
}

func TestAPIVersionCompare(t *testing.T) {
	tests := []struct {
		v1       APIVersion
		v2       APIVersion
		expected int
	}{
		{APIVersion{1, 0, 0}, APIVersion{1, 0, 0}, 0},
		{APIVersion{1, 0, 0}, APIVersion{1, 0, 1}, -1},
		{APIVersion{1, 0, 1}, APIVersion{1, 0, 0}, 1},
		{APIVersion{1, 0, 0}, APIVersion{1, 1, 0}, -1},
		{APIVersion{1, 1, 0}, APIVersion{1, 0, 0}, 1},
		{APIVersion{1, 0, 0}, APIVersion{2, 0, 0}, -1},
		{APIVersion{2, 0, 0}, APIVersion{1, 0, 0}, 1},
	}

	for _, test := range tests {
		t.Run(test.v1.String()+" vs "+test.v2.String(), func(t *testing.T) {
			result := test.v1.Compare(test.v2)
			if result != test.expected {
				t.Errorf("expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestVersionConfig(t *testing.T) {
	config := NewVersionConfig()

	// Test supported versions
	if !config.IsVersionSupported(Version_1_0_0) {
		t.Error("v1.0.0 should be supported")
	}
	if !config.IsVersionSupported(Version_1_1_0) {
		t.Error("v1.1.0 should be supported")
	}
	if config.IsVersionSupported(APIVersion{2, 0, 0}) {
		t.Error("v2.0.0 should not be supported yet")
	}

	// Test deprecated versions
	if !config.IsVersionDeprecated(Version_1_0_0) {
		t.Error("v1.0.0 should be deprecated")
	}
	if config.IsVersionDeprecated(Version_1_1_0) {
		t.Error("v1.1.0 should not be deprecated")
	}
}

func TestNegotiateVersion(t *testing.T) {
	config := NewVersionConfig()

	tests := []struct {
		name     string
		headers  map[string]string
		url      string
		expected APIVersion
	}{
		{
			name:     "Accept-Version header",
			headers:  map[string]string{"Accept-Version": "v1.0.0"},
			url:      "/api/v1/files",
			expected: Version_1_0_0,
		},
		{
			name:     "Accept header with version",
			headers:  map[string]string{"Accept": "application/json; version=v1.1.0"},
			url:      "/api/v1/files",
			expected: Version_1_1_0,
		},
		{
			name:     "URL path version",
			headers:  map[string]string{},
			url:      "/api/v1/files",
			expected: CurrentVersion, // Current implementation uses /api/v1/ for all versions
		},
		{
			name:     "No version specified - defaults to current",
			headers:  map[string]string{},
			url:      "/api/v1/files",
			expected: CurrentVersion,
		},
		{
			name:     "Unsupported version - defaults to current",
			headers:  map[string]string{"Accept-Version": "v2.0.0"},
			url:      "/api/v1/files",
			expected: CurrentVersion,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", test.url, nil)
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			result := config.NegotiateVersion(req)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestVersionMiddleware(t *testing.T) {
	config := NewVersionConfig()
	middleware := VersionMiddleware(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version, ok := GetVersionFromContext(r.Context())
		if !ok {
			t.Error("version not found in context")
			return
		}
		w.Header().Set("X-Test-Version", version.String())
		w.WriteHeader(http.StatusOK)
	}))

	// Test with Accept-Version header
	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("Accept-Version", "v1.0.0")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("X-API-Version") != "v1.0.0" {
		t.Errorf("expected X-API-Version header to be v1.0.0, got %s", w.Header().Get("X-API-Version"))
	}
	if w.Header().Get("X-API-Current-Version") != CurrentVersion.String() {
		t.Errorf("expected X-API-Current-Version header to be %s, got %s", CurrentVersion.String(), w.Header().Get("X-API-Current-Version"))
	}
	if w.Header().Get("X-Test-Version") != "v1.0.0" {
		t.Errorf("expected X-Test-Version header to be v1.0.0, got %s", w.Header().Get("X-Test-Version"))
	}

	// Test deprecated version warning
	if w.Header().Get("X-API-Deprecated") != "true" {
		t.Error("expected X-API-Deprecated header to be true for deprecated version")
	}
	if w.Header().Get("Warning") == "" {
		t.Error("expected Warning header for deprecated version")
	}
}

func TestRequireVersion(t *testing.T) {
	ctx := context.Background()

	// Test without version in context
	err := RequireVersion(ctx, Version_1_0_0)
	if err == nil {
		t.Error("expected error when version not in context")
	}

	// Test with version in context
	ctx = context.WithValue(ctx, VersionContextKey{}, Version_1_1_0)

	// Should succeed for lower requirement
	err = RequireVersion(ctx, Version_1_0_0)
	if err != nil {
		t.Errorf("expected no error for sufficient version, got: %v", err)
	}

	// Should fail for higher requirement
	err = RequireVersion(ctx, APIVersion{2, 0, 0})
	if err == nil {
		t.Error("expected error for insufficient version")
	}
}
