package ipfs

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/content"
	"github.com/stretchr/testify/assert"
)

func TestNewIPFSCompatibility(t *testing.T) {
	ic := NewIPFSCompatibility()

	assert.NotNil(t, ic)
	assert.NotNil(t, ic.contentAddresser)
	assert.NotNil(t, ic.nodes)
	assert.NotNil(t, ic.blocks)
	assert.NotNil(t, ic.dagNodes)
	assert.NotNil(t, ic.pins)
	assert.Equal(t, 0, len(ic.nodes))
	assert.Equal(t, 0, len(ic.blocks))
	assert.Equal(t, 0, len(ic.dagNodes))
	assert.Equal(t, 0, len(ic.pins))
}

func TestIPFSCompatibility_AddBlock(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	data := []byte("test data")
	codec := "raw"

	cid, err := ic.AddBlock(ctx, data, codec)
	assert.NoError(t, err)
	assert.NotNil(t, cid)
	assert.NotEmpty(t, cid.Hash)

	// Verify block was stored
	block, err := ic.GetBlock(ctx, cid)
	assert.NoError(t, err)
	assert.Equal(t, cid, block.CID)
	assert.Equal(t, data, block.Data)
	assert.Equal(t, int64(len(data)), block.Size)
	assert.NotZero(t, block.Created)
}

func TestIPFSCompatibility_GetBlock(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test getting non-existent block
	nonExistentCID := &content.CID{Hash: "nonexistent"}
	_, err := ic.GetBlock(ctx, nonExistentCID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "block not found")

	// Add a block and then retrieve it
	data := []byte("test data")
	cid, err := ic.AddBlock(ctx, data, "raw")
	assert.NoError(t, err)

	block, err := ic.GetBlock(ctx, cid)
	assert.NoError(t, err)
	assert.Equal(t, cid, block.CID)
	assert.Equal(t, data, block.Data)
}

func TestIPFSCompatibility_AddDAGNode(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	data := []byte("dag node data")
	links := []*IPFSDAGLink{
		{
			Name: "child1",
			Size: 100,
			CID:  &content.CID{Hash: "child1_hash"},
		},
		{
			Name: "child2",
			Size: 200,
			CID:  &content.CID{Hash: "child2_hash"},
		},
	}
	codec := "dag-pb"

	cid, err := ic.AddDAGNode(ctx, data, links, codec)
	assert.NoError(t, err)
	assert.NotNil(t, cid)
	assert.NotEmpty(t, cid.Hash)

	// Verify DAG node was stored
	dagNode, err := ic.GetDAGNode(ctx, cid)
	assert.NoError(t, err)
	assert.Equal(t, cid, dagNode.CID)
	assert.Equal(t, data, dagNode.Data)
	assert.Equal(t, links, dagNode.Links)
	assert.Equal(t, int64(len(data)), dagNode.Size)
	assert.NotZero(t, dagNode.Created)
}

func TestIPFSCompatibility_GetDAGNode(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test getting non-existent DAG node
	nonExistentCID := &content.CID{Hash: "nonexistent"}
	_, err := ic.GetDAGNode(ctx, nonExistentCID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DAG node not found")

	// Add a DAG node and then retrieve it
	data := []byte("dag node data")
	cid, err := ic.AddDAGNode(ctx, data, nil, "dag-pb")
	assert.NoError(t, err)

	dagNode, err := ic.GetDAGNode(ctx, cid)
	assert.NoError(t, err)
	assert.Equal(t, cid, dagNode.CID)
	assert.Equal(t, data, dagNode.Data)
}

func TestIPFSCompatibility_PinObject(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	cid := &content.CID{Hash: "test_hash"}
	name := "test_pin"
	pinType := "recursive"

	err := ic.PinObject(ctx, cid, name, pinType)
	assert.NoError(t, err)

	// Verify pin was created
	pins, err := ic.ListPins(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pins))
	assert.Equal(t, cid, pins[0].CID)
	assert.Equal(t, name, pins[0].Name)
	assert.Equal(t, pinType, pins[0].Type)
	assert.NotZero(t, pins[0].Created)
}

