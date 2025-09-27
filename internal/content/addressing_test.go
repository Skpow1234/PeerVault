package content

import (
	"bytes"
	"encoding/base32"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContentAddresser(t *testing.T) {
	ca := NewContentAddresser()
	assert.NotNil(t, ca)
	assert.NotNil(t, ca.hashers)
}

func TestContentAddresser_GenerateCID(t *testing.T) {
	ca := NewContentAddresser()
	
	tests := []struct {
		name   string
		data   []byte
		codec  string
		hasErr bool
	}{
		{
			name:   "valid data with raw codec",
			data:   []byte("hello world"),
			codec:  "raw",
			hasErr: false,
		},
		{
			name:   "valid data with json codec",
			data:   []byte(`{"key": "value"}`),
			codec:  "json",
			hasErr: false,
		},
		{
			name:   "empty data",
			data:   []byte(""),
			codec:  "raw",
			hasErr: false,
		},
		{
			name:   "large data",
			data:   bytes.Repeat([]byte("a"), 1024*1024), // 1MB
			codec:  "raw",
			hasErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cid, err := ca.GenerateCID(tt.data, tt.codec)
			
			if tt.hasErr {
				assert.Error(t, err)
				assert.Nil(t, cid)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cid)
				assert.Equal(t, 1, cid.Version)
				assert.Equal(t, tt.codec, cid.Codec)
				assert.Equal(t, "sha2-256", cid.Algorithm)
				assert.NotEmpty(t, cid.Hash)
			}
		})
	}
}

func TestContentAddresser_GenerateContentID(t *testing.T) {
	ca := NewContentAddresser()
	
	tests := []struct {
		name   string
		data   []byte
		hasErr bool
	}{
		{
			name:   "valid data",
			data:   []byte("hello world"),
			hasErr: false,
		},
		{
			name:   "empty data",
			data:   []byte(""),
			hasErr: false,
		},
		{
			name:   "large data",
			data:   bytes.Repeat([]byte("a"), 1024*1024), // 1MB
			hasErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentID, err := ca.GenerateContentID(tt.data)
			
			if tt.hasErr {
				assert.Error(t, err)
				assert.Nil(t, contentID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, contentID)
				assert.Equal(t, "sha256", contentID.Algorithm)
				assert.Equal(t, int64(len(tt.data)), contentID.Size)
				assert.NotEmpty(t, contentID.Hash)
				assert.Len(t, contentID.Hash, 64) // SHA-256 hex string length
			}
		})
	}
}

func TestContentAddresser_GenerateContentIDFromReader(t *testing.T) {
	ca := NewContentAddresser()
	
	tests := []struct {
		name   string
		data   []byte
		hasErr bool
	}{
		{
			name:   "valid data",
			data:   []byte("hello world"),
			hasErr: false,
		},
		{
			name:   "empty data",
			data:   []byte(""),
			hasErr: false,
		},
		{
			name:   "large data",
			data:   bytes.Repeat([]byte("a"), 1024*1024), // 1MB
			hasErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			contentID, err := ca.GenerateContentIDFromReader(reader)
			
			if tt.hasErr {
				assert.Error(t, err)
				assert.Nil(t, contentID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, contentID)
				assert.Equal(t, "sha256", contentID.Algorithm)
				assert.Equal(t, int64(len(tt.data)), contentID.Size)
				assert.NotEmpty(t, contentID.Hash)
				assert.Len(t, contentID.Hash, 64) // SHA-256 hex string length
			}
		})
	}
}

func TestContentAddresser_GenerateContentIDFromReader_Error(t *testing.T) {
	ca := NewContentAddresser()
	
	// Create a reader that will return an error
	errorReader := &errorReader{}
	
	contentID, err := ca.GenerateContentIDFromReader(errorReader)
	assert.Error(t, err)
	assert.Nil(t, contentID)
	assert.Contains(t, err.Error(), "failed to read data")
}

