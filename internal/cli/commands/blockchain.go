package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/blockchain"
	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// BlockchainCommand handles blockchain operations
type BlockchainCommand struct {
	BaseCommand
	walletManager   *blockchain.WalletManager
	contractManager *blockchain.ContractManager
}

// NewBlockchainCommand creates a new blockchain command
func NewBlockchainCommand(client *client.Client, formatter *formatter.Formatter, walletManager *blockchain.WalletManager, contractManager *blockchain.ContractManager) *BlockchainCommand {
	return &BlockchainCommand{
		BaseCommand: BaseCommand{
			name:        "blockchain",
			description: "Blockchain and smart contract operations",
			usage:       "blockchain [command] [options]",
			client:      client,
			formatter:   formatter,
		},
		walletManager:   walletManager,
		contractManager: contractManager,
	}
}

// Execute executes the blockchain command
func (c *BlockchainCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp()
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "wallet":
		return c.handleWalletCommand(ctx, subArgs)
	case "contract":
		return c.handleContractCommand(ctx, subArgs)
	case "tx", "transaction":
		return c.handleTransactionCommand(ctx, subArgs)
	case "stats":
		return c.handleStatsCommand(ctx, subArgs)
	case "help":
		return c.showHelp()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// Wallet commands
func (c *BlockchainCommand) handleWalletCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: blockchain wallet [create|list|get|balance] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createWallet(ctx, subArgs)
	case "list":
		return c.listWallets(ctx, subArgs)
	case "get":
		return c.getWallet(ctx, subArgs)
	case "balance":
		return c.updateWalletBalance(ctx, subArgs)
	default:
		return fmt.Errorf("unknown wallet subcommand: %s", subcommand)
	}
}

// Contract commands
func (c *BlockchainCommand) handleContractCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: blockchain contract [deploy|list|get|call] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "deploy":
		return c.deployContract(ctx, subArgs)
	case "list":
		return c.listContracts(ctx, subArgs)
	case "get":
		return c.getContract(ctx, subArgs)
	case "call":
		return c.callContract(ctx, subArgs)
	default:
		return fmt.Errorf("unknown contract subcommand: %s", subcommand)
	}
}

// Transaction commands
func (c *BlockchainCommand) handleTransactionCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: blockchain tx [create|list|get|status] [options]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createTransaction(ctx, subArgs)
	case "list":
		return c.listTransactions(ctx, subArgs)
	case "get":
		return c.getTransaction(ctx, subArgs)
	case "status":
		return c.updateTransactionStatus(ctx, subArgs)
	default:
		return fmt.Errorf("unknown transaction subcommand: %s", subcommand)
	}
}

// Stats command
func (c *BlockchainCommand) handleStatsCommand(ctx context.Context, _ []string) error {
	return c.getBlockchainStats(ctx)
}

// Wallet operations
func (c *BlockchainCommand) createWallet(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: blockchain wallet create <name> <currency> <network>")
	}

	name := args[0]
	currency := args[1]
	network := args[2]

	wallet, err := c.walletManager.CreateWallet(ctx, name, currency, network)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Wallet created successfully: %s", wallet.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", wallet.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Address: %s", wallet.Address))
	c.formatter.PrintInfo(fmt.Sprintf("  Currency: %s", wallet.Currency))
	c.formatter.PrintInfo(fmt.Sprintf("  Network: %s", wallet.Network))
	c.formatter.PrintInfo(fmt.Sprintf("  Balance: %s", wallet.Balance))

	return nil
}