func TestIPFSCompatibility_UnpinObject(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	cid := &content.CID{Hash: "test_hash"}

	// Pin object first
	err := ic.PinObject(ctx, cid, "test_pin", "recursive")
	assert.NoError(t, err)

	// Verify pin exists
	pins, err := ic.ListPins(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pins))

	// Unpin object
	err = ic.UnpinObject(ctx, cid)
	assert.NoError(t, err)

	// Verify pin was removed
	pins, err = ic.ListPins(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pins))
}

func TestIPFSCompatibility_ListPins(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test empty pins list
	pins, err := ic.ListPins(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pins))

	// Add multiple pins
	cid1 := &content.CID{Hash: "hash1"}
	cid2 := &content.CID{Hash: "hash2"}

	err = ic.PinObject(ctx, cid1, "pin1", "recursive")
	assert.NoError(t, err)
	err = ic.PinObject(ctx, cid2, "pin2", "direct")
	assert.NoError(t, err)

	// List pins
	pins, err = ic.ListPins(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(pins))

	// Verify both pins are in the list
	pinCIDs := make(map[string]bool)
	for _, pin := range pins {
		pinCIDs[pin.CID.Hash] = true
	}
	assert.True(t, pinCIDs["hash1"])
	assert.True(t, pinCIDs["hash2"])
}

func TestIPFSCompatibility_AddNode(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	node := &IPFSNode{
		ID:       "node1",
		Address:  "127.0.0.1:4001",
		Protocol: "ipfs",
		Version:  "0.12.0",
		Metadata: map[string]interface{}{
			"region": "us-east-1",
		},
	}

	err := ic.AddNode(ctx, node)
	assert.NoError(t, err)

	// Verify node was added
	retrievedNode, err := ic.GetNode(ctx, "node1")
	assert.NoError(t, err)
	assert.Equal(t, node.ID, retrievedNode.ID)
	assert.Equal(t, node.Address, retrievedNode.Address)
	assert.Equal(t, node.Protocol, retrievedNode.Protocol)
	assert.Equal(t, node.Version, retrievedNode.Version)
}

