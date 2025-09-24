package contracts

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestContractVerification tests contract verification
func TestContractVerification(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
			}); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		case "/api/nodes":
			switch r.Method {
			case "GET":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode([]map[string]interface{}{
					{
						"id":     "node-1",
						"name":   "test-node",
						"status": "active",
					},
				}); err != nil {
					http.Error(w, "Failed to encode response", http.StatusInternalServerError)
					return
				}
			case "POST":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				if err := json.NewEncoder(w).Encode(map[string]interface{}{
					"id":     "node-2",
					"name":   "new-node",
					"status": "active",
				}); err != nil {
					http.Error(w, "Failed to encode response", http.StatusInternalServerError)
					return
				}
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
