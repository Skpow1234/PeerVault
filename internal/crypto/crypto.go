package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

const (
	// GCM nonce size is 12 bytes (96 bits) for optimal performance
	GCMNonceSize = 12
	// GCM tag size is 16 bytes (128 bits)
	GCMTagSize = 16
)

func GenerateID() string {
	buf := make([]byte, 32)
	io.ReadFull(rand.Reader, buf)
	return hex.EncodeToString(buf)
}

func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

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