func TestIPFSCompatibility_GetNode(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test getting non-existent node
	_, err := ic.GetNode(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")

	// Add a node and then retrieve it
	node := &IPFSNode{
		ID:      "node1",
		Address: "127.0.0.1:4001",
	}

	err = ic.AddNode(ctx, node)
	assert.NoError(t, err)

	retrievedNode, err := ic.GetNode(ctx, "node1")
	assert.NoError(t, err)
	assert.Equal(t, node.ID, retrievedNode.ID)
	assert.Equal(t, node.Address, retrievedNode.Address)
}

func TestIPFSCompatibility_ListNodes(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test empty nodes list
	nodes, err := ic.ListNodes(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(nodes))

	// Add multiple nodes
	node1 := &IPFSNode{ID: "node1", Address: "127.0.0.1:4001"}
	node2 := &IPFSNode{ID: "node2", Address: "127.0.0.1:4002"}

	err = ic.AddNode(ctx, node1)
	assert.NoError(t, err)
	err = ic.AddNode(ctx, node2)
	assert.NoError(t, err)

	// List nodes
	nodes, err = ic.ListNodes(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(nodes))

	// Verify both nodes are in the list
	nodeIDs := make(map[string]bool)
	for _, node := range nodes {
		nodeIDs[node.ID] = true
	}
	assert.True(t, nodeIDs["node1"])
	assert.True(t, nodeIDs["node2"])
}

func TestIPFSCompatibility_ResolvePath(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test valid IPFS path
	path := "/ipfs/QmTestHash"
	cid, err := ic.ResolvePath(ctx, path)
	assert.NoError(t, err)
	assert.NotNil(t, cid)
	assert.Equal(t, "QmTestHash", cid.Hash)

	// Test invalid path
	invalidPath := "/invalid/path"
	_, err = ic.ResolvePath(ctx, invalidPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IPFS path")

	// Test path without /ipfs/ prefix
	noPrefixPath := "QmTestHash"
	_, err = ic.ResolvePath(ctx, noPrefixPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IPFS path")
}

func TestIPFSCompatibility_Cat(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test getting non-existent content
	nonExistentCID := &content.CID{Hash: "nonexistent"}
	_, err := ic.Cat(ctx, nonExistentCID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content not found")

	// Add a block and retrieve its data
	data := []byte("test data")
	cid, err := ic.AddBlock(ctx, data, "raw")
	assert.NoError(t, err)

	reader, err := ic.Cat(ctx, cid)
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	// Read the data
	readData, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, data, readData)

	// Add a DAG node and retrieve its data
	dagData := []byte("dag data")
	dagCID, err := ic.AddDAGNode(ctx, dagData, nil, "dag-pb")
	assert.NoError(t, err)

	dagReader, err := ic.Cat(ctx, dagCID)
	assert.NoError(t, err)
	assert.NotNil(t, dagReader)

	// Read the DAG data
	readDagData, err := io.ReadAll(dagReader)
	assert.NoError(t, err)
	assert.Equal(t, dagData, readDagData)
}

func TestIPFSCompatibility_Stat(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test getting stats for non-existent content
	nonExistentCID := &content.CID{Hash: "nonexistent"}
	_, err := ic.Stat(ctx, nonExistentCID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content not found")

	// Add a block and get its stats
	data := []byte("test data")
	cid, err := ic.AddBlock(ctx, data, "raw")
	assert.NoError(t, err)

	stats, err := ic.Stat(ctx, cid)
	assert.NoError(t, err)
	assert.Equal(t, "block", stats["type"])
	assert.Equal(t, int64(len(data)), stats["size"])
	assert.NotNil(t, stats["created"])

	// Add a DAG node and get its stats
	dagData := []byte("dag data")
	links := []*IPFSDAGLink{
		{Name: "link1", Size: 100, CID: &content.CID{Hash: "link1_hash"}},
	}
	dagCID, err := ic.AddDAGNode(ctx, dagData, links, "dag-pb")
	assert.NoError(t, err)

	dagStats, err := ic.Stat(ctx, dagCID)
	assert.NoError(t, err)
	assert.Equal(t, "dag", dagStats["type"])
	assert.Equal(t, int64(len(dagData)), dagStats["size"])
	assert.Equal(t, 1, dagStats["links"])
	assert.NotNil(t, dagStats["created"])
}

func TestIPFSCompatibility_ExportIPFSData(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test exporting non-existent content
	nonExistentCID := &content.CID{Hash: "nonexistent"}
	_, err := ic.ExportIPFSData(ctx, nonExistentCID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content not found")

	// Add a block and export it
	data := []byte("test data")
	cid, err := ic.AddBlock(ctx, data, "raw")
	assert.NoError(t, err)

	exportedData, err := ic.ExportIPFSData(ctx, cid)
	assert.NoError(t, err)
	assert.NotEmpty(t, exportedData)

	// Verify exported data is valid JSON
	var exportedBlock IPFSBlock
	err = json.Unmarshal(exportedData, &exportedBlock)
	assert.NoError(t, err)
	assert.Equal(t, cid.Hash, exportedBlock.CID.Hash)
	assert.Equal(t, data, exportedBlock.Data)

	// Add a DAG node and export it
	dagData := []byte("dag data")
	dagCID, err := ic.AddDAGNode(ctx, dagData, nil, "dag-pb")
	assert.NoError(t, err)

	exportedDagData, err := ic.ExportIPFSData(ctx, dagCID)
	assert.NoError(t, err)
	assert.NotEmpty(t, exportedDagData)

	// Verify exported DAG data is valid JSON
	var exportedDagNode IPFSDAGNode
	err = json.Unmarshal(exportedDagData, &exportedDagNode)
	assert.NoError(t, err)
	assert.Equal(t, dagCID.Hash, exportedDagNode.CID.Hash)
	assert.Equal(t, dagData, exportedDagNode.Data)
}

func TestIPFSCompatibility_ImportIPFSData(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test importing invalid data
	invalidData := []byte("invalid json")
	_, err := ic.ImportIPFSData(ctx, invalidData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse IPFS data")

	// Create and import a block
	originalBlock := &IPFSBlock{
		CID:     &content.CID{Hash: "imported_block_hash"},
		Data:    []byte("imported block data"),
		Size:    20,
		Created: time.Now(),
	}

	blockData, err := json.Marshal(originalBlock)
	assert.NoError(t, err)

	importedCID, err := ic.ImportIPFSData(ctx, blockData)
	assert.NoError(t, err)
	assert.Equal(t, originalBlock.CID.Hash, importedCID.Hash)

	// Verify block was imported
	importedBlock, err := ic.GetBlock(ctx, importedCID)
	assert.NoError(t, err)
	assert.Equal(t, originalBlock.Data, importedBlock.Data)
	assert.Equal(t, originalBlock.Size, importedBlock.Size)

	// Create and import a DAG node
	originalDagNode := &IPFSDAGNode{
		CID:     &content.CID{Hash: "imported_dag_hash"},
		Data:    []byte("imported dag data"),
		Size:    18,
		Created: time.Now(),
		Links: []*IPFSDAGLink{
			{
				Name: "child1",
				Size: 100,
				CID:  &content.CID{Hash: "child1_hash"},
			},
		},
	}

	dagData, err := json.Marshal(originalDagNode)
	assert.NoError(t, err)

	importedDagCID, err := ic.ImportIPFSData(ctx, dagData)
	assert.NoError(t, err)
	assert.NotNil(t, importedDagCID)
	assert.Equal(t, originalDagNode.CID.Hash, importedDagCID.Hash)

	// Verify DAG node was imported
	importedDagNode, err := ic.GetDAGNode(ctx, importedDagCID)
	assert.NoError(t, err)
	assert.NotNil(t, importedDagNode)
	assert.Equal(t, originalDagNode.Data, importedDagNode.Data)
	assert.Equal(t, originalDagNode.Size, importedDagNode.Size)
	assert.Equal(t, len(originalDagNode.Links), len(importedDagNode.Links))
}

func TestIPFSCompatibility_GetStorageStats(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test initial stats
	stats, err := ic.GetStorageStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, stats["blocks"])
	assert.Equal(t, 0, stats["dag_nodes"])
	assert.Equal(t, 0, stats["pins"])
	assert.Equal(t, 0, stats["nodes"])
	assert.Equal(t, int64(0), stats["total_size"])

	// Add some content
	data1 := []byte("data1")
	data2 := []byte("data2")
	dagData := []byte("dag data")

	cid1, err := ic.AddBlock(ctx, data1, "raw")
	assert.NoError(t, err)
	_, err = ic.AddBlock(ctx, data2, "raw")
	assert.NoError(t, err)
	dagCID, err := ic.AddDAGNode(ctx, dagData, nil, "dag-pb")
	assert.NoError(t, err)

	// Add pins
	err = ic.PinObject(ctx, cid1, "pin1", "recursive")
	assert.NoError(t, err)
	err = ic.PinObject(ctx, dagCID, "pin2", "direct")
	assert.NoError(t, err)

	// Add nodes
	node := &IPFSNode{ID: "node1", Address: "127.0.0.1:4001"}
	err = ic.AddNode(ctx, node)
	assert.NoError(t, err)

	// Test updated stats
	stats, err = ic.GetStorageStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, stats["blocks"])
	assert.Equal(t, 1, stats["dag_nodes"])
	assert.Equal(t, 2, stats["pins"])
	assert.Equal(t, 1, stats["nodes"])

	expectedTotalSize := int64(len(data1) + len(data2) + len(dagData))
	assert.Equal(t, expectedTotalSize, stats["total_size"])
}

func TestIPFSCompatibility_EdgeCases(t *testing.T) {
	ic := NewIPFSCompatibility()
	ctx := context.Background()

	// Test with empty data
	emptyData := []byte("")
	cid, err := ic.AddBlock(ctx, emptyData, "raw")
	assert.NoError(t, err)
	assert.NotNil(t, cid)

	block, err := ic.GetBlock(ctx, cid)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), block.Size)
	assert.Equal(t, emptyData, block.Data)

	// Test with large data
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	largeCID, err := ic.AddBlock(ctx, largeData, "raw")
	assert.NoError(t, err)
	assert.NotNil(t, largeCID)

	largeBlock, err := ic.GetBlock(ctx, largeCID)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(largeData)), largeBlock.Size)
	assert.Equal(t, largeData, largeBlock.Data)

	// Test with special characters in data
	specialData := []byte("Hello, 世界! 🌍")
	specialCID, err := ic.AddBlock(ctx, specialData, "raw")
	assert.NoError(t, err)
	assert.NotNil(t, specialCID)

	specialBlock, err := ic.GetBlock(ctx, specialCID)
	assert.NoError(t, err)
	assert.Equal(t, specialData, specialBlock.Data)
}
