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

	fmt.Println(len(payload))
	fmt.Println(len(dst.String()))

	out := new(bytes.Buffer)
    nw, err := internalcrypto.CopyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	if nw != 16+len(payload) {
		t.Fail()
	}

	if out.String() != payload {
		t.Errorf("decryption failed!!!")
	}
}
