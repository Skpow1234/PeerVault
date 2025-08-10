package storage

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/anthdm/foreverstore/internal/crypto"
)

func TestPathTransformFunc(t *testing.T) {
    key := "momsbestpicture"
    pathKey := CASPathTransformFunc(key)
    expectedFilename := "6804429f74181a63c50c3d81d733a12f14a353ff"
    expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
    assert.Equal(t, expectedPathName, pathKey.PathName)
    assert.Equal(t, expectedFilename, pathKey.Filename)
}

func TestStore(t *testing.T) {
    s := newStore()
    id := crypto.GenerateID()
    defer teardown(t, s)
    for i := 0; i < 50; i++ {
        key := fmt.Sprintf("foo_%d", i)
        data := []byte("some jpg bytes")
        if _, err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
            t.Error(err)
        }
        if ok := s.Has(id, key); !ok {
            t.Errorf("expected to have key %s", key)
        }
        _, r, err := s.Read(id, key)
        if err != nil {
            t.Error(err)
        }
        b, _ := ioutil.ReadAll(r)
        _ = r.Close()
        if string(b) != string(data) {
            t.Errorf("want %s have %s", data, b)
        }
        if err := s.Delete(id, key); err != nil {
            t.Error(err)
        }
        if ok := s.Has(id, key); ok {
            t.Errorf("expected to NOT have key %s", key)
        }
    }
}

func newStore() *Store {
    opts := StoreOpts{PathTransformFunc: CASPathTransformFunc}
    return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
    if err := s.Clear(); err != nil {
        t.Error(err)
    }
}


