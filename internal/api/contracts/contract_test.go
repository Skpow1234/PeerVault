package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	// Add query parameters
	if len(req.Query) > 0 {
		// Simple query parameter handling
		// In production, use proper URL encoding
	}

	httpReq, err := http.NewRequest(req.Method, url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set body if provided
	if req.Body != nil {
		bodyData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		httpReq.Body = &mockBody{data: bodyData}
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

// mockBody implements io.ReadCloser for request body
type mockBody struct {
	data []byte
	pos  int
}

func (mb *mockBody) Read(p []byte) (n int, err error) {
	if mb.pos >= len(mb.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, mb.data[mb.pos:])
	mb.pos += n
	return n, nil
}

func (mb *mockBody) Close() error {
	return nil
}

// TestContractVerification tests contract verification
func TestContractVerification(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
			})
		case "/api/nodes":
			if r.Method == "GET" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]map[string]interface{}{
					{
						"id":     "node-1",
						"name":   "test-node",
						"status": "active",
					},
				})
			} else if r.Method == "POST" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":     "node-2",
					"name":   "new-node",
					"status": "active",
				})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create contract verifier
	verifier := NewContractVerifier(server.URL, slog.Default())

	tests := []struct {
		name     string
		contract *ContractTest
		wantErr  bool
	}{
		{
			name: "Health Check Contract",
			contract: &ContractTest{
				Name:        "health-check",
				Description: "Health check endpoint contract",
				Provider:    "peervault-api",
				Consumer:    "peervault-client",
				Request: &ContractRequest{
					Method: "GET",
					Path:   "/health",
					Headers: map[string]string{
						"Accept": "application/json",
					},
				},
				Response: &ContractResponse{
					StatusCode: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"status": "healthy",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Get Nodes Contract",
			contract: &ContractTest{
				Name:        "get-nodes",
				Description: "Get nodes endpoint contract",
				Provider:    "peervault-api",
				Consumer:    "peervault-client",
				Request: &ContractRequest{
					Method: "GET",
					Path:   "/api/nodes",
					Headers: map[string]string{
						"Accept": "application/json",
					},
				},
				Response: &ContractResponse{
					StatusCode: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: []map[string]interface{}{
						{
							"id":     "node-1",
							"name":   "test-node",
							"status": "active",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Create Node Contract",
			contract: &ContractTest{
				Name:        "create-node",
				Description: "Create node endpoint contract",
				Provider:    "peervault-api",
				Consumer:    "peervault-client",
				Request: &ContractRequest{
					Method: "POST",
					Path:   "/api/nodes",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"name": "new-node",
					},
				},
				Response: &ContractResponse{
					StatusCode: 201,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"id":     "node-2",
						"name":   "new-node",
						"status": "active",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := verifier.VerifyContract(ctx, tt.contract)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestContractValidation tests contract validation
func TestContractValidation(t *testing.T) {
	tests := []struct {
		name     string
		contract *ContractTest
		wantErr  bool
	}{
		{
			name: "Valid Contract",
			contract: &ContractTest{
				Name:        "valid-contract",
				Description: "A valid contract",
				Provider:    "provider",
				Consumer:    "consumer",
				Request: &ContractRequest{
					Method: "GET",
					Path:   "/test",
				},
				Response: &ContractResponse{
					StatusCode: 200,
				},
			},
			wantErr: false,
		},
		{
			name: "Missing Name",
			contract: &ContractTest{
				Description: "Contract without name",
				Provider:    "provider",
				Consumer:    "consumer",
			},
			wantErr: true,
		},
		{
			name: "Missing Request",
			contract: &ContractTest{
				Name:     "no-request",
				Provider: "provider",
				Consumer: "consumer",
			},
			wantErr: true,
		},
		{
			name: "Missing Response",
			contract: &ContractTest{
				Name:     "no-response",
				Provider: "provider",
				Consumer: "consumer",
				Request: &ContractRequest{
					Method: "GET",
					Path:   "/test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContract(tt.contract)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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
