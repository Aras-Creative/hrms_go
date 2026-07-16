package entity

import (
	"time"

	"github.com/google/uuid"
)

type SpecialClause struct {
	Title       string
	Description string
}

type Block struct {
	Type     string
	Title    *string
	Content  *string
	Position *string
}

type ContractTemplateData struct {
	DesignationID    *string
	WorkingPatternID *string
	JobDuties        []string
	InventoryItems   []string
	SpecialClauses   []SpecialClause
	EmployeeFields   []string
}

type ContractTemplatePartials struct {
	Blocks []Block
}

type ContractTemplate struct {
	ID           string
	Name         string
	ContractType ContractType
	Description  string
	IsActive     bool
	Data         ContractTemplateData
	Templates    ContractTemplatePartials
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Update applies the given fields to the template and touches UpdatedAt.
func (t *ContractTemplate) Update(
	name string,
	contractType ContractType,
	description string,
	isActive bool,
	data ContractTemplateData,
	templates ContractTemplatePartials,
) {
	t.Name = name
	t.ContractType = contractType
	t.Description = description
	t.IsActive = isActive
	t.Data = data
	t.Templates = templates
	t.UpdatedAt = time.Now()
}

func NewContractTemplate(
	name string,
	contractType ContractType,
	description string,
	data ContractTemplateData,
	templates ContractTemplatePartials,
) *ContractTemplate {
	now := time.Now()
	return &ContractTemplate{
		ID:           uuid.New().String(),
		Name:         name,
		ContractType: contractType,
		Description:  description,
		IsActive:     true,
		Data:         data,
		Templates:    templates,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func ReconstituteContractTemplate(
	id string,
	name string,
	contractType ContractType,
	description string,
	isActive bool,
	data ContractTemplateData,
	templates ContractTemplatePartials,
	createdAt time.Time,
	updatedAt time.Time,
) *ContractTemplate {
	return &ContractTemplate{
		ID:           id,
		Name:         name,
		ContractType: contractType,
		Description:  description,
		IsActive:     isActive,
		Data:         data,
		Templates:    templates,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
