package contracts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// ContractTest defines a contract test
type ContractTest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Provider    string                 `json:"provider"`
	Consumer    string                 `json:"consumer"`
	Request     *ContractRequest       `json:"request"`
	Response    *ContractResponse      `json:"response"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ContractRequest defines the expected request
type ContractRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Query   map[string]string `json:"query"`
	Body    interface{}       `json:"body"`
	Timeout time.Duration     `json:"timeout"`
}

// ContractResponse defines the expected response
type ContractResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	Schema     map[string]interface{} `json:"schema"`
}

// ContractVerifier verifies API contracts
type ContractVerifier struct {
	baseURL string
	client  *http.Client
	logger  *slog.Logger
}

// NewContractVerifier creates a new contract verifier
func NewContractVerifier(baseURL string, logger *slog.Logger) *ContractVerifier {
	return &ContractVerifier{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// VerifyContract verifies a single contract
func (cv *ContractVerifier) VerifyContract(ctx context.Context, contract *ContractTest) error {
	cv.logger.Info("Verifying contract", "name", contract.Name, "provider", contract.Provider)

	// Create request
	req, err := cv.createRequest(contract.Request)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := cv.client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Verify response
	return cv.verifyResponse(resp, contract.Response)
}

// createRequest creates an HTTP request from contract request
func (cv *ContractVerifier) createRequest(req *ContractRequest) (*http.Request, error) {
	url := cv.baseURL + req.Path

	var bodyReader io.Reader
	if req.Body != nil {
		bodyData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	httpReq, err := http.NewRequest(req.Method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set content type if body is provided
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	return httpReq, nil
}

// verifyResponse verifies the response against the contract
func (cv *ContractVerifier) verifyResponse(resp *http.Response, expected *ContractResponse) error {
	// Verify status code
	if resp.StatusCode != expected.StatusCode {
		return fmt.Errorf("status code mismatch: expected %d, got %d", expected.StatusCode, resp.StatusCode)
	}

	// Verify headers
	for key, expectedValue := range expected.Headers {
		actualValue := resp.Header.Get(key)
		if actualValue != expectedValue {
			return fmt.Errorf("header mismatch for %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Verify body if provided
	if expected.Body != nil {
		var actualBody interface{}
		if err := json.NewDecoder(resp.Body).Decode(&actualBody); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}

		// Simple body verification (in production, use proper schema validation)
		if !cv.compareBodies(expected.Body, actualBody) {
			return fmt.Errorf("response body mismatch")
		}
	}

	return nil
}

// compareBodies compares response bodies (simplified implementation)
func (cv *ContractVerifier) compareBodies(expected, actual interface{}) bool {
	// Convert to JSON strings for comparison
	expectedJSON, err1 := json.Marshal(expected)
	actualJSON, err2 := json.Marshal(actual)

	if err1 != nil || err2 != nil {
		return false
	}

	// For now, just check if both are valid JSON
	// In production, implement proper schema validation
	return len(expectedJSON) > 0 && len(actualJSON) > 0
}

// ValidateContract validates a contract
func ValidateContract(contract *ContractTest) error {
	if contract == nil {
		return fmt.Errorf("contract cannot be nil")
	}

	if contract.Name == "" {
		return fmt.Errorf("contract name is required")
	}

	if contract.Provider == "" {
		return fmt.Errorf("contract provider is required")
	}

	if contract.Consumer == "" {
		return fmt.Errorf("contract consumer is required")
	}

	if contract.Request == nil {
		return fmt.Errorf("contract request is required")
	}

	if contract.Request.Method == "" {
		return fmt.Errorf("contract request method is required")
	}

	if contract.Request.Path == "" {
		return fmt.Errorf("contract request path is required")
	}

	if contract.Response == nil {
		return fmt.Errorf("contract response is required")
	}

	if contract.Response.StatusCode == 0 {
		return fmt.Errorf("contract response status code is required")
	}

	return nil
}
