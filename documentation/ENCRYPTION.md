# Encryption Strategy

## Overview

This document describes the encryption strategy implemented in the PeerVault distributed storage system.

## Encryption Approach

### **Encryption at Rest + Encryption in Transit**

The system implements a comprehensive encryption strategy that protects data both when stored locally and when transmitted over the network.

## Encryption Strategy Details

### 1. Encryption at Rest

- **What**: All data stored locally on disk is encrypted using AES-GCM
- **When**: Data is encrypted before being written to disk
- **How**: Uses `crypto.CopyEncrypt()` with a unique nonce for each file
- **Key Management**: Uses the KeyManager for key derivation and rotation

**Storage Flow:**

```bash
Plaintext Data → AES-GCM Encryption → Encrypted Data → Disk Storage
```

**Retrieval Flow:**

```bash
Disk Storage → Encrypted Data → AES-GCM Decryption → Plaintext Data
```

### 2. Encryption in Transit

- **What**: All data transmitted between peers is encrypted
- **When**: Data is encrypted before network transmission
- **How**: Uses `crypto.CopyEncrypt()` for streaming encryption
- **Key Management**: Same encryption keys used for both at-rest and in-transit

**Network Transmission Flow:**

```bash
Encrypted Data from Disk → Network Transmission → Encrypted Data to Disk
```

## Technical Implementation

### AES-GCM Encryption

- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key Size**: 256 bits
- **Nonce Size**: 12 bytes (96 bits)
- **Authentication**: Built-in message authentication (AEAD)
- **Overhead**: ~28 bytes per file (nonce + tag)

### Key Management

- **Key Derivation**: HMAC-SHA256 from cluster key
- **Key Rotation**: Automatic key rotation with configurable intervals
- **Key Storage**: Environment variables or secure key management
- **Fallback**: Legacy key generation for backward compatibility

### File Storage Format

```bash
[Nonce: 12 bytes][Ciphertext: variable][Authentication Tag: 16 bytes]
```

## Security Benefits

### 1. Data Confidentiality

- Files are unreadable without the encryption key
- Protection against unauthorized access to stored files
- Protection against network interception

### 2. Data Integrity

- AES-GCM provides built-in authentication
- Detection of tampering or corruption
- Protection against replay attacks

### 3. Key Security

- Keys are derived from a master cluster key
- Automatic key rotation reduces exposure window
- No hardcoded keys in the application

## Performance Considerations

### Encryption Overhead

- **CPU**: Minimal overhead with AES-NI hardware acceleration
- **Storage**: ~28 bytes per file for nonce and authentication tag
- **Memory**: Streaming encryption/decryption without buffering entire files

### Scalability

- Encryption/decryption scales linearly with file size
- No performance impact on small files
- Efficient for large file streaming

## Configuration

### Environment Variables

```bash
# Master cluster key (required for key derivation)
PEERVAULT_CLUSTER_KEY=your-secure-cluster-key

# Authentication token for peer verification
PEERVAULT_AUTH_TOKEN=your-auth-token
```

### Key Rotation Settings

```go
// Default key rotation interval: 24 hours
// Can be configured via KeyManager options
keyManager := crypto.NewKeyManager()
```

## Migration and Compatibility

### Backward Compatibility

- Legacy unencrypted files are still supported
- Automatic fallback to legacy key generation
- Gradual migration to encrypted storage

### File Format Changes

- New files are automatically encrypted
- Existing files remain in their current format
- No automatic re-encryption of existing files

## Best Practices

### 1. Key Management

- Use strong, randomly generated cluster keys
- Rotate keys regularly
- Store keys securely (not in code or config files)

### 2. Network Security

- Use TLS for additional network security if needed
- Implement proper peer authentication
- Monitor for unauthorized access attempts

### 3. Storage Security

- Ensure disk encryption at the OS level
- Protect against physical access to storage
- Regular security audits and updates

## Testing

The encryption strategy is thoroughly tested with:

- **Unit Tests**: Individual encryption/decryption functions
- **Integration Tests**: End-to-end file storage and retrieval
- **Performance Tests**: Large file encryption/decryption
- **Security Tests**: Key rotation and authentication

Run tests with:

```bash
go test encryption_strategy_test.go -v
```

## Conclusion

The implemented encryption strategy provides comprehensive protection for data both at rest and in transit, ensuring confidentiality, integrity, and authenticity while maintaining good performance characteristics.
