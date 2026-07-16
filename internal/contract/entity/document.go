package entity

import (
	"time"

	"github.com/google/uuid"
)

type ContractDocument struct {
	ID          string
	ContractID  string
	DocumentID  string
	ContentHash string
	CreatedAt   time.Time
}

func NewContractDocument(contractID, documentID, contentHash string) *ContractDocument {
	return &ContractDocument{
		ID:          uuid.New().String(),
		ContractID:  contractID,
		DocumentID:  documentID,
		ContentHash: contentHash,
		CreatedAt:   time.Now(),
	}
}

func ReconstituteContractDocument(id, contractID, documentID, contentHash string, createdAt time.Time) *ContractDocument {
	return &ContractDocument{
		ID:          id,
		ContractID:  contractID,
		DocumentID:  documentID,
		ContentHash: contentHash,
		CreatedAt:   createdAt,
	}
}