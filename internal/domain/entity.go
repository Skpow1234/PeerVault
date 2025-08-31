package domain

// NodeID represents the unique identifier of a node in the network.
type NodeID string

// FileKey represents the logical key clients use to identify a file.
// This key may differ from the physical storage key used on disk.
type FileKey string

// FileMetadata captures information about a file that is stored or requested
// within the distributed store.
type FileMetadata struct {
	// ID is the node identifier that owns or serves the file.
	ID NodeID

	// LogicalKey is the client-provided key for the file.
	LogicalKey string

	// HashedKey is the transformed key used internally for CAS storage.
	HashedKey string

	// Size is the expected size in bytes when streaming (may be 0 for requests).
	Size int64
}