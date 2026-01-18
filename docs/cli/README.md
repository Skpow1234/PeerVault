# PeerVault CLI

A comprehensive command-line interface for PeerVault distributed storage system.

## Features

- **Interactive Shell**: Full-featured REPL with command history and tab completion
- **File Operations**: Store, retrieve, list, and delete files
- **Peer Management**: Connect to nodes, manage peer connections
- **System Monitoring**: Health checks, metrics, and status monitoring
- **Multiple Output Formats**: Table, JSON, and YAML output
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Installation

The CLI is not part of the Docker-only workflow. Use the APIs and web interfaces instead.

## Usage

### Interactive Mode

Interactive CLI execution is not available in the Docker-only workflow.

### Command Line Arguments

CLI flags are not available in the Docker-only workflow.

## Commands

### File Operations

#### Store a File

```bash
peervault> store ./documents/report.pdf
📁 Storing file: report.pdf
✅ File stored successfully: QmAbCdEf...
```

#### Retrieve a File

```bash
peervault> get QmAbCdEf... --output=./downloaded-report.pdf
📥 Downloading from peer network...
✅ Downloaded: 2.3MB in 1.2s
```

#### List Files

```bash
peervault> list
📁 Files (15 total)
┌─────────────────────────────────────────────────────────────┬─────────────┬─────────────────────────────────────────────────────────────┐
│ Key                                                         │ Size        │ Created At                                               │
├─────────────────────────────────────────────────────────────┼─────────────┼─────────────────────────────────────────────────────────────┤
│ documents/report.pdf                                        │ 2.3 MB      │ 2024-01-15 14:30:25                                      │
│ images/photo.jpg                                            │ 1.8 MB      │ 2024-01-15 14:25:10                                      │
└─────────────────────────────────────────────────────────────┴─────────────┴─────────────────────────────────────────────────────────────┘
```

#### Delete a File

```bash
peervault> delete QmAbCdEf...
🗑️ Deleting file: QmAbCdEf...
✅ File deleted successfully: QmAbCdEf...
```

### Peer Management

#### List Peers

```bash
peervault> peers list
🌐 Peers (3 total)
┌─────────────────────────────────────────────────────────────┬─────────────┬─────────────┬─────────────┬─────────────────────────────────────────────────────────────┐
│ Address                                                     │ Status      │ Latency     │ Storage     │ Last Seen                                               │
├─────────────────────────────────────────────────────────────┼─────────────┼─────────────┼─────────────┼─────────────────────────────────────────────────────────────┤
│ node1:3000                                                  │ 🟢 Healthy  │ 12ms        │ 45.2GB/1TB  │ 2s ago                                                  │
│ node2:7000                                                  │ 🟡 Degraded │ 156ms       │ 892.1GB/1TB │ 45s ago                                                 │
│ node3:5000                                                  │ 🟢 Healthy  │ 8ms         │ 234.7GB/1TB │ 1s ago                                                  │
└─────────────────────────────────────────────────────────────┴─────────────┴─────────────┴─────────────┴─────────────────────────────────────────────────────────────┘
```

#### Add a Peer

```bash
peervault> peers add node4.example.com:3000
🔍 Adding peer: node4.example.com:3000
✅ Peer added successfully: peer-123
```

#### Remove a Peer

```bash
peervault> peers remove peer-123
🗑️ Removing peer: peer-123
✅ Peer removed successfully: peer-123
```

### System Monitoring

#### Health Check

```bash
peervault> health
🏥 System Health
┌─────────────────┬─────────────────────────────────────────────────────────────┐
│ Field           │ Value                                                       │
├─────────────────┼─────────────────────────────────────────────────────────────┤
│ Overall Status  │ 🟢 healthy                                                  │
│ Timestamp       │ 2024-01-15T14:30:25Z                                       │
└─────────────────┴─────────────────────────────────────────────────────────────┘

🔧 Service Status
┌─────────────────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────┐
│ Service                                                     │ Status                                                       │
├─────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────┤
│ storage                                                     │ 🟢 healthy                                                  │
│ network                                                     │ 🟢 healthy                                                  │
│ encryption                                                  │ 🟢 healthy                                                  │
└─────────────────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────┘
```

#### System Metrics

```bash
peervault> metrics
📊 System Metrics
┌─────────────────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────┐
│ Metric                                                      │ Value                                                       │
├─────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────┤
│ Files Stored                                                │ 1,247                                                       │
│ Network Traffic (MB/s)                                      │ 45.2                                                        │
│ Active Peers                                                │ 12                                                          │
│ Storage Used                                                │ 2.1 TB                                                      │
└─────────────────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────┘
```

#### Live Metrics

