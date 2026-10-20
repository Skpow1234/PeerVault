package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/Skpow1234/Peervault/internal/blockchain"
)

func main() {
	var (
		command = flag.String("command", "help", "Command to execute (network, deploy, identity, transaction, help)")
		network = flag.String("network", "ethereum", "Blockchain network name")
		rpc     = flag.String("rpc", "http://localhost:8545", "RPC URL")
		chainID = flag.Int64("chain-id", 1, "Chain ID")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help || *command == "help" {
		showHelp()
		return
	}

	// Create blockchain integration
	blockchainIntegration := blockchain.NewBlockchainIntegration()
	ctx := context.Background()

	// Add default network
	defaultNetwork := &blockchain.BlockchainNetwork{
		Name:     *network,
		ChainID:  *chainID,
		RPCURL:   *rpc,
		WSURL:    "",
		Explorer: "",
	}

	err := blockchainIntegration.AddNetwork(ctx, defaultNetwork)
	if err != nil {
		log.Fatalf("Failed to add network: %v", err)
	}

	switch *command {
	case "network":
		handleNetworkCommand(ctx, blockchainIntegration)
	case "deploy":
		handleDeployCommand(ctx, blockchainIntegration, *network)
	case "identity":
		handleIdentityCommand(ctx, blockchainIntegration, *network)
	case "transaction":
		handleTransactionCommand(ctx, blockchainIntegration, *network)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func handleNetworkCommand(ctx context.Context, bi *blockchain.BlockchainIntegration) {
	// List networks
	networks, err := bi.ListNetworks(ctx)
	if err != nil {
		log.Fatalf("Failed to list networks: %v", err)
	}

	fmt.Printf("Blockchain Networks:\n")
	for _, network := range networks {
		fmt.Printf("  Name: %s\n", network.Name)
		fmt.Printf("  Chain ID: %d\n", network.ChainID)
		fmt.Printf("  RPC URL: %s\n", network.RPCURL)
		fmt.Printf("  Explorer: %s\n", network.Explorer)
		fmt.Printf("  ---\n")
	}

	// Get network stats
	for _, network := range networks {
		stats, err := bi.GetNetworkStats(ctx, network.Name)
		if err != nil {
			log.Printf("Failed to get stats for network %s: %v", network.Name, err)
			continue
		}

		fmt.Printf("Network Stats for %s:\n", network.Name)
		fmt.Printf("  Contracts: %d\n", stats["contracts"])
		fmt.Printf("  Identities: %d\n", stats["identities"])
		fmt.Printf("  ---\n")
	}
}

func handleDeployCommand(ctx context.Context, bi *blockchain.BlockchainIntegration, networkName string) {
	// Create a sample smart contract
	contract := &blockchain.SmartContract{
		Address:  "0x1234567890123456789012345678901234567890",
		ABI:      `[{"inputs":[],"name":"getValue","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`,
		Bytecode: "0x608060405234801561001057600080fd5b50600436106100365760003560e01c8063209652551461003b5780635524107714610059575b600080fd5b610043610075565b60405161005091906100a1565b60405180910390f35b61006161007b565b60405161006e91906100a1565b60405180910390f35b60005481565b60008054905090565b6000819050919050565b61009b81610088565b82525050565b60006020820190506100b66000830184610092565b9291505056fea2646970667358221220...",
		Name:     "SampleContract",
		Version:  "1.0.0",
		Metadata: map[string]interface{}{
			"description": "A sample smart contract",
			"author":      "PeerVault",
		},
	}

	// Deploy contract
	tx, err := bi.DeployContract(ctx, contract, networkName)
	if err != nil {
		log.Fatalf("Failed to deploy contract: %v", err)
	}

	fmt.Printf("Contract Deployed Successfully!\n")
	fmt.Printf("Contract Address: %s\n", contract.Address)
	fmt.Printf("Transaction Hash: %s\n", tx.Hash)
	fmt.Printf("Block Number: %s\n", tx.BlockNumber.String())
	fmt.Printf("Gas Used: %d\n", tx.GasLimit)
	fmt.Printf("Gas Price: %s\n", tx.GasPrice.String())

	// List contracts
	contracts, err := bi.ListContracts(ctx)
	if err != nil {
		log.Fatalf("Failed to list contracts: %v", err)
	}

	fmt.Printf("\nDeployed Contracts:\n")
	for _, contract := range contracts {
		fmt.Printf("  Address: %s\n", contract.Address)
		fmt.Printf("  Name: %s\n", contract.Name)
		fmt.Printf("  Version: %s\n", contract.Version)
		fmt.Printf("  Deployed At: %s\n", contract.DeployedAt.Format(time.RFC3339))
		fmt.Printf("  ---\n")
	}
}

func handleIdentityCommand(ctx context.Context, bi *blockchain.BlockchainIntegration, networkName string) {
	// Create a new identity
	identity, err := bi.CreateIdentity(ctx, networkName)
	if err != nil {
		log.Fatalf("Failed to create identity: %v", err)
	}

	fmt.Printf("Identity Created Successfully!\n")
	fmt.Printf("DID: %s\n", identity.DID)
	fmt.Printf("Address: %s\n", identity.Address)
	fmt.Printf("Public Key: %s\n", identity.PublicKey)
	fmt.Printf("Network: %s\n", identity.Network.Name)
	fmt.Printf("Created At: %s\n", identity.CreatedAt.Format(time.RFC3339))

	// List identities
	identities, err := bi.ListIdentities(ctx)
	if err != nil {
		log.Fatalf("Failed to list identities: %v", err)
	}

	fmt.Printf("\nCreated Identities:\n")
	for _, identity := range identities {
		fmt.Printf("  DID: %s\n", identity.DID)
		fmt.Printf("  Address: %s\n", identity.Address)
		fmt.Printf("  Network: %s\n", identity.Network.Name)
		fmt.Printf("  Created At: %s\n", identity.CreatedAt.Format(time.RFC3339))
		fmt.Printf("  ---\n")
	}
}

func handleTransactionCommand(ctx context.Context, bi *blockchain.BlockchainIntegration, networkName string) {
	// Create a sample transaction
	tx := &blockchain.Transaction{
		Hash:     "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		From:     "0x1111111111111111111111111111111111111111",
		To:       "0x2222222222222222222222222222222222222222",
		Value:    big.NewInt(1000000000000000000), // 1 ETH in wei
		GasLimit: 21000,
		GasPrice: big.NewInt(20000000000), // 20 gwei
		Nonce:    1,
		Data:     []byte{},
		Status:   "pending",
		Metadata: map[string]interface{}{
			"description": "Sample transaction",
			"type":        "transfer",
		},
	}

	// Send transaction
	err := bi.SendTransaction(ctx, tx, networkName)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction Sent Successfully!\n")
	fmt.Printf("Hash: %s\n", tx.Hash)
	fmt.Printf("From: %s\n", tx.From)
	fmt.Printf("To: %s\n", tx.To)
	fmt.Printf("Value: %s ETH\n", new(big.Float).SetInt(tx.Value).Quo(new(big.Float).SetInt(tx.Value), big.NewFloat(1e18)).String())
	fmt.Printf("Gas Limit: %d\n", tx.GasLimit)
	fmt.Printf("Gas Price: %s gwei\n", new(big.Float).SetInt(tx.GasPrice).Quo(new(big.Float).SetInt(tx.GasPrice), big.NewFloat(1e9)).String())
	fmt.Printf("Status: %s\n", tx.Status)

	// Get transaction details
	retrievedTx, err := bi.GetTransaction(ctx, tx.Hash, networkName)
	if err != nil {
		log.Fatalf("Failed to get transaction: %v", err)
	}

	fmt.Printf("\nTransaction Details:\n")
	fmt.Printf("Hash: %s\n", retrievedTx.Hash)
	fmt.Printf("Block Number: %s\n", retrievedTx.BlockNumber.String())
	fmt.Printf("Block Hash: %s\n", retrievedTx.BlockHash)
	fmt.Printf("Status: %s\n", retrievedTx.Status)
	fmt.Printf("Created At: %s\n", retrievedTx.CreatedAt.Format(time.RFC3339))
}

func showHelp() {
	fmt.Printf("PeerVault Blockchain Integration Tool\n\n")
	fmt.Printf("Usage: peervault-chain -command <command> [options]\n\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  network     List blockchain networks and statistics\n")
	fmt.Printf("  deploy      Deploy a smart contract\n")
	fmt.Printf("  identity    Create and manage decentralized identities\n")
	fmt.Printf("  transaction Send and manage blockchain transactions\n")
	fmt.Printf("  help        Show this help message\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -network <name>    Blockchain network name (default: ethereum)\n")
	fmt.Printf("  -rpc <url>         RPC URL (default: http://localhost:8545)\n")
	fmt.Printf("  -chain-id <id>     Chain ID (default: 1)\n")
	fmt.Printf("  -help              Show this help message\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  peervault-chain -command network\n")
	fmt.Printf("  peervault-chain -command deploy -network ethereum\n")
	fmt.Printf("  peervault-chain -command identity -network ethereum\n")
	fmt.Printf("  peervault-chain -command transaction -network ethereum\n")
}
