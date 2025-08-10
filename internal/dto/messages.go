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
