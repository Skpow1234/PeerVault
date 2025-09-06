package ipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/content"
)

// IPFSNode represents an IPFS-compatible node
type IPFSNode struct {
	ID       string                 `json:"id"`
	Address  string                 `json:"address"`
	Protocol string                 `json:"protocol"`
	Version  string                 `json:"version"`
	Metadata map[string]interface{} `json:"metadata"`
}

// IPFSBlock represents an IPFS block
type IPFSBlock struct {
	CID     *content.CID `json:"cid"`
	Data    []byte       `json:"data"`
	Size    int64        `json:"size"`
	Created time.Time    `json:"created"`
}

// IPFSDAGNode represents a DAG node in IPFS
type IPFSDAGNode struct {
	CID     *content.CID   `json:"cid"`
	Links   []*IPFSDAGLink `json:"links"`
	Data    []byte         `json:"data"`
	Size    int64          `json:"size"`
	Created time.Time      `json:"created"`
}

// IPFSDAGLink represents a link in a DAG node
type IPFSDAGLink struct {
	Name string       `json:"name"`
	Size int64        `json:"size"`
	CID  *content.CID `json:"cid"`
}

// IPFSPin represents a pinned object in IPFS
type IPFSPin struct {
	CID     *content.CID `json:"cid"`
	Name    string       `json:"name"`
	Type    string       `json:"type"`
	Created time.Time    `json:"created"`
}

// IPFSCompatibility provides IPFS compatibility features
type IPFSCompatibility struct {
	contentAddresser *content.ContentAddresser
	nodes            map[string]*IPFSNode
	blocks           map[string]*IPFSBlock
	dagNodes         map[string]*IPFSDAGNode
	pins             map[string]*IPFSPin
}

// NewIPFSCompatibility creates a new IPFS compatibility layer
func NewIPFSCompatibility() *IPFSCompatibility {
	return &IPFSCompatibility{
		contentAddresser: content.NewContentAddresser(),
		nodes:            make(map[string]*IPFSNode),
		blocks:           make(map[string]*IPFSBlock),
		dagNodes:         make(map[string]*IPFSDAGNode),
		pins:             make(map[string]*IPFSPin),
	}
}

// AddBlock adds a block to the IPFS-compatible storage
func (ic *IPFSCompatibility) AddBlock(ctx context.Context, data []byte, codec string) (*content.CID, error) {
	// Generate CID for the data
	cid, err := ic.contentAddresser.GenerateCID(data, codec)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CID: %w", err)
	}

	// Create IPFS block
	block := &IPFSBlock{
		CID:     cid,
		Data:    data,
		Size:    int64(len(data)),
		Created: time.Now(),
	}

	// Store block
	ic.blocks[cid.Hash] = block

	return cid, nil
}

// GetBlock retrieves a block by CID
func (ic *IPFSCompatibility) GetBlock(ctx context.Context, cid *content.CID) (*IPFSBlock, error) {
	block, exists := ic.blocks[cid.Hash]
	if !exists {
		return nil, fmt.Errorf("block not found: %s", cid.Hash)
	}

	return block, nil
}

// AddDAGNode adds a DAG node to the IPFS-compatible storage
func (ic *IPFSCompatibility) AddDAGNode(ctx context.Context, data []byte, links []*IPFSDAGLink, codec string) (*content.CID, error) {
	// Generate CID for the DAG node
	cid, err := ic.contentAddresser.GenerateCID(data, codec)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CID: %w", err)
	}

	// Create DAG node
	dagNode := &IPFSDAGNode{
		CID:     cid,
		Links:   links,
		Data:    data,
		Size:    int64(len(data)),
		Created: time.Now(),
	}

	// Store DAG node
	ic.dagNodes[cid.Hash] = dagNode

	return cid, nil
}

// GetDAGNode retrieves a DAG node by CID
func (ic *IPFSCompatibility) GetDAGNode(ctx context.Context, cid *content.CID) (*IPFSDAGNode, error) {
	dagNode, exists := ic.dagNodes[cid.Hash]
	if !exists {
		return nil, fmt.Errorf("DAG node not found: %s", cid.Hash)
	}

	return dagNode, nil
}

// PinObject pins an object in IPFS-compatible storage
func (ic *IPFSCompatibility) PinObject(ctx context.Context, cid *content.CID, name string, pinType string) error {
	pin := &IPFSPin{
		CID:     cid,
		Name:    name,
		Type:    pinType,
		Created: time.Now(),
	}

	ic.pins[cid.Hash] = pin
	return nil
}

// UnpinObject unpins an object from IPFS-compatible storage
func (ic *IPFSCompatibility) UnpinObject(ctx context.Context, cid *content.CID) error {
	delete(ic.pins, cid.Hash)
	return nil
}

