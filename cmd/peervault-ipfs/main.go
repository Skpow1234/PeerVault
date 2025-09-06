package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Skpow1234/Peervault/internal/content"
	"github.com/Skpow1234/Peervault/internal/ipfs"
)

func main() {
	var (
		command = flag.String("command", "help", "Command to execute (add, get, cat, stat, pin, unpin, list)")
		file    = flag.String("file", "", "File to process")
		cid     = flag.String("cid", "", "CID to process")
		codec   = flag.String("codec", "raw", "Codec to use")
		output  = flag.String("output", "", "Output file")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help || *command == "help" {
		showHelp()
		return
	}

	// Create IPFS compatibility layer
	ipfsCompat := ipfs.NewIPFSCompatibility()
	ctx := context.Background()

	switch *command {
	case "add":
		if *file == "" {
			log.Fatal("File path is required for add command")
		}
		addFile(ctx, ipfsCompat, *file, *codec)
	case "get":
		if *cid == "" {
			log.Fatal("CID is required for get command")
		}
		getFile(ctx, ipfsCompat, *cid, *output)
	case "cat":
		if *cid == "" {
			log.Fatal("CID is required for cat command")
		}
		catFile(ctx, ipfsCompat, *cid)
	case "stat":
		if *cid == "" {
			log.Fatal("CID is required for stat command")
		}
		statFile(ctx, ipfsCompat, *cid)
	case "pin":
		if *cid == "" {
			log.Fatal("CID is required for pin command")
		}
		pinFile(ctx, ipfsCompat, *cid)
	case "unpin":
		if *cid == "" {
			log.Fatal("CID is required for unpin command")
		}
		unpinFile(ctx, ipfsCompat, *cid)
	case "list":
		listFiles(ctx, ipfsCompat)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func addFile(ctx context.Context, ipfsCompat *ipfs.IPFSCompatibility, filePath, codec string) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Add to IPFS
	cid, err := ipfsCompat.AddBlock(ctx, data, codec)
	if err != nil {
		log.Fatalf("Failed to add file: %v", err)
	}

	fmt.Printf("Added file: %s\n", filePath)
	fmt.Printf("CID: %s\n", cid.Hash)
	fmt.Printf("Size: %d bytes\n", len(data))
}

func getFile(ctx context.Context, ipfsCompat *ipfs.IPFSCompatibility, cidStr, output string) {
	// Parse CID
	contentAddresser := content.NewContentAddresser()
	cid, err := contentAddresser.ParseCID(cidStr)
	if err != nil {
		log.Fatalf("Failed to parse CID: %v", err)
	}

	// Get file
	reader, err := ipfsCompat.Cat(ctx, cid)
	if err != nil {
		log.Fatalf("Failed to get file: %v", err)
	}

	// Read all data
	data := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			data = append(data, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Write to output file
	if output == "" {
		output = "output.bin"
	}

	err = os.WriteFile(output, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("Retrieved file: %s\n", output)
	fmt.Printf("Size: %d bytes\n", len(data))
}

func catFile(ctx context.Context, ipfsCompat *ipfs.IPFSCompatibility, cidStr string) {
	// Parse CID
	contentAddresser := content.NewContentAddresser()
	cid, err := contentAddresser.ParseCID(cidStr)
	if err != nil {
		log.Fatalf("Failed to parse CID: %v", err)
	}

	// Get file
	reader, err := ipfsCompat.Cat(ctx, cid)
	if err != nil {
		log.Fatalf("Failed to get file: %v", err)
	}

	// Read and print data
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			fmt.Print(string(buffer[:n]))
		}
		if err != nil {
			break
		}
	}
}

func statFile(ctx context.Context, ipfsCompat *ipfs.IPFSCompatibility, cidStr string) {
	// Parse CID
	contentAddresser := content.NewContentAddresser()
	cid, err := contentAddresser.ParseCID(cidStr)
	if err != nil {
		log.Fatalf("Failed to parse CID: %v", err)
	}

	// Get stats
	stats, err := ipfsCompat.Stat(ctx, cid)
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}

	fmt.Printf("CID: %s\n", cid.Hash)
	fmt.Printf("Type: %s\n", stats["type"])
	fmt.Printf("Size: %d bytes\n", stats["size"])
	if stats["created"] != nil {
		fmt.Printf("Created: %s\n", stats["created"])
	}
	if stats["links"] != nil {
		fmt.Printf("Links: %d\n", stats["links"])
	}
}

func pinFile(ctx context.Context, ipfsCompat *ipfs.IPFSCompatibility, cidStr string) {
	// Parse CID
	contentAddresser := content.NewContentAddresser()
	cid, err := contentAddresser.ParseCID(cidStr)
	if err != nil {
		log.Fatalf("Failed to parse CID: %v", err)
	}

	// Pin file
	err = ipfsCompat.PinObject(ctx, cid, cidStr, "recursive")
	if err != nil {
		log.Fatalf("Failed to pin file: %v", err)
	}

	fmt.Printf("Pinned file: %s\n", cidStr)
}

func unpinFile(ctx context.Context, ipfsCompat *ipfs.IPFSCompatibility, cidStr string) {
	// Parse CID
	contentAddresser := content.NewContentAddresser()
	cid, err := contentAddresser.ParseCID(cidStr)
	if err != nil {
		log.Fatalf("Failed to parse CID: %v", err)
	}

	// Unpin file
	err = ipfsCompat.UnpinObject(ctx, cid)
	if err != nil {
		log.Fatalf("Failed to unpin file: %v", err)
	}

	fmt.Printf("Unpinned file: %s\n", cidStr)
}

func listFiles(ctx context.Context, ipfsCompat *ipfs.IPFSCompatibility) {
	// Get storage stats
	stats, err := ipfsCompat.GetStorageStats(ctx)
	if err != nil {
		log.Fatalf("Failed to get storage stats: %v", err)
	}

	fmt.Printf("Storage Statistics:\n")
	fmt.Printf("Blocks: %d\n", stats["blocks"])
	fmt.Printf("DAG Nodes: %d\n", stats["dag_nodes"])
	fmt.Printf("Pins: %d\n", stats["pins"])
	fmt.Printf("Nodes: %d\n", stats["nodes"])
	fmt.Printf("Total Size: %d bytes\n", stats["total_size"])

	// List pins
	pins, err := ipfsCompat.ListPins(ctx)
	if err != nil {
		log.Fatalf("Failed to list pins: %v", err)
	}

	if len(pins) > 0 {
		fmt.Printf("\nPinned Objects:\n")
		for _, pin := range pins {
			fmt.Printf("  %s (%s) - %s\n", pin.CID.Hash, pin.Type, pin.Name)
		}
	}
}

func showHelp() {
	fmt.Printf("PeerVault IPFS Compatibility Tool\n\n")
	fmt.Printf("Usage: peervault-ipfs -command <command> [options]\n\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  add     Add a file to IPFS storage\n")
	fmt.Printf("  get     Retrieve a file by CID\n")
	fmt.Printf("  cat     Display file content by CID\n")
	fmt.Printf("  stat    Show file statistics by CID\n")
	fmt.Printf("  pin     Pin a file by CID\n")
	fmt.Printf("  unpin   Unpin a file by CID\n")
	fmt.Printf("  list    List storage statistics and pinned objects\n")
	fmt.Printf("  help    Show this help message\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -file <path>     File path (for add command)\n")
	fmt.Printf("  -cid <cid>       CID (for get, cat, stat, pin, unpin commands)\n")
	fmt.Printf("  -codec <codec>   Codec to use (default: raw)\n")
	fmt.Printf("  -output <path>   Output file path (for get command)\n")
	fmt.Printf("  -help            Show this help message\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  peervault-ipfs -command add -file example.txt\n")
	fmt.Printf("  peervault-ipfs -command get -cid QmHash -output retrieved.txt\n")
	fmt.Printf("  peervault-ipfs -command cat -cid QmHash\n")
	fmt.Printf("  peervault-ipfs -command stat -cid QmHash\n")
	fmt.Printf("  peervault-ipfs -command pin -cid QmHash\n")
	fmt.Printf("  peervault-ipfs -command list\n")
}
