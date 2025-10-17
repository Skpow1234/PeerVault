package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/google/uuid"
)

// ContractManager manages smart contracts
type ContractManager struct {
	mu        sync.RWMutex
	client    *client.Client
	configDir string
	contracts map[string]*SmartContract
	calls     map[string]*ContractCall
	stats     *ContractStats
}

// ContractStats represents smart contract statistics
type ContractStats struct {
	TotalContracts  int       `json:"total_contracts"`
	TotalCalls      int       `json:"total_calls"`
	ActiveContracts int       `json:"active_contracts"`
	SuccessfulCalls int       `json:"successful_calls"`
	FailedCalls     int       `json:"failed_calls"`
	TotalGasUsed    string    `json:"total_gas_used"`
	LastUpdated     time.Time `json:"last_updated"`
}

// NewContractManager creates a new contract manager
func NewContractManager(client *client.Client, configDir string) *ContractManager {
	cm := &ContractManager{
		client:    client,
		configDir: configDir,
		contracts: make(map[string]*SmartContract),
		calls:     make(map[string]*ContractCall),
		stats:     &ContractStats{},
	}
	_ = cm.loadContracts() // Ignore error for initialization
	_ = cm.loadCalls()     // Ignore error for initialization
	_ = cm.loadStats()     // Ignore error for initialization
	return cm
}

// DeployContract deploys a new smart contract
func (cm *ContractManager) DeployContract(ctx context.Context, name, abi, bytecode, network, deployedBy string) (*SmartContract, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	contract := &SmartContract{
		ID:         uuid.New().String(),
		Name:       name,
		Address:    fmt.Sprintf("0x%x", uuid.New()),
		ABI:        abi,
		Bytecode:   bytecode,
		Network:    network,
		DeployedBy: deployedBy,
		CreatedAt:  time.Now(),
		IsActive:   true,
		Metadata:   make(map[string]interface{}),
	}

	cm.contracts[contract.ID] = contract

	// Simulate API call - store contract data as JSON
	contractData, err := json.Marshal(contract)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contract: %v", err)
	}

	tempFilePath := filepath.Join(cm.configDir, fmt.Sprintf("contracts/%s.json", contract.ID))
	if err := os.WriteFile(tempFilePath, contractData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write contract data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = cm.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store contract: %v", err)
	}

	cm.stats.TotalContracts++
	cm.stats.ActiveContracts++
	_ = cm.saveStats()
	_ = cm.saveContracts()
	return contract, nil
}

// ListContracts returns all contracts
func (cm *ContractManager) ListContracts(ctx context.Context) ([]*SmartContract, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	contracts := make([]*SmartContract, 0, len(cm.contracts))
	for _, contract := range cm.contracts {
		if contract.IsActive {
			contracts = append(contracts, contract)
		}
	}
	return contracts, nil
}

// GetContract returns a contract by ID
func (cm *ContractManager) GetContract(ctx context.Context, contractID string) (*SmartContract, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	contract, exists := cm.contracts[contractID]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractID)
	}
	return contract, nil
}

// GetContractByAddress returns a contract by address
func (cm *ContractManager) GetContractByAddress(ctx context.Context, address string) (*SmartContract, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, contract := range cm.contracts {
		if contract.Address == address && contract.IsActive {
			return contract, nil
		}
	}
	return nil, fmt.Errorf("contract not found for address: %s", address)
}

// CallContract calls a smart contract function
func (cm *ContractManager) CallContract(ctx context.Context, contractID, functionName string, parameters []string, value, gasLimit string) (*ContractCall, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	_, exists := cm.contracts[contractID]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractID)
	}

	call := &ContractCall{
		ID:           uuid.New().String(),
		ContractID:   contractID,
		FunctionName: functionName,
		Parameters:   parameters,
		Value:        value,
		GasLimit:     gasLimit,
		TxHash:       fmt.Sprintf("0x%x", uuid.New()),
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	cm.calls[call.ID] = call

	// Simulate API call - store call data as JSON
	callData, err := json.Marshal(call)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contract call: %v", err)
	}

	tempFilePath := filepath.Join(cm.configDir, fmt.Sprintf("contract_calls/%s.json", call.ID))
	if err := os.WriteFile(tempFilePath, callData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write call data to temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFilePath) }() // Clean up temp file

	_, err = cm.client.StoreFile(ctx, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to store contract call: %v", err)
	}

	cm.stats.TotalCalls++
	_ = cm.saveStats()
	_ = cm.saveCalls()
	return call, nil
}