```bash
peervault> metrics --live
📊 Live System Metrics (refreshing every 5s)
┌─────────────────┬─────────────┬─────────────┬─────────────┐
│ Metric          │ Current     │ 1h Avg      │ 24h Avg     │
├─────────────────┼─────────────┼─────────────┼─────────────┤
│ Files Stored    │ 1,247       │ 1,203       │ 1,156       │
│ Network Traffic │ 45.2 MB/s   │ 42.1 MB/s   │ 38.7 MB/s   │
│ Active Peers    │ 12          │ 11.8        │ 11.2        │
│ Storage Used    │ 2.1 TB      │ 2.0 TB      │ 1.9 TB      │
└─────────────────┴─────────────┴─────────────┴─────────────┘
```

### Connection Management

#### Connect to a Node

```bash
peervault> connect node1.example.com:3000
🔗 Connecting to: node1.example.com:3000
✅ Connected to: node1.example.com:3000
peervault[node1]> # Prompt changes to show current connection
```

#### Disconnect

```bash
peervault[node1]> disconnect
🔌 Disconnecting...
✅ Disconnected
peervault> # Back to default prompt
```

### Utility Commands

#### Help

```bash
peervault> help
🚀 PeerVault CLI - Available Commands

📁 File Operations:
  store           - Store a file in the PeerVault network
  get             - Retrieve a file from the PeerVault network
  list            - List files in the PeerVault network
  delete          - Delete a file from the PeerVault network

🌐 Network Operations:
  peers           - Manage peer connections
  connect         - Connect to a PeerVault node
  disconnect      - Disconnect from current node

🔧 System Operations:
  health          - Check system health
  metrics         - Show system metrics
  status          - Show system status

⚙️  Utility Commands:
  help            - Show help information
  exit            - Exit the CLI
  clear           - Clear the screen
  history         - Show command history

Type 'help <command>' for detailed information about a specific command.
```

#### Command History

```bash
peervault> history
📜 Command History:
  1  help
  2  peers list
  3  health
  4  metrics
  5  store document.pdf
  6  get QmAbCdEf...
```

#### Clear Screen

```bash
peervault> clear
# Screen is cleared
```

## Configuration

The CLI stores configuration in `~/.peervault/config.json`:

```json
{
  "server_url": "http://localhost:8080",
  "auth_token": "demo-token",
  "output_format": "table",
  "theme": "default",
  "auto_complete": true,
  "verbose": false
}
```

### Configuration Commands

```bash
# Show current configuration
peervault> config show

# Set configuration values
peervault> set server_url http://node1.example.com:8080
peervault> set output_format json
peervault> set verbose true

# Get configuration values
peervault> get server_url
peervault> get output_format
```

## Output Formats

### Table Format (Default)

```bash
peervault> list
# Shows formatted tables with borders and alignment
```

### JSON Format

```bash
peervault> set output_format json
peervault> list
{
  "files": [
    {
      "id": "file-123",
      "key": "document.pdf",
      "size": 2457600,
      "hash": "QmAbCdEf...",
      "created_at": "2024-01-15T14:30:25Z",
      "owner": "user-123"
    }
  ],
  "total": 1
}
```

### YAML Format

```bash
peervault> set output_format yaml
peervault> list
files:
  - id: file-123
    key: document.pdf
    size: 2457600
    hash: QmAbCdEf...
    created_at: 2024-01-15T14:30:25Z
    owner: user-123
total: 1
```

## Advanced Features

### Tab Completion

- Press `Tab` to complete commands and file paths
- Works with command arguments and options

### Command History Arrows

- Use `↑` and `↓` arrows to navigate command history
- History is persisted between sessions

### Aliases

```bash
# Built-in aliases
ls = list
bc = blockchain
quit = exit
```

### Scripting

CLI scripting is not available in the Docker-only workflow.

## Integration

### With PeerVault APIs

The CLI automatically connects to PeerVault's REST API by default, but can be configured to use:

- GraphQL API
- gRPC API
- WebSocket API

### With External Tools

```bash
# Export data for analysis
peervault> metrics --format json > metrics.json

# Chain with other tools
peervault> peers list --format json | jq '.peers[].status'
```

## Troubleshooting

### Connection Issues

```bash
# Check if server is running
peervault> health

# Test connection
peervault> connect localhost:8080
```

### Authentication Issues

```bash
# Set auth token
peervault> set auth_token your-token-here

Command line arguments are not available in the Docker-only workflow.
```

### Performance Issues

```bash
# Check system metrics
peervault> metrics

# Monitor live performance
peervault> metrics --live
```

## Development

### Building

Build steps are handled in CI using containerized workflows.

### Testing

CLI tests run in CI using containerized workflows.

### Contributing

1. Add new commands in `internal/cli/commands/`
2. Update help text and documentation
3. Add tests for new functionality
4. Update this README

## License

Same as PeerVault project license.
