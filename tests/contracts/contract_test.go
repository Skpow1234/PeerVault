package contracts

import (
	"testing"

	"github.com/Skpow1234/Peervault/internal/api/contracts"
)

func TestContractVerification(t *testing.T) {
	// This is a placeholder test that would normally test against a running service
	// For now, we'll skip if no service is available
	t.Skip("Skipping contract verification - requires running service")
}

func TestContractValidation(t *testing.T) {
	// Test contract validation logic
	validContract := &contracts.ContractTest{
		Name:        "test-contract",
		Description: "A test contract",
		Provider:    "test-provider",
		Consumer:    "test-consumer",
		Request: &contracts.ContractRequest{
			Method: "GET",
			Path:   "/test",
		},
		Response: &contracts.ContractResponse{
			StatusCode: 200,
		},
	}

	if err := contracts.ValidateContract(validContract); err != nil {
		t.Errorf("Valid contract should pass validation: %v", err)
	}

	// Test invalid contract
	invalidContract := &contracts.ContractTest{
		Name: "", // Missing name
	}

	if err := contracts.ValidateContract(invalidContract); err == nil {
		t.Error("Invalid contract should fail validation")
	}
}