// ListContractCalls returns all contract calls
func (cm *ContractManager) ListContractCalls(ctx context.Context) ([]*ContractCall, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	calls := make([]*ContractCall, 0, len(cm.calls))
	for _, call := range cm.calls {
		calls = append(calls, call)
	}
	return calls, nil
}

// GetContractCall returns a contract call by ID
func (cm *ContractManager) GetContractCall(ctx context.Context, callID string) (*ContractCall, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	call, exists := cm.calls[callID]
	if !exists {
		return nil, fmt.Errorf("contract call not found: %s", callID)
	}
	return call, nil
}

// UpdateContractCallStatus updates contract call status
func (cm *ContractManager) UpdateContractCallStatus(ctx context.Context, callID, status, result string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	call, exists := cm.calls[callID]
	if !exists {
		return fmt.Errorf("contract call not found: %s", callID)
	}

	call.Status = status
	call.Result = result
	switch status {
	case "completed":
		now := time.Now()
		call.ExecutedAt = &now
		cm.stats.SuccessfulCalls++
	case "failed":
		cm.stats.FailedCalls++
	}

	_ = cm.saveCalls()
	_ = cm.saveStats()
	return nil
}

// GetContractStats returns contract statistics
func (cm *ContractManager) GetContractStats(ctx context.Context) (*ContractStats, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Update stats
	cm.stats.LastUpdated = time.Now()
	return cm.stats, nil
}

// File operations
func (cm *ContractManager) loadContracts() error {
	contractsFile := filepath.Join(cm.configDir, "contracts.json")
	if _, err := os.Stat(contractsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(contractsFile)
	if err != nil {
		return fmt.Errorf("failed to read contracts file: %w", err)
	}

	var contracts []*SmartContract
	if err := json.Unmarshal(data, &contracts); err != nil {
		return fmt.Errorf("failed to unmarshal contracts: %w", err)
	}

	for _, contract := range contracts {
		cm.contracts[contract.ID] = contract
	}
	return nil
}

func (cm *ContractManager) saveContracts() error {
	contractsFile := filepath.Join(cm.configDir, "contracts.json")

	var contracts []*SmartContract
	for _, contract := range cm.contracts {
		contracts = append(contracts, contract)
	}

	data, err := json.MarshalIndent(contracts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal contracts: %w", err)
	}

	return os.WriteFile(contractsFile, data, 0644)
}

func (cm *ContractManager) loadCalls() error {
	callsFile := filepath.Join(cm.configDir, "contract_calls.json")
	if _, err := os.Stat(callsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(callsFile)
	if err != nil {
		return fmt.Errorf("failed to read calls file: %w", err)
	}

	var calls []*ContractCall
	if err := json.Unmarshal(data, &calls); err != nil {
		return fmt.Errorf("failed to unmarshal calls: %w", err)
	}

	for _, call := range calls {
		cm.calls[call.ID] = call
	}
	return nil
}

func (cm *ContractManager) saveCalls() error {
	callsFile := filepath.Join(cm.configDir, "contract_calls.json")

	var calls []*ContractCall
	for _, call := range cm.calls {
		calls = append(calls, call)
	}

	data, err := json.MarshalIndent(calls, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal calls: %w", err)
	}

	return os.WriteFile(callsFile, data, 0644)
}

func (cm *ContractManager) loadStats() error {
	statsFile := filepath.Join(cm.configDir, "contract_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats ContractStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	cm.stats = &stats
	return nil
}

func (cm *ContractManager) saveStats() error {
	statsFile := filepath.Join(cm.configDir, "contract_stats.json")

	data, err := json.MarshalIndent(cm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
