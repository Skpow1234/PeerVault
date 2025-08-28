package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"time"
)

const (
	// GCM nonce size is 12 bytes (96 bits) for optimal performance
	GCMNonceSize = 12
	// GCM tag size is 16 bytes (128 bits)
	GCMTagSize = 16
	// Key derivation constants
	KeyDerivationSalt = "peervault-cluster-salt-v1"
	KeyRotationPeriod = 24 * time.Hour // Rotate keys every 24 hours
)

// KeyManager handles encryption key generation, derivation, and rotation
type KeyManager struct {
	clusterKey []byte
	derivedKey []byte
	keyID      string
	createdAt  time.Time
}

// NewKeyManager creates a new key manager with proper key derivation
func NewKeyManager() (*KeyManager, error) {
	// Try to load cluster key from environment first
	clusterKey := os.Getenv("PEERVAULT_CLUSTER_KEY")
	if clusterKey == "" {
		// Generate a new cluster key if not provided
		clusterKeyBytes := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, clusterKeyBytes); err != nil {
			return nil, err
		}
		clusterKey = hex.EncodeToString(clusterKeyBytes)
		// In production, you'd want to securely store this key
		// For demo purposes, we'll just use it in memory
	}

	// Decode the cluster key
	clusterKeyBytes, err := hex.DecodeString(clusterKey)
	if err != nil {
		return nil, err
	}

	// Derive the encryption key using HKDF-like approach
	derivedKey := deriveKey(clusterKeyBytes, KeyDerivationSalt)

	// Generate a key ID for rotation tracking
	keyID := generateKeyID(derivedKey)

	return &KeyManager{
		clusterKey: clusterKeyBytes,
		derivedKey: derivedKey,
		keyID:      keyID,
		createdAt:  time.Now(),
	}, nil
}

// deriveKey derives an encryption key from the cluster key using HMAC-SHA256
func deriveKey(clusterKey []byte, salt string) []byte {
	h := hmac.New(sha256.New, clusterKey)
	h.Write([]byte(salt))
	return h.Sum(nil)
}

// generateKeyID creates a unique identifier for the key
func generateKeyID(key []byte) string {
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:8]) // First 8 bytes as key ID
}

// GetEncryptionKey returns the current encryption key
func (km *KeyManager) GetEncryptionKey() []byte {
	return km.derivedKey
}

// GetKeyID returns the current key identifier
func (km *KeyManager) GetKeyID() string {
	return km.keyID
}

// ShouldRotate checks if the key should be rotated
func (km *KeyManager) ShouldRotate() bool {
	return time.Since(km.createdAt) > KeyRotationPeriod
}

// RotateKey generates a new derived key
func (km *KeyManager) RotateKey() error {
	// For demo simplicity, we'll just regenerate the derived key
	// In production, you'd want to implement proper key rotation with key versioning
	km.derivedKey = deriveKey(km.clusterKey, KeyDerivationSalt+time.Now().Format("2006-01-02"))
	km.keyID = generateKeyID(km.derivedKey)
	km.createdAt = time.Now()
	return nil
}

func GenerateID() string {
	buf := make([]byte, 32)
	io.ReadFull(rand.Reader, buf)
	return hex.EncodeToString(buf)
}

func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// NewEncryptionKey is deprecated - use KeyManager instead
// This is kept for backward compatibility
func NewEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}

// CopyEncrypt encrypts data using AES-GCM and writes the nonce + ciphertext + tag
func CopyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return 0, err
	}

	// Read all data from source
	plaintext, err := io.ReadAll(src)
	if err != nil {
		return 0, err
	}

	// Encrypt with GCM
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Write nonce + ciphertext
	if _, err := dst.Write(nonce); err != nil {
		return 0, err
	}
	n, err := dst.Write(ciphertext)
	if err != nil {
		return 0, err
	}

	return n + len(nonce), nil
}

// CopyDecrypt decrypts data using AES-GCM, reading nonce + ciphertext + tag
func CopyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	// Read nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(src, nonce); err != nil {
		return 0, err
	}

	// Read ciphertext
	ciphertext, err := io.ReadAll(src)
	if err != nil {
		return 0, err
	}

	// Decrypt with GCM
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, err
	}

	// Write decrypted data
	n, err := dst.Write(plaintext)
	if err != nil {
		return 0, err
	}

	return n, nil
}
