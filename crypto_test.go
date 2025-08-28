package main

import (
	"bytes"
	"fmt"
	"testing"

	internalcrypto "github.com/anthdm/foreverstore/internal/crypto"
)

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo not bar"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := internalcrypto.NewEncryptionKey()
	_, err := internalcrypto.CopyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Original payload length:", len(payload))
	fmt.Println("Encrypted data length:", len(dst.String()))

	// Create a new reader from the encrypted data
	encryptedData := dst.Bytes()
	encryptedReader := bytes.NewReader(encryptedData)

	out := new(bytes.Buffer)
	nw, err := internalcrypto.CopyDecrypt(key, encryptedReader, out)
	if err != nil {
		t.Error(err)
	}

	// With AES-GCM: nonce (12 bytes) + ciphertext + tag (16 bytes)
	// The exact size depends on the ciphertext length, but it should be larger than original
	encryptedSize := len(encryptedData)
	if encryptedSize <= len(payload) {
		t.Errorf("encrypted size (%d) should be larger than original (%d)", encryptedSize, len(payload))
	}

	if out.String() != payload {
		t.Errorf("decryption failed! Expected: %s, Got: %s", payload, out.String())
	}

	fmt.Printf("Successfully encrypted and decrypted %d bytes\n", nw)
}
