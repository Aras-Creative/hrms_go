package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var validBlockTypes = map[string]bool{
	"document_header": true,
	"article":         true,
	"paragraph":       true,
	"first_party":     true,
	"second_party":    true,
	"sign_table":      true,
	"text":            true,
	"raw":             true,
}

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

func validateTemplateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("template name is required")
	}
	return nil
}

func validateBlocks(blocks []Block) error {
	for i, b := range blocks {
		if b.Type != "" && !validBlockTypes[b.Type] {
			return fmt.Errorf("block[%d]: invalid type %q (valid: document_header, article, paragraph, first_party, second_party, sign_table, text)", i, b.Type)
		}
	}
	return nil
}

// Update applies the given fields to the template and touches UpdatedAt.
func (t *ContractTemplate) Update(
	name string,
	contractType ContractType,
	description string,
	isActive bool,
	data ContractTemplateData,
	templates ContractTemplatePartials,
) error {
	if err := validateTemplateName(name); err != nil {
		return err
	}
	if err := validateBlocks(templates.Blocks); err != nil {
		return err
	}
	t.Name = name
	t.ContractType = contractType
	t.Description = description
	t.IsActive = isActive
	t.Data = data
	t.Templates = templates
	t.UpdatedAt = time.Now()
	return nil
}

func NewContractTemplate(
	name string,
	contractType ContractType,
	description string,
	data ContractTemplateData,
	templates ContractTemplatePartials,
) (*ContractTemplate, error) {
	if err := validateTemplateName(name); err != nil {
		return nil, err
	}
	if err := validateBlocks(templates.Blocks); err != nil {
		return nil, err
	}
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
	}, nil
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
