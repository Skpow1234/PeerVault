package security

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecurityTesterCreation tests the creation of a security tester
func TestSecurityTesterCreation(t *testing.T) {
	config := &SecurityTestConfig{
		BaseURL:     "http://localhost:3000",
		Timeout:     30 * time.Second,
		EnableOWASP: true,
	}

	tester := NewSecurityTester(config, slog.Default())
	require.NotNil(t, tester)
	assert.Equal(t, config, tester.config)
}

// TestDefaultSecurityTestConfig tests the default configuration
func TestDefaultSecurityTestConfig(t *testing.T) {
	config := DefaultSecurityTestConfig()
	require.NotNil(t, config)

	assert.Equal(t, "http://localhost:3000", config.BaseURL)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.NotNil(t, config.Headers)
	assert.NotNil(t, config.TestPayloads)
	assert.True(t, config.EnableOWASP)
	assert.True(t, config.EnableCustom)
}

// TestVulnerabilityCreation tests vulnerability creation
func TestVulnerabilityCreation(t *testing.T) {
	vuln := Vulnerability{
		Type:        "SQL Injection",
		Severity:    "High",
		Description: "SQL injection vulnerability detected",
		Endpoint:    "/api/users",
		Method:      "POST",
		Payload:     "'; DROP TABLE users; --",
		Response:    "Database error",
		Remediation: "Use parameterized queries",
	}

	assert.Equal(t, "SQL Injection", vuln.Type)
	assert.Equal(t, "High", vuln.Severity)
	assert.Equal(t, "SQL injection vulnerability detected", vuln.Description)
}

// TestSecurityTestCreation tests security test creation
func TestSecurityTestCreation(t *testing.T) {
	test := SecurityTest{
		Name:        "XSS Test",
		Type:        "Cross-Site Scripting",
		Endpoint:    "/api/search",
		Method:      "GET",
		Payload:     "<script>alert('xss')</script>",
		Expected:    "No script execution",
		Actual:      "Script executed",
		Passed:      false,
		Severity:    "Medium",
		Description: "XSS vulnerability detected",
	}

	assert.Equal(t, "XSS Test", test.Name)
	assert.Equal(t, "Cross-Site Scripting", test.Type)
	assert.False(t, test.Passed)
}

// TestSecurityTestPayloads tests security test payloads
func TestSecurityTestPayloads(t *testing.T) {
	config := DefaultSecurityTestConfig()

	// Test that default payloads are loaded
	assert.NotEmpty(t, config.TestPayloads)

	// Test specific payload types
	payloads := config.TestPayloads
	xssPayloads := 0
	sqlPayloads := 0

	for _, payload := range payloads {
		if strings.Contains(payload, "<script>") {
			xssPayloads++
		}
		if strings.Contains(payload, "DROP TABLE") || strings.Contains(payload, "UNION") {
			sqlPayloads++
		}
	}

	// Should have at least one of each type
	assert.Greater(t, xssPayloads, 0, "Should have XSS payloads")
	assert.Greater(t, sqlPayloads, 0, "Should have SQL injection payloads")
}

// TestSecurityTestWithContext tests security testing with context
func TestSecurityTestWithContext(t *testing.T) {
	config := &SecurityTestConfig{
		BaseURL:     "http://localhost:3000",
		Timeout:     1 * time.Second,
		EnableOWASP: true,
	}

	tester := NewSecurityTester(config, slog.Default())
	ctx := context.Background()

	// This tests the basic structure - current implementation doesn't make actual requests
	result, err := tester.RunSecurityTests(ctx, []string{"/health", "/api/nodes"})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