func (c *BlockchainCommand) listWallets(ctx context.Context, _ []string) error {
	wallets, err := c.walletManager.ListWallets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list wallets: %v", err)
	}

	if len(wallets) == 0 {
		c.formatter.PrintInfo("No wallets found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d wallets:", len(wallets)))
	for _, wallet := range wallets {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s %s on %s",
			wallet.Name, wallet.ID[:8], wallet.Balance, wallet.Currency, wallet.Network))
		c.formatter.PrintInfo(fmt.Sprintf("    Address: %s", wallet.Address))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", wallet.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *BlockchainCommand) getWallet(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: blockchain wallet get <wallet_id>")
	}

	walletID := args[0]
	wallet, err := c.walletManager.GetWallet(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Wallet Details: %s", wallet.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  ID: %s", wallet.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Address: %s", wallet.Address))
	c.formatter.PrintInfo(fmt.Sprintf("  Public Key: %s", wallet.PublicKey))
	c.formatter.PrintInfo(fmt.Sprintf("  Balance: %s %s", wallet.Balance, wallet.Currency))
	c.formatter.PrintInfo(fmt.Sprintf("  Network: %s", wallet.Network))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", wallet.CreatedAt.Format(time.RFC3339)))
	c.formatter.PrintInfo(fmt.Sprintf("  Last Used: %s", wallet.LastUsed.Format(time.RFC3339)))
	c.formatter.PrintInfo(fmt.Sprintf("  Active: %v", wallet.IsActive))

	return nil
}

func (c *BlockchainCommand) updateWalletBalance(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: blockchain wallet balance <wallet_id> <new_balance>")
	}

	walletID := args[0]
	balance := args[1]

	err := c.walletManager.UpdateWalletBalance(ctx, walletID, balance)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Wallet balance updated successfully: %s %s", balance, "ETH"))
	return nil
}

// Contract operations
func (c *BlockchainCommand) deployContract(ctx context.Context, args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: blockchain contract deploy <name> <abi> <bytecode> <network> <deployed_by>")
	}

	name := args[0]
	abi := args[1]
	bytecode := args[2]
	network := args[3]
	deployedBy := args[4]

	contract, err := c.contractManager.DeployContract(ctx, name, abi, bytecode, network, deployedBy)
	if err != nil {
		return fmt.Errorf("failed to deploy contract: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Contract deployed successfully: %s", contract.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Name: %s", contract.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  Address: %s", contract.Address))
	c.formatter.PrintInfo(fmt.Sprintf("  Network: %s", contract.Network))
	c.formatter.PrintInfo(fmt.Sprintf("  Deployed By: %s", contract.DeployedBy))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", contract.CreatedAt.Format(time.RFC3339)))

	return nil
}

