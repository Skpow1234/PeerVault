package dto

// StoreFile announces an incoming file to peers so they can prepare to receive it.
type StoreFile struct {
	ID   string
	Key  string
	Size int64
}

// GetFile requests a file from peers.
type GetFile struct {
	ID  string
	Key string
}

// GetFileAck acknowledges a GetFile request
type GetFileAck struct {
	RequestID string // ID of the original GetFile request
	Key       string
	HasFile   bool
	FileSize  int64
}

// StoreFileAck acknowledges a StoreFile request
type StoreFileAck struct {
	RequestID string // ID of the original StoreFile request
	Key       string
	Success   bool
	Error     string // Empty if success
}
