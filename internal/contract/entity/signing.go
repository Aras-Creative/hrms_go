package entity

import (
	"time"

	"github.com/google/uuid"
)

type ContractSigning struct {
	ID              string
	ContractID      string
	Party           string // "first" or "second"
	SignedBy        string
	SignedByName    string
	SignedByTitle   string
	Place           string
	SignatureBase64 string
	SignedAt        time.Time
	CreatedAt       time.Time
}

func NewContractSigning(party, contractID, signedBy, signedByName, signedByTitle, place, signatureBase64 string) *ContractSigning {
	now := time.Now()
	return &ContractSigning{
		ID:              uuid.New().String(),
		ContractID:      contractID,
		Party:           party,
		SignedBy:        signedBy,
		SignedByName:    signedByName,
		SignedByTitle:   signedByTitle,
		Place:           place,
		SignatureBase64: signatureBase64,
		SignedAt:        now,
		CreatedAt:       now,
	}
}

func ReconstituteContractSigning(id, party, contractID, signedBy, signedByName, signedByTitle, place, signatureBase64 string, signedAt, createdAt time.Time) *ContractSigning {
	return &ContractSigning{
		ID:              id,
		ContractID:      contractID,
		Party:           party,
		SignedBy:        signedBy,
		SignedByName:    signedByName,
		SignedByTitle:   signedByTitle,
		Place:           place,
		SignatureBase64: signatureBase64,
		SignedAt:        signedAt,
		CreatedAt:       createdAt,
	}
}