func TestContentAddresser_ValidateCID(t *testing.T) {
	ca := NewContentAddresser()
	
	tests := []struct {
		name    string
		cidStr  string
		hasErr  bool
		contains string
	}{
		{
			name:    "valid CID v0",
			cidStr:  "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
			hasErr:  false,
		},
		{
			name:    "valid CID v1",
			cidStr:  "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			hasErr:  false,
		},
		{
			name:     "empty CID",
			cidStr:   "",
			hasErr:   true,
			contains: "CID cannot be empty",
		},
		{
			name:     "invalid CID format",
			cidStr:   "invalid-cid",
			hasErr:   true,
			contains: "invalid CID format",
		},
		{
			name:     "CID without proper prefix",
			cidStr:   "Zm9vYmFy", // base64 encoded "foobar"
			hasErr:   true,
			contains: "invalid CID format",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ca.ValidateCID(tt.cidStr)
			
			if tt.hasErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.contains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentAddresser_ParseCID(t *testing.T) {
	ca := NewContentAddresser()
	
	tests := []struct {
		name   string
		cidStr string
		hasErr bool
	}{
		{
			name:   "valid CID v0",
			cidStr: "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
			hasErr: false,
		},
		{
			name:   "valid CID v1",
			cidStr: "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			hasErr: false,
		},
		{
			name:   "invalid CID",
			cidStr: "invalid-cid",
			hasErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cid, err := ca.ParseCID(tt.cidStr)
			
			if tt.hasErr {
				assert.Error(t, err)
				assert.Nil(t, cid)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cid)
				assert.Equal(t, 1, cid.Version)
				assert.Equal(t, "raw", cid.Codec)
				assert.Equal(t, "sha2-256", cid.Algorithm)
				assert.Equal(t, tt.cidStr, cid.Hash)
			}
		})
	}
}

func TestContentAddresser_ContentIDToCID(t *testing.T) {
	ca := NewContentAddresser()
	
	contentID := &ContentID{
		Hash:      "abcd1234",
		Algorithm: "sha256",
		Size:      100,
	}
	
	tests := []struct {
		name  string
		codec string
	}{
		{
			name:  "raw codec",
			codec: "raw",
		},
		{
			name:  "json codec",
			codec: "json",
		},
		{
			name:  "protobuf codec",
			codec: "protobuf",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cid, err := ca.ContentIDToCID(contentID, tt.codec)
			
			assert.NoError(t, err)
			assert.NotNil(t, cid)
			assert.Equal(t, 1, cid.Version)
			assert.Equal(t, tt.codec, cid.Codec)
			assert.Equal(t, contentID.Hash, cid.Hash)
			assert.Equal(t, contentID.Algorithm, cid.Algorithm)
		})
	}
}

func TestContentAddresser_CIDToContentID(t *testing.T) {
	ca := NewContentAddresser()
	
	cid := &CID{
		Version:   1,
		Codec:     "raw",
		Hash:      "abcd1234",
		Algorithm: "sha2-256",
	}
	
	size := int64(100)
	
	contentID, err := ca.CIDToContentID(cid, size)
	
	assert.NoError(t, err)
	assert.NotNil(t, contentID)
	assert.Equal(t, cid.Hash, contentID.Hash)
	assert.Equal(t, cid.Algorithm, contentID.Algorithm)
	assert.Equal(t, size, contentID.Size)
}

func TestContentAddresser_GenerateBase32Hash(t *testing.T) {
	ca := NewContentAddresser()
	
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "valid data",
			data: []byte("hello world"),
		},
		{
			name: "empty data",
			data: []byte(""),
		},
		{
			name: "large data",
			data: bytes.Repeat([]byte("a"), 1024*1024), // 1MB
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := ca.GenerateBase32Hash(tt.data)
			
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.Equal(t, strings.ToLower(hash), hash) // Should be lowercase
			
			// Base32 encoding should be valid
			_, err = base32.StdEncoding.DecodeString(strings.ToUpper(hash))
			assert.NoError(t, err)
		})
	}
}

