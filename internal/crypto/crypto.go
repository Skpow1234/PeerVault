package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
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
	hash := md5.Sum([]byte(key))
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

	// Write nonce + ciphertext (GCM.Seal prepends the nonce)
	n, err := dst.Write(ciphertext)
	if err != nil {
		return 0, err
	}

	return n, nil
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

	// Read all encrypted data (nonce + ciphertext + tag)
	ciphertext, err := io.ReadAll(src)
	if err != nil {
		return 0, err
	}

	// GCM.Open expects the nonce to be prepended to the ciphertext
	if len(ciphertext) < gcm.NonceSize() {
		return 0, io.ErrUnexpectedEOF
	}

	// Decrypt with GCM (this also verifies the authentication tag)
	// GCM.Open automatically extracts the nonce from the beginning of ciphertext
	plaintext, err := gcm.Open(nil, nil, ciphertext, nil)
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
