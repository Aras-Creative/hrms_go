package models

import (
	"hrms/internal/contract/entity"
	"time"
)

type CreateContractInput struct {
	TemplateID  string
	EmployeeIDs []string
	StartDate   time.Time
	EndDate     *time.Time
}

type BulkCreateContractResult struct {
	Contracts []*entity.Contract
}

type BulkSignContractInput struct {
	ContractIDs     []string
	Party           string // "first" or "second"
	SignedBy        string
	SignedByName    string
	SignedByTitle   string
	Place           string
	SignatureBase64 string
}

type CheckActiveContractInput struct {
	EmployeeIDs []string
}

type CheckActiveContractItem struct {
	EmployeeID string
	EndDate    *time.Time
}

type CheckActiveContractResult struct {
	Items []CheckActiveContractItem
}

type ListContractInput struct {
	Page          int
	PerPage       int
	Status        string
	ContractType  string
	DesignationID string
	EmployeeID    string
	ExcludeDraft  bool
}

type ListContractResult struct {
	Items []*entity.Contract
	Total int64
}