func TestContentAddresser_VerifyContent(t *testing.T) {
	ca := NewContentAddresser()
	
	data := []byte("hello world")
	expectedID, err := ca.GenerateContentID(data)
	require.NoError(t, err)
	
	tests := []struct {
		name        string
		data        []byte
		expectedID  *ContentID
		shouldMatch bool
		hasErr      bool
	}{
		{
			name:        "matching content",
			data:        data,
			expectedID:  expectedID,
			shouldMatch: true,
			hasErr:      false,
		},
		{
			name:        "non-matching content",
			data:        []byte("different content"),
			expectedID:  expectedID,
			shouldMatch: false,
			hasErr:      false,
		},
		{
			name:        "empty content",
			data:        []byte(""),
			expectedID:  expectedID,
			shouldMatch: false,
			hasErr:      false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := ca.VerifyContent(tt.data, tt.expectedID)
			
			if tt.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.shouldMatch, matches)
			}
		})
	}
}

func TestContentAddresser_GetContentPath(t *testing.T) {
	ca := NewContentAddresser()
	
	tests := []struct {
		name      string
		contentID *ContentID
		expected  string
	}{
		{
			name: "normal hash",
			contentID: &ContentID{
				Hash: "abcdef1234567890",
			},
			expected: "ab/cdef1234567890",
		},
		{
			name: "short hash",
			contentID: &ContentID{
				Hash: "ab",
			},
			expected: "ab/",
		},
		{
			name: "single character hash",
			contentID: &ContentID{
				Hash: "a",
			},
			expected: "a",
		},
		{
			name: "empty hash",
			contentID: &ContentID{
				Hash: "",
			},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := ca.GetContentPath(tt.contentID)
			assert.Equal(t, tt.expected, path)
		})
	}
}

func TestContentAddresser_GetCIDPath(t *testing.T) {
	ca := NewContentAddresser()
	
	tests := []struct {
		name     string
		cid      *CID
		expected string
	}{
		{
			name: "normal hash",
			cid: &CID{
				Hash: "abcdef1234567890",
			},
			expected: "ab/cdef1234567890",
		},
		{
			name: "short hash",
			cid: &CID{
				Hash: "ab",
			},
			expected: "ab/",
		},
		{
			name: "single character hash",
			cid: &CID{
				Hash: "a",
			},
			expected: "a",
		},
		{
			name: "empty hash",
			cid: &CID{
				Hash: "",
			},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := ca.GetCIDPath(tt.cid)
			assert.Equal(t, tt.expected, path)
		})
	}
}

func TestContentAddresser_DeterministicHashing(t *testing.T) {
	ca := NewContentAddresser()
	
	data := []byte("hello world")
	
	// Generate multiple content IDs for the same data
	id1, err1 := ca.GenerateContentID(data)
	id2, err2 := ca.GenerateContentID(data)
	
	require.NoError(t, err1)
	require.NoError(t, err2)
	
	// Should be identical
	assert.Equal(t, id1.Hash, id2.Hash)
	assert.Equal(t, id1.Algorithm, id2.Algorithm)
	assert.Equal(t, id1.Size, id2.Size)
	
	// Test CID generation
	cid1, err1 := ca.GenerateCID(data, "raw")
	cid2, err2 := ca.GenerateCID(data, "raw")
	
	require.NoError(t, err1)
	require.NoError(t, err2)
	
	// Should be identical
	assert.Equal(t, cid1.Hash, cid2.Hash)
	assert.Equal(t, cid1.Codec, cid2.Codec)
	assert.Equal(t, cid1.Algorithm, cid2.Algorithm)
}

func TestContentAddresser_DifferentDataDifferentHashes(t *testing.T) {
	ca := NewContentAddresser()
	
	data1 := []byte("hello world")
	data2 := []byte("hello universe")
	
	id1, err1 := ca.GenerateContentID(data1)
	id2, err2 := ca.GenerateContentID(data2)
	
	require.NoError(t, err1)
	require.NoError(t, err2)
	
	// Should be different
	assert.NotEqual(t, id1.Hash, id2.Hash)
	assert.Equal(t, id1.Algorithm, id2.Algorithm) // Same algorithm
	assert.NotEqual(t, id1.Size, id2.Size)        // Different sizes
}

// errorReader is a helper for testing error conditions
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
