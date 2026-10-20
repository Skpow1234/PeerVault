# Plugins

PeerVault defines a plugin-oriented extension model for storage, auth, transport, and processing.

## Plugin types

- **Storage** — alternate backends (e.g. S3)
- **Authentication** — custom auth mechanisms
- **Transport** — alternate protocols
- **Processing** — transform / pipeline hooks

## Example sketch (S3 storage)

```go
package s3_storage

import "github.com/Skpow1234/Peervault/internal/plugins"

type S3StoragePlugin struct {
	bucket string
}

func (p *S3StoragePlugin) Store(key string, data []byte) error {
	// implement store
	return nil
}

func (p *S3StoragePlugin) Retrieve(key string) ([]byte, error) {
	// implement retrieve
	return nil, nil
}
```

## Developing a plugin

1. Create a module under `plugins/` (see `plugins/s3-storage/`)
2. Implement the interfaces in `internal/plugins`
3. Register and configure the plugin with your node/API process

Full guide: [internal/plugins/README.md](../internal/plugins/README.md).
