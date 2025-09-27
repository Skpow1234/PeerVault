package mapper

import (
	"testing"

	"github.com/Skpow1234/Peervault/internal/domain"
	"github.com/Skpow1234/Peervault/internal/dto"
	"github.com/stretchr/testify/assert"
)

func TestToDTO(t *testing.T) {
	tests := []struct {
		name     string
		meta     domain.FileMetadata
		expected dto.StoreFile
	}{
		{
			name: "basic metadata",
			meta: domain.FileMetadata{
				ID:        domain.NodeID("test-id-123"),
				HashedKey: "hashed-key-456",
				Size:      1024,
			},
			expected: dto.StoreFile{
				ID:   "test-id-123",
				Key:  "hashed-key-456",
				Size: 1024,
			},
		},
		{
			name: "empty metadata",
			meta: domain.FileMetadata{
				ID:        domain.NodeID(""),
				HashedKey: "",
				Size:      0,
			},
			expected: dto.StoreFile{
				ID:   "",
				Key:  "",
				Size: 0,
			},
		},
		{
			name: "large size",
			meta: domain.FileMetadata{
				ID:        domain.NodeID("large-file-id"),
				HashedKey: "large-file-hash",
				Size:      1024 * 1024 * 1024, // 1GB
			},
			expected: dto.StoreFile{
				ID:   "large-file-id",
				Key:  "large-file-hash",
				Size: 1024 * 1024 * 1024,
			},
		},
		{
			name: "special characters in ID",
			meta: domain.FileMetadata{
				ID:        domain.NodeID("test-id-with-special-chars-!@#$%^&*()"),
				HashedKey: "hash-with-special-chars-!@#$%^&*()",
				Size:      512,
			},
			expected: dto.StoreFile{
				ID:   "test-id-with-special-chars-!@#$%^&*()",
				Key:  "hash-with-special-chars-!@#$%^&*()",
				Size: 512,
			},
		},
		{
			name: "unicode characters",
			meta: domain.FileMetadata{
				ID:        domain.NodeID("test-id-中文-日本語-한국어"),
				HashedKey: "hash-中文-日本語-한국어",
				Size:      256,
			},
			expected: dto.StoreFile{
				ID:   "test-id-中文-日本語-한국어",
				Key:  "hash-中文-日本語-한국어",
				Size: 256,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToDTO(tt.meta)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToDomainGet(t *testing.T) {
	tests := []struct {
		name     string
		get      dto.GetFile
		expected domain.FileMetadata
	}{
		{
			name: "basic get file",
			get: dto.GetFile{
				ID:  "test-id-123",
				Key: "hashed-key-456",
			},
			expected: domain.FileMetadata{
				ID:        domain.NodeID("test-id-123"),
				HashedKey: "hashed-key-456",
			},
		},
		{
			name: "empty get file",
			get: dto.GetFile{
				ID:  "",
				Key: "",
			},
			expected: domain.FileMetadata{
				ID:        domain.NodeID(""),
				HashedKey: "",
			},
		},
		{
			name: "long ID and key",
			get: dto.GetFile{
				ID:  "very-long-id-that-might-be-used-in-real-world-scenarios-with-many-characters",
				Key: "very-long-hash-that-might-be-used-in-real-world-scenarios-with-many-characters",
			},
			expected: domain.FileMetadata{
				ID:        domain.NodeID("very-long-id-that-might-be-used-in-real-world-scenarios-with-many-characters"),
				HashedKey: "very-long-hash-that-might-be-used-in-real-world-scenarios-with-many-characters",
			},
		},
		{
			name: "special characters",
			get: dto.GetFile{
				ID:  "test-id-with-special-chars-!@#$%^&*()",
				Key: "hash-with-special-chars-!@#$%^&*()",
			},
			expected: domain.FileMetadata{
				ID:        domain.NodeID("test-id-with-special-chars-!@#$%^&*()"),
				HashedKey: "hash-with-special-chars-!@#$%^&*()",
			},
		},
		{
			name: "unicode characters",
			get: dto.GetFile{
				ID:  "test-id-中文-日本語-한국어",
				Key: "hash-中文-日本語-한국어",
			},
			expected: domain.FileMetadata{
				ID:        domain.NodeID("test-id-中文-日本語-한국어"),
				HashedKey: "hash-中文-日本語-한국어",
			},
		},
		{
			name: "numeric ID",
			get: dto.GetFile{
				ID:  "123456789",
				Key: "numeric-hash-987654321",
			},
			expected: domain.FileMetadata{
				ID:        domain.NodeID("123456789"),
				HashedKey: "numeric-hash-987654321",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToDomainGet(tt.get)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToDTO_ToDomainGet_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		meta domain.FileMetadata
	}{
		{
			name: "basic round trip",
			meta: domain.FileMetadata{
				ID:        domain.NodeID("test-id-123"),
				HashedKey: "hashed-key-456",
				Size:      1024,
			},
		},
		{
			name: "empty round trip",
			meta: domain.FileMetadata{
				ID:        domain.NodeID(""),
				HashedKey: "",
				Size:      0,
			},
		},
		{
			name: "large size round trip",
			meta: domain.FileMetadata{
				ID:        domain.NodeID("large-file-id"),
				HashedKey: "large-file-hash",
				Size:      1024 * 1024 * 1024,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to DTO
			dtoResult := ToDTO(tt.meta)

			// Convert back to domain (note: Size is lost in this conversion)
			getFile := dto.GetFile{
				ID:  dtoResult.ID,
				Key: dtoResult.Key,
			}
			domainResult := ToDomainGet(getFile)

			// Check that ID and Key are preserved
			assert.Equal(t, tt.meta.ID, domainResult.ID)
			assert.Equal(t, tt.meta.HashedKey, domainResult.HashedKey)
		})
	}
}

func TestToDomainGet_ToDTO_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		get  dto.GetFile
	}{
		{
			name: "basic round trip",
			get: dto.GetFile{
				ID:  "test-id-123",
				Key: "hashed-key-456",
			},
		},
		{
			name: "empty round trip",
			get: dto.GetFile{
				ID:  "",
				Key: "",
			},
		},
		{
			name: "special characters round trip",
			get: dto.GetFile{
				ID:  "test-id-with-special-chars-!@#$%^&*()",
				Key: "hash-with-special-chars-!@#$%^&*()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to domain
			domain := ToDomainGet(tt.get)

			// Convert back to DTO
			dto := ToDTO(domain)

			// Check that ID and Key are preserved
			assert.Equal(t, tt.get.ID, dto.ID)
			assert.Equal(t, tt.get.Key, dto.Key)
			// Size should be 0 since it wasn't in the original GetFile
			assert.Equal(t, int64(0), dto.Size)
		})
	}
}

func TestMapper_EdgeCases(t *testing.T) {
	// Test with very long strings
	longID := string(make([]byte, 10000))
	for i := range longID {
		longID = longID[:i] + "a" + longID[i+1:]
	}

	longKey := string(make([]byte, 10000))
	for i := range longKey {
		longKey = longKey[:i] + "b" + longKey[i+1:]
	}

	meta := domain.FileMetadata{
		ID:        domain.NodeID(longID),
		HashedKey: longKey,
		Size:      999999999,
	}

	dtoResult := ToDTO(meta)
	assert.Equal(t, longID, dtoResult.ID)
	assert.Equal(t, longKey, dtoResult.Key)
	assert.Equal(t, int64(999999999), dtoResult.Size)

	// Test round trip with long strings
	getFile := dto.GetFile{
		ID:  dtoResult.ID,
		Key: dtoResult.Key,
	}
	domainResult := ToDomainGet(getFile)
	assert.Equal(t, domain.NodeID(longID), domainResult.ID)
	assert.Equal(t, longKey, domainResult.HashedKey)
}

func TestMapper_TypeConversions(t *testing.T) {
	// Test that type conversions work correctly
	meta := domain.FileMetadata{
		ID:        domain.NodeID("test-id"),
		HashedKey: "test-key",
		Size:      12345,
	}

	dtoResult := ToDTO(meta)

	// Verify that NodeID is converted to string
	assert.IsType(t, "", dtoResult.ID)
	assert.Equal(t, "test-id", dtoResult.ID)

	// Verify that string is converted to NodeID
	getFile := dto.GetFile{
		ID:  dtoResult.ID,
		Key: dtoResult.Key,
	}
	domainResult := ToDomainGet(getFile)
	assert.IsType(t, domain.NodeID(""), domainResult.ID)
	assert.Equal(t, domain.NodeID("test-id"), domainResult.ID)
}
