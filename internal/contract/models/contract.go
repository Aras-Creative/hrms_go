package models

import (
	"hrms/internal/contract/entity"
)

type CreateTemplateInput struct {
	Name         string
	ContractType string
	Description  string
	Data         entity.ContractTemplateData
	Templates    entity.ContractTemplatePartials
}

type UpdateTemplateInput struct {
	ID           string
	Name         string
	ContractType string
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
