package models

import (
	"hrms/internal/contract/entity"
)

type CreateTemplateInput struct {
	Name         string
	ContractType entity.ContractType
	Description  string
	Data         entity.ContractTemplateData
	Templates    entity.ContractTemplatePartials
}

type UpdateTemplateInput struct {
	ID           string
	Name         string
	ContractType entity.ContractType
	Description  string
	IsActive     bool
	Data         entity.ContractTemplateData
	Templates    entity.ContractTemplatePartials
}

type ListTemplateInput struct {
	Page         int
	PerPage      int
	SearchName   string
	ContractType string
	IsActive     *bool
}

type ListTemplateResult struct {
	Entities []*entity.ContractTemplate
	Total    int64
}