func (c *BlockchainCommand) listContracts(ctx context.Context, _ []string) error {
	contracts, err := c.contractManager.ListContracts(ctx)
	if err != nil {
		return fmt.Errorf("failed to list contracts: %v", err)
	}

	if len(contracts) == 0 {
		c.formatter.PrintInfo("No contracts found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d contracts:", len(contracts)))
	for _, contract := range contracts {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s",
			contract.Name, contract.ID[:8], contract.Address))
		c.formatter.PrintInfo(fmt.Sprintf("    Network: %s", contract.Network))
		c.formatter.PrintInfo(fmt.Sprintf("    Deployed By: %s", contract.DeployedBy))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", contract.CreatedAt.Format(time.RFC3339)))
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *BlockchainCommand) getContract(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: blockchain contract get <contract_id>")
	}

	contractID := args[0]
	contract, err := c.contractManager.GetContract(ctx, contractID)
	if err != nil {
		return fmt.Errorf("failed to get contract: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Contract Details: %s", contract.Name))
	c.formatter.PrintInfo(fmt.Sprintf("  ID: %s", contract.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Address: %s", contract.Address))
	c.formatter.PrintInfo(fmt.Sprintf("  Network: %s", contract.Network))
	c.formatter.PrintInfo(fmt.Sprintf("  Deployed By: %s", contract.DeployedBy))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", contract.CreatedAt.Format(time.RFC3339)))
	c.formatter.PrintInfo(fmt.Sprintf("  Active: %v", contract.IsActive))
	c.formatter.PrintInfo(fmt.Sprintf("  ABI Length: %d characters", len(contract.ABI)))
	c.formatter.PrintInfo(fmt.Sprintf("  Bytecode Length: %d characters", len(contract.Bytecode)))

	return nil
}

func (c *BlockchainCommand) callContract(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: blockchain contract call <contract_id> <function_name> <parameters> <value> <gas_limit>")
	}

	contractID := args[0]
	functionName := args[1]
	parameters := strings.Split(args[2], ",")
	value := args[3]
	gasLimit := args[4]

	call, err := c.contractManager.CallContract(ctx, contractID, functionName, parameters, value, gasLimit)
	if err != nil {
		return fmt.Errorf("failed to call contract: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Contract call initiated: %s", call.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  Contract: %s", call.ContractID))
	c.formatter.PrintInfo(fmt.Sprintf("  Function: %s", call.FunctionName))
	c.formatter.PrintInfo(fmt.Sprintf("  Parameters: %v", call.Parameters))
	c.formatter.PrintInfo(fmt.Sprintf("  Value: %s", call.Value))
	c.formatter.PrintInfo(fmt.Sprintf("  Gas Limit: %s", call.GasLimit))
	c.formatter.PrintInfo(fmt.Sprintf("  Transaction Hash: %s", call.TxHash))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", call.Status))

	return nil
}

// Transaction operations
func (c *BlockchainCommand) createTransaction(ctx context.Context, args []string) error {
	if len(args) < 6 {
		return fmt.Errorf("usage: blockchain tx create <from> <to> <amount> <currency> <gas_price> <gas_limit>")
	}

	from := args[0]
	to := args[1]
	amount := args[2]
	currency := args[3]
	gasPrice := args[4]
	gasLimit := args[5]

	transaction, err := c.walletManager.CreateTransaction(ctx, from, to, amount, currency, gasPrice, gasLimit)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Transaction created successfully: %s", transaction.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  From: %s", transaction.From))
	c.formatter.PrintInfo(fmt.Sprintf("  To: %s", transaction.To))
	c.formatter.PrintInfo(fmt.Sprintf("  Amount: %s %s", transaction.Amount, transaction.Currency))
	c.formatter.PrintInfo(fmt.Sprintf("  Gas Price: %s", transaction.GasPrice))
	c.formatter.PrintInfo(fmt.Sprintf("  Gas Limit: %s", transaction.GasLimit))
	c.formatter.PrintInfo(fmt.Sprintf("  Hash: %s", transaction.Hash))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", transaction.Status))

	return nil
}

func (c *BlockchainCommand) listTransactions(ctx context.Context, _ []string) error {
	transactions, err := c.walletManager.ListTransactions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list transactions: %v", err)
	}

	if len(transactions) == 0 {
		c.formatter.PrintInfo("No transactions found")
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Found %d transactions:", len(transactions)))
	for _, tx := range transactions {
		c.formatter.PrintInfo(fmt.Sprintf("  %s (%s) - %s %s",
			tx.ID[:8], tx.Status, tx.Amount, tx.Currency))
		c.formatter.PrintInfo(fmt.Sprintf("    From: %s", tx.From))
		c.formatter.PrintInfo(fmt.Sprintf("    To: %s", tx.To))
		c.formatter.PrintInfo(fmt.Sprintf("    Hash: %s", tx.Hash))
		c.formatter.PrintInfo(fmt.Sprintf("    Created: %s", tx.CreatedAt.Format(time.RFC3339)))
		if tx.ConfirmedAt != nil {
			c.formatter.PrintInfo(fmt.Sprintf("    Confirmed: %s", tx.ConfirmedAt.Format(time.RFC3339)))
		}
		c.formatter.PrintInfo("")
	}

	return nil
}

