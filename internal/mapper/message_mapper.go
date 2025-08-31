package mapper

import (
	"github.com/Skpow1234/Peervault/internal/domain"
	"github.com/Skpow1234/Peervault/internal/dto"
)

// ToDTO converts domain metadata to a wire/message DTO used for broadcasting.
func ToDTO(meta domain.FileMetadata) dto.StoreFile {
	return dto.StoreFile{
		ID:   string(meta.ID),
		Key:  meta.HashedKey,
		Size: meta.Size,
	}
}

// ToDomainGet converts a GetFile DTO into domain-friendly values.
func ToDomainGet(get dto.GetFile) domain.FileMetadata {
	return domain.FileMetadata{
		ID:        domain.NodeID(get.ID),
		HashedKey: get.Key,
	}
}
