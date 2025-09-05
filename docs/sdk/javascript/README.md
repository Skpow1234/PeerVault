# PeerVault JavaScript/TypeScript SDK

Official JavaScript/TypeScript SDK for PeerVault distributed file storage system.

## Installation

```bash
npm install @peervault/sdk
# or
yarn add @peervault/sdk
```

## Quick Start

```typescript
import { PeerVaultClient } from '@peervault/sdk';

// Create client
const client = new PeerVaultClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'your-api-key' // optional
});

// Upload file
const file = new File(['Hello World'], 'example.txt');
const result = await client.files.upload('my-key', file);

console.log(`Uploaded file: ${result.key} (size: ${result.size})`);
```

## Authentication

```typescript
// Login to get token
const auth = await client.auth.login('username', 'password');

// Set token for subsequent requests
client.setToken(auth.token);

// Or use API key
const client = new PeerVaultClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'your-api-key'
});
```

## File Operations

### Upload File

```typescript
// From File object
const file = new File(['content'], 'example.txt');
const result = await client.files.upload('my-key', file);

// From Blob
const blob = new Blob(['content']);
const result = await client.files.upload('my-key', blob);

// From ArrayBuffer
const buffer = new ArrayBuffer(1024);
const result = await client.files.upload('my-key', buffer);

// From string
const result = await client.files.uploadString('my-key', 'content');
```

### Download File

```typescript
// As Blob
const blob = await client.files.download('my-key');

// As ArrayBuffer
const buffer = await client.files.downloadBuffer('my-key');

// As string
const content = await client.files.downloadString('my-key');

// Save to file (browser)
const blob = await client.files.download('my-key');
const url = URL.createObjectURL(blob);
const a = document.createElement('a');
a.href = url;
a.download = 'downloaded-file.txt';
a.click();
```

### List Files

```typescript
const files = await client.files.list({
  limit: 10,
  offset: 0,
  prefix: 'documents/'
});

files.forEach(file => {
  console.log(`File: ${file.key} (size: ${file.size})`);
});
```

### Delete File

```typescript
await client.files.delete('my-key');
```

## Peer Operations

### List Peers

```typescript
const peers = await client.peers.list();

peers.forEach(peer => {
  console.log(`Peer: ${peer.id} (status: ${peer.status})`);
});
```

### Get Peer Health

```typescript
const health = await client.peers.getHealth('peer-id');
console.log(`Health: ${health.status} (uptime: ${health.uptime})`);
```

## GraphQL Client

```typescript
import { GraphQLClient } from '@peervault/sdk/graphql';

// Create GraphQL client
const gqlClient = new GraphQLClient('http://localhost:8081');

// Execute query
const result = await gqlClient.query(`
  query {
    files {
      key
      size
      createdAt
    }
  }
`);

// Execute mutation
const uploadResult = await gqlClient.mutate(`
  mutation UploadFile($input: FileUploadInput!) {
    uploadFile(input: $input) {
      key
      size
      success
    }
  }
`, {
  input: {
    key: 'my-file',
    content: 'file content'
  }
});

// Subscribe to real-time updates
const subscription = gqlClient.subscribe(`
  subscription {
    fileEvents {
      type
      file {
        key
        size
      }
    }
  }
`, (data) => {
  console.log('File event:', data);
});
```

## WebSocket Client

```typescript
import { WebSocketClient } from '@peervault/sdk/websocket';

// Create WebSocket client
const wsClient = new WebSocketClient('ws://localhost:8082');

// Connect
await wsClient.connect();

// Subscribe to events
wsClient.subscribe('file.uploaded', (data) => {
  console.log('File uploaded:', data);
});

wsClient.subscribe('peer.connected', (data) => {
  console.log('Peer connected:', data);
});

// Send message
wsClient.send('ping', { timestamp: Date.now() });
```

## Configuration

```typescript
const client = new PeerVaultClient({
  baseURL: 'http://localhost:8080',
  timeout: 30000,
  retries: 3,
  rateLimit: 100, // requests per minute
  headers: {
    'User-Agent': 'my-app/1.0'
  }
});
```

## Error Handling

```typescript
try {
  const result = await client.files.upload('key', file);
} catch (error) {
  if (error instanceof FileNotFoundError) {
    console.log(`File not found: ${error.key}`);
  } else if (error instanceof RateLimitError) {
    console.log(`Rate limited: retry after ${error.retryAfter}ms`);
  } else if (error instanceof AuthError) {
    console.log(`Authentication failed: ${error.message}`);
  } else {
    console.log(`Unexpected error: ${error.message}`);
  }
}
```

## Streaming

```typescript
// Stream upload
const uploader = client.files.createUploader('large-file');

// Upload in chunks
const chunkSize = 64 * 1024; // 64KB
const file = new File(['large content...'], 'large-file.bin');

for (let offset = 0; offset < file.size; offset += chunkSize) {
  const chunk = file.slice(offset, offset + chunkSize);
  await uploader.write(chunk);
}

const result = await uploader.finish();
```

## Webhooks

```typescript
// Verify webhook signature
import { verifyWebhookSignature } from '@peervault/sdk/webhooks';

function handleWebhook(payload: string, signature: string, secret: string) {
  if (!verifyWebhookSignature(payload, signature, secret)) {
    throw new Error('Invalid webhook signature');
  }
  
  const event = JSON.parse(payload);
  console.log('Webhook event:', event);
}
```

## React Integration

```typescript
import { usePeerVault } from '@peervault/sdk/react';

function FileUploader() {
  const { upload, download, list } = usePeerVault();
  const [files, setFiles] = useState([]);
  
  const handleUpload = async (file: File) => {
    const result = await upload(file.name, file);
    setFiles(prev => [...prev, result]);
  };
  
  return (
    <div>
      <input type="file" onChange={(e) => handleUpload(e.target.files[0])} />
      {files.map(file => (
        <div key={file.key}>
          {file.key} ({file.size} bytes)
        </div>
      ))}
    </div>
  );
}
```

## Node.js Usage

```typescript
import { PeerVaultClient } from '@peervault/sdk';
import fs from 'fs';

const client = new PeerVaultClient({
  baseURL: 'http://localhost:8080'
});

// Upload file from filesystem
const fileStream = fs.createReadStream('example.txt');
const result = await client.files.uploadStream('my-key', fileStream);

// Download to filesystem
const downloadStream = fs.createWriteStream('downloaded.txt');
await client.files.downloadStream('my-key', downloadStream);
```

## Examples

See the [examples directory](examples/) for complete working examples:

- [Basic file operations](examples/basic/)
- [Authentication](examples/auth/)
- [GraphQL queries](examples/graphql/)
- [WebSocket real-time](examples/websocket/)
- [React integration](examples/react/)
- [Node.js usage](examples/nodejs/)
- [Webhook handling](examples/webhooks/)
- [Error handling](examples/errors/)
