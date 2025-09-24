package security

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/security"
)

func TestSecurityTests(t *testing.T) {
	config := &security.SecurityTestConfig{
		BaseURL: "http://localhost:3000",
		Timeout: 30 * time.Second,
		Headers: make(map[string]string),
		TestPayloads: []string{
			"<script>alert('xss')</script>",
			"'; DROP TABLE users; --",
			"../../../etc/passwd",
		},
		EnableOWASP:  true,
		EnableCustom: true,
	}

	logger := slog.Default()
	tester := security.NewSecurityTester(config, logger)

	endpoints := []string{
		"/health",
		"/api/status",
		"/api/auth/login",
	}

	ctx := context.Background()
	result, err := tester.RunSecurityTests(ctx, endpoints)
	if err != nil {
		t.Fatalf("Security tests failed: %v", err)
	}

	// Basic assertions
	if result.TotalTests == 0 {
		t.Error("Expected at least one test")
	}

	t.Logf("Security test results:")
	t.Logf("  Total tests: %d", result.TotalTests)
	t.Logf("  Passed tests: %d", result.PassedTests)
	t.Logf("  Failed tests: %d", result.FailedTests)
	t.Logf("  Vulnerabilities found: %d", len(result.Vulnerabilities))

	// Log vulnerabilities
	for i, vuln := range result.Vulnerabilities {
		t.Logf("  Vulnerability %d:", i+1)
		t.Logf("    Type: %s", vuln.Type)
		t.Logf("    Severity: %s", vuln.Severity)
		t.Logf("    Description: %s", vuln.Description)
		t.Logf("    Endpoint: %s %s", vuln.Method, vuln.Endpoint)
		t.Logf("    Remediation: %s", vuln.Remediation)
	}
}

func TestSecurityTestsWithMockServer(t *testing.T) {
	// This test would use a mock server for more reliable testing
	// For now, we'll skip if no server is available
	t.Skip("Skipping security tests with mock server - requires running server")
}
