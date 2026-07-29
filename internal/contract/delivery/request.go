package delivery

import (
	"time"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/models"
	errors "hrms/internal/pkg/apperror"
)

// ---- Template Requests ----

type SpecialClauseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type BlockRequest struct {
	Type     string  `json:"type"`
	Title    *string `json:"title,omitempty"`
	Content  *string `json:"content,omitempty"`
	Position *string `json:"position,omitempty"`
}

type TemplateDataRequest struct {
	DesignationID    *string              `json:"designation_id" validate:"omitempty,uuid"`
	WorkingPatternID *string              `json:"working_pattern_id" validate:"omitempty,uuid"`
	JobDuties        []string             `json:"job_duties"`
	InventoryItems   []string             `json:"inventory_items"`
	SpecialClauses   []SpecialClauseRequest `json:"special_clauses"`
	EmployeeFields   []string             `json:"employee_fields"`
}

type TemplatePartialsRequest struct {
	Blocks []BlockRequest `json:"blocks"`
}

type CreateTemplateRequest struct {
	Name         string                 `json:"name" validate:"required,min=1,max=255"`
	ContractType string                 `json:"contract_type" validate:"required,oneof=PKWT PKWTT"`
	Description  string                 `json:"description"`
	Data         TemplateDataRequest    `json:"data"`
	Templates    TemplatePartialsRequest `json:"templates"`
}

type CopyTemplateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type UpdateTemplateRequest struct {
	Name         string                 `json:"name" validate:"required,min=1,max=255"`
	ContractType string                 `json:"contract_type" validate:"required,oneof=PKWT PKWTT"`
	Description  string                 `json:"description"`
	IsActive     *bool                  `json:"is_active"`
	Data         TemplateDataRequest    `json:"data"`
	Templates    TemplatePartialsRequest `json:"templates"`
}

// ---- Contract Requests ----

type ContractItemRequest struct {
	EmployeeID string `json:"employee_id" validate:"required,uuid"`
	Number     string `json:"number" validate:"required"`
}

type CreateContractRequest struct {
	TemplateID       string               `json:"template_id" validate:"required,uuid"`
	Contracts        []ContractItemRequest `json:"contracts" validate:"required,min=1,dive"`
	StartDate        string               `json:"start_date" validate:"required"`
	EndDate          *string              `json:"end_date,omitempty"`
	DesignationID    *string              `json:"designation_id,omitempty" validate:"omitempty,uuid"`
	WorkingPatternID *string              `json:"working_pattern_id,omitempty" validate:"omitempty,uuid"`
	JobDuties        []string             `json:"job_duties,omitempty"`
	InventoryItems   []string             `json:"inventory_items,omitempty"`
}

type SignContractRequest struct {
	ContractIDs     []string `json:"contract_ids" validate:"required,min=1"`
	Party           string   `json:"party" validate:"required,oneof=first second"`
	SignedBy        string   `json:"signed_by" validate:"required"`
	SignedByName    string   `json:"signed_by_name" validate:"required"`
	SignedByTitle   string   `json:"signed_by_title"`
	Place           string   `json:"place" validate:"required"`
	SignatureBase64 string   `json:"signature_base64" validate:"required"`
}

type CheckActiveContractRequest struct {
	EmployeeIDs []string `json:"employee_ids" validate:"required,min=1,dive,uuid"`
}

type TerminateContractRequest struct {
	TerminationDate string `json:"termination_date"`
}

func (r *CreateContractRequest) ToInput() (*models.CreateContractInput, error) {
	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return nil, errors.NewInvalidInput("invalid start_date, expected YYYY-MM-DD")
	}

	var endDate *time.Time
	if r.EndDate != nil {
		t, err := time.Parse("2006-01-02", *r.EndDate)
		if err != nil {
			return nil, errors.NewInvalidInput("invalid end_date, expected YYYY-MM-DD")
		}
		endDate = &t
		if !endDate.After(startDate) {
			return nil, errors.NewInvalidInput("end_date must be after start_date")
		}
	}

	contracts := make([]models.ContractItemInput, len(r.Contracts))
	for i, c := range r.Contracts {
		contracts[i] = models.ContractItemInput{
			EmployeeID: c.EmployeeID,
			Number:     c.Number,
		}
	}

	return &models.CreateContractInput{
		TemplateID:       r.TemplateID,
		Contracts:        contracts,
		StartDate:        startDate,
		EndDate:          endDate,
		DesignationID:    r.DesignationID,
		WorkingPatternID: r.WorkingPatternID,
		JobDuties:        r.JobDuties,
		InventoryItems:   r.InventoryItems,
	}, nil
}

func (r *TerminateContractRequest) GetTerminationDate() (time.Time, error) {
	if r.TerminationDate == "" {
		return time.Now(), nil
	}
	parsed, err := time.Parse("2006-01-02", r.TerminationDate)
	if err != nil {
		return time.Time{}, errors.NewInvalidInput("invalid termination_date, expected YYYY-MM-DD")
	}
	return parsed, nil
}

// ---- Update Draft Contract Request ----

type UpdateDraftContractRequest struct {
	Number           string                 `json:"number"`
	StartDate        string                 `json:"start_date" validate:"required"`
	EndDate          *string                `json:"end_date,omitempty"`
	Salary           string                 `json:"salary"`
	DesignationID    *string                `json:"designation_id,omitempty"`
	DesignationTitle string                 `json:"designation_title"`
	Data             TemplateDataRequest    `json:"data"`
	Templates        TemplatePartialsRequest `json:"templates"`
}

func (r *UpdateDraftContractRequest) ToInput(id string) (*models.UpdateContractInput, error) {
	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return nil, errors.NewInvalidInput("invalid start_date, expected YYYY-MM-DD")
	}

	var endDate *time.Time
	if r.EndDate != nil {
		t, err := time.Parse("2006-01-02", *r.EndDate)
		if err != nil {
			return nil, errors.NewInvalidInput("invalid end_date, expected YYYY-MM-DD")
		}
		endDate = &t
	}

	return &models.UpdateContractInput{
		ID:               id,
		Number:           r.Number,
		StartDate:        startDate,
		EndDate:          endDate,
		Salary:           r.Salary,
		DesignationID:    r.DesignationID,
		DesignationTitle: r.DesignationTitle,
		Data:             toEntityData(r.Data),
		Templates:        toEntityPartials(r.Templates),
	}, nil
}

func toEntityData(d TemplateDataRequest) entity.ContractTemplateData {
	clauses := make([]entity.SpecialClause, len(d.SpecialClauses))
	for i, c := range d.SpecialClauses {
		clauses[i] = entity.SpecialClause{Title: c.Title, Description: c.Description}
	}
	return entity.ContractTemplateData{
		DesignationID: d.DesignationID, WorkingPatternID: d.WorkingPatternID,
		JobDuties: d.JobDuties, InventoryItems: d.InventoryItems,
		SpecialClauses: clauses, EmployeeFields: d.EmployeeFields,
	}
}

func toEntityPartials(t TemplatePartialsRequest) entity.ContractTemplatePartials {
	blocks := make([]entity.Block, len(t.Blocks))
	for i, b := range t.Blocks {
		blocks[i] = entity.Block{Type: b.Type, Title: b.Title, Content: b.Content, Position: b.Position}
	}
	return entity.ContractTemplatePartials{Blocks: blocks}
}