// ListPins lists all pinned objects
func (ic *IPFSCompatibility) ListPins(ctx context.Context) ([]*IPFSPin, error) {
	pins := make([]*IPFSPin, 0, len(ic.pins))
	for _, pin := range ic.pins {
		pins = append(pins, pin)
	}
	return pins, nil
}

// AddNode adds an IPFS node to the network
func (ic *IPFSCompatibility) AddNode(ctx context.Context, node *IPFSNode) error {
	ic.nodes[node.ID] = node
	return nil
}

// GetNode retrieves an IPFS node by ID
func (ic *IPFSCompatibility) GetNode(ctx context.Context, nodeID string) (*IPFSNode, error) {
	node, exists := ic.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	return node, nil
}

// ListNodes lists all IPFS nodes
func (ic *IPFSCompatibility) ListNodes(ctx context.Context) ([]*IPFSNode, error) {
	nodes := make([]*IPFSNode, 0, len(ic.nodes))
	for _, node := range ic.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// ResolvePath resolves an IPFS path to a CID
func (ic *IPFSCompatibility) ResolvePath(ctx context.Context, path string) (*content.CID, error) {
	// Simple path resolution - in a real implementation, this would be more complex
	if strings.HasPrefix(path, "/ipfs/") {
		cidStr := strings.TrimPrefix(path, "/ipfs/")
		return ic.contentAddresser.ParseCID(cidStr)
	}

	return nil, fmt.Errorf("invalid IPFS path: %s", path)
}

// Cat retrieves and returns the data for a given CID
func (ic *IPFSCompatibility) Cat(ctx context.Context, cid *content.CID) (io.Reader, error) {
	// Try to get as block first
	if block, exists := ic.blocks[cid.Hash]; exists {
		return strings.NewReader(string(block.Data)), nil
	}

	// Try to get as DAG node
	if dagNode, exists := ic.dagNodes[cid.Hash]; exists {
		return strings.NewReader(string(dagNode.Data)), nil
	}

	return nil, fmt.Errorf("content not found: %s", cid.Hash)
}

// Stat returns statistics about a CID
func (ic *IPFSCompatibility) Stat(ctx context.Context, cid *content.CID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Check if it's a block
	if block, exists := ic.blocks[cid.Hash]; exists {
		stats["type"] = "block"
		stats["size"] = block.Size
		stats["created"] = block.Created
		return stats, nil
	}

	// Check if it's a DAG node
	if dagNode, exists := ic.dagNodes[cid.Hash]; exists {
		stats["type"] = "dag"
		stats["size"] = dagNode.Size
		stats["links"] = len(dagNode.Links)
		stats["created"] = dagNode.Created
		return stats, nil
	}

	return nil, fmt.Errorf("content not found: %s", cid.Hash)
}

// ExportIPFSData exports data in IPFS format
func (ic *IPFSCompatibility) ExportIPFSData(ctx context.Context, cid *content.CID) ([]byte, error) {
	// Try to get as block first
	if block, exists := ic.blocks[cid.Hash]; exists {
		return json.Marshal(block)
	}

	// Try to get as DAG node
	if dagNode, exists := ic.dagNodes[cid.Hash]; exists {
		return json.Marshal(dagNode)
	}

	return nil, fmt.Errorf("content not found: %s", cid.Hash)
}

// ImportIPFSData imports data from IPFS format
func (ic *IPFSCompatibility) ImportIPFSData(ctx context.Context, data []byte) (*content.CID, error) {
	// Try to parse as block
	var block IPFSBlock
	if err := json.Unmarshal(data, &block); err == nil {
		ic.blocks[block.CID.Hash] = &block
		return block.CID, nil
	}

	// Try to parse as DAG node
	var dagNode IPFSDAGNode
	if err := json.Unmarshal(data, &dagNode); err == nil {
		ic.dagNodes[dagNode.CID.Hash] = &dagNode
		return dagNode.CID, nil
	}

	return nil, fmt.Errorf("failed to parse IPFS data")
}

// GetStorageStats returns storage statistics
func (ic *IPFSCompatibility) GetStorageStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	stats["blocks"] = len(ic.blocks)
	stats["dag_nodes"] = len(ic.dagNodes)
	stats["pins"] = len(ic.pins)
	stats["nodes"] = len(ic.nodes)

	// Calculate total size
	totalSize := int64(0)
	for _, block := range ic.blocks {
		totalSize += block.Size
	}
	for _, dagNode := range ic.dagNodes {
		totalSize += dagNode.Size
	}

	stats["total_size"] = totalSize

	return stats, nil
}