func (c *BlockchainCommand) getTransaction(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: blockchain tx get <transaction_id>")
	}

	txID := args[0]
	transaction, err := c.walletManager.GetTransaction(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %v", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Transaction Details: %s", transaction.ID))
	c.formatter.PrintInfo(fmt.Sprintf("  From: %s", transaction.From))
	c.formatter.PrintInfo(fmt.Sprintf("  To: %s", transaction.To))
	c.formatter.PrintInfo(fmt.Sprintf("  Amount: %s %s", transaction.Amount, transaction.Currency))
	c.formatter.PrintInfo(fmt.Sprintf("  Gas Price: %s", transaction.GasPrice))
	c.formatter.PrintInfo(fmt.Sprintf("  Gas Limit: %s", transaction.GasLimit))
	c.formatter.PrintInfo(fmt.Sprintf("  Nonce: %d", transaction.Nonce))
	c.formatter.PrintInfo(fmt.Sprintf("  Hash: %s", transaction.Hash))
	c.formatter.PrintInfo(fmt.Sprintf("  Status: %s", transaction.Status))
	c.formatter.PrintInfo(fmt.Sprintf("  Block Number: %d", transaction.BlockNumber))
	c.formatter.PrintInfo(fmt.Sprintf("  Created: %s", transaction.CreatedAt.Format(time.RFC3339)))
	if transaction.ConfirmedAt != nil {
		c.formatter.PrintInfo(fmt.Sprintf("  Confirmed: %s", transaction.ConfirmedAt.Format(time.RFC3339)))
	}

	return nil
}

func (c *BlockchainCommand) updateTransactionStatus(ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: blockchain tx status <transaction_id> <status> <block_number>")
	}

	txID := args[0]
	status := args[1]
	blockNumber, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid block number: %v", err)
	}

	err = c.walletManager.UpdateTransactionStatus(ctx, txID, status, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %v", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Transaction status updated successfully: %s -> %s", txID, status))
	return nil
}

// Stats operation
func (c *BlockchainCommand) getBlockchainStats(ctx context.Context) error {
	walletStats, err := c.walletManager.GetBlockchainStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get wallet stats: %v", err)
	}

	contractStats, err := c.contractManager.GetContractStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get contract stats: %v", err)
	}

	c.formatter.PrintInfo("Blockchain Statistics:")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Wallets:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Wallets: %d", walletStats.TotalWallets))
	c.formatter.PrintInfo(fmt.Sprintf("  Active Wallets: %d", walletStats.ActiveWallets))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Volume: %s", walletStats.TotalVolume))
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Transactions:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Transactions: %d", walletStats.TotalTransactions))
	c.formatter.PrintInfo(fmt.Sprintf("  Pending Transactions: %d", walletStats.PendingTxs))
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Smart Contracts:")
	c.formatter.PrintInfo(fmt.Sprintf("  Total Contracts: %d", contractStats.TotalContracts))
	c.formatter.PrintInfo(fmt.Sprintf("  Active Contracts: %d", contractStats.ActiveContracts))
	c.formatter.PrintInfo(fmt.Sprintf("  Total Calls: %d", contractStats.TotalCalls))
	c.formatter.PrintInfo(fmt.Sprintf("  Successful Calls: %d", contractStats.SuccessfulCalls))
	c.formatter.PrintInfo(fmt.Sprintf("  Failed Calls: %d", contractStats.FailedCalls))
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo(fmt.Sprintf("Last Updated: %s", walletStats.LastUpdated.Format(time.RFC3339)))

	return nil
}

// Help
func (c *BlockchainCommand) showHelp() error {
	c.formatter.PrintInfo("Blockchain Command Help:")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Usage: blockchain [command] [options]")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Commands:")
	c.formatter.PrintInfo("  wallet [create|list|get|balance]  - Wallet management")
	c.formatter.PrintInfo("  contract [deploy|list|get|call]   - Smart contract operations")
	c.formatter.PrintInfo("  tx [create|list|get|status]       - Transaction management")
	c.formatter.PrintInfo("  stats                              - Show blockchain statistics")
	c.formatter.PrintInfo("  help                               - Show this help")
	c.formatter.PrintInfo("")
	c.formatter.PrintInfo("Examples:")
	c.formatter.PrintInfo("  blockchain wallet create MyWallet ETH mainnet")
	c.formatter.PrintInfo("  blockchain wallet list")
	c.formatter.PrintInfo("  blockchain contract deploy MyContract '[...]' '0x...' mainnet user123")
	c.formatter.PrintInfo("  blockchain tx create 0x123... 0x456... 1.5 ETH 20000000000 21000")
	c.formatter.PrintInfo("  blockchain stats")

	return nil
}
