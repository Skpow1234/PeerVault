package content

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"strings"

	"github.com/multiformats/go-multihash"
)

// ContentID represents a content-addressed identifier
type ContentID struct {
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
	Size      int64  `json:"size"`
}

// CID represents a Content Identifier compatible with IPFS
type CID struct {
	Version   int    `json:"version"`
	Codec     string `json:"codec"`
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
}

// ContentAddresser provides content addressing functionality
type ContentAddresser struct {
	hashers map[string]hash.Hash
}

// NewContentAddresser creates a new content addresser
func NewContentAddresser() *ContentAddresser {
	return &ContentAddresser{
		hashers: make(map[string]hash.Hash),
	}
}

// GenerateCID generates a Content Identifier for the given data
func (ca *ContentAddresser) GenerateCID(data []byte, codec string) (*CID, error) {
	// Use SHA-256 for content addressing
	hasher := sha256.New()
	hasher.Write(data)
	hashBytes := hasher.Sum(nil)

	// Create multihash
	mh, err := multihash.Encode(hashBytes, multihash.SHA2_256)
	if err != nil {
		return nil, fmt.Errorf("failed to encode multihash: %w", err)
	}

	// Generate CID
	cid := &CID{
		Version:   1,
		Codec:     codec,
		Hash:      hex.EncodeToString(mh),
		Algorithm: "sha2-256",
	}

	return cid, nil
}

// GenerateContentID generates a content ID for the given data
func (ca *ContentAddresser) GenerateContentID(data []byte) (*ContentID, error) {
	hasher := sha256.New()
	hasher.Write(data)
	hashBytes := hasher.Sum(nil)

	return &ContentID{
		Hash:      hex.EncodeToString(hashBytes),
		Algorithm: "sha256",
		Size:      int64(len(data)),
	}, nil
}

// GenerateContentIDFromReader generates a content ID from a reader
func (ca *ContentAddresser) GenerateContentIDFromReader(reader io.Reader) (*ContentID, error) {
	hasher := sha256.New()
	size, err := io.Copy(hasher, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	hashBytes := hasher.Sum(nil)

	return &ContentID{
		Hash:      hex.EncodeToString(hashBytes),
		Algorithm: "sha256",
		Size:      size,
	}, nil
}

// ValidateCID validates a CID format
func (ca *ContentAddresser) ValidateCID(cidStr string) error {
	if cidStr == "" {
		return fmt.Errorf("CID cannot be empty")
	}

	// Basic CID format validation
	if !strings.HasPrefix(cidStr, "Qm") && !strings.HasPrefix(cidStr, "bafy") {
		return fmt.Errorf("invalid CID format")
	}

	return nil
}

// ParseCID parses a CID string into a CID struct
func (ca *ContentAddresser) ParseCID(cidStr string) (*CID, error) {
	if err := ca.ValidateCID(cidStr); err != nil {
		return nil, err
	}

	// Parse CID components
	// This is a simplified parser - in a real implementation, you'd use a proper CID library
	cid := &CID{
		Version:   1,
		Codec:     "raw", // Default codec
		Hash:      cidStr,
		Algorithm: "sha2-256",
	}

	return cid, nil
}

// ContentIDToCID converts a ContentID to a CID
func (ca *ContentAddresser) ContentIDToCID(contentID *ContentID, codec string) (*CID, error) {
	return &CID{
		Version:   1,
		Codec:     codec,
		Hash:      contentID.Hash,
		Algorithm: contentID.Algorithm,
	}, nil
}

// CIDToContentID converts a CID to a ContentID
func (ca *ContentAddresser) CIDToContentID(cid *CID, size int64) (*ContentID, error) {
	return &ContentID{
		Hash:      cid.Hash,
		Algorithm: cid.Algorithm,
		Size:      size,
	}, nil
}

// GenerateBase32Hash generates a base32 encoded hash for the given data
func (ca *ContentAddresser) GenerateBase32Hash(data []byte) (string, error) {
	hasher := sha256.New()
	hasher.Write(data)
	hashBytes := hasher.Sum(nil)

	// Encode as base32 (IPFS style)
	encoded := base32.StdEncoding.EncodeToString(hashBytes)
	return strings.ToLower(encoded), nil
}

// VerifyContent verifies that the given data matches the expected content ID
func (ca *ContentAddresser) VerifyContent(data []byte, expectedID *ContentID) (bool, error) {
	actualID, err := ca.GenerateContentID(data)
	if err != nil {
		return false, err
	}

	return actualID.Hash == expectedID.Hash, nil
}

// GetContentPath generates a content-addressed path for storage
func (ca *ContentAddresser) GetContentPath(contentID *ContentID) string {
	// Use first 2 characters as directory, rest as filename
	if len(contentID.Hash) < 2 {
		return contentID.Hash
	}

	return fmt.Sprintf("%s/%s", contentID.Hash[:2], contentID.Hash[2:])
}

// GetCIDPath generates a CID-based path for storage
func (ca *ContentAddresser) GetCIDPath(cid *CID) string {
	// Use first 2 characters as directory, rest as filename
	if len(cid.Hash) < 2 {
		return cid.Hash
	}

	return fmt.Sprintf("%s/%s", cid.Hash[:2], cid.Hash[2:])
}
