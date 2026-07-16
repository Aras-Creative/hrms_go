package delivery

import (
	"time"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/usecase"
)

// ---- Template Responses ----

type SpecialClauseResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type BlockResponse struct {
	Type     string  `json:"type"`
	Title    *string `json:"title,omitempty"`
	Content  *string `json:"content,omitempty"`
	Position *string `json:"position,omitempty"`
}

type TemplateDataResponse struct {
	DesignationID    *string                `json:"designation_id"`
	WorkingPatternID *string                `json:"working_pattern_id"`
	JobDuties        []string               `json:"job_duties"`
	InventoryItems   []string               `json:"inventory_items"`
	SpecialClauses   []SpecialClauseResponse `json:"special_clauses"`
	EmployeeFields   []string               `json:"employee_fields"`
}

type TemplatePartialsResponse struct {
	Blocks []BlockResponse `json:"blocks"`
}

type TemplateResponse struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	ContractType string                   `json:"contract_type"`
	Description  string                   `json:"description"`
	IsActive     bool                     `json:"is_active"`
	Data         TemplateDataResponse     `json:"data"`
	Templates    TemplatePartialsResponse `json:"templates"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

type TemplateListItemResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ContractType string    `json:"contract_type"`
	Description  string    `json:"description"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListTemplateResponse struct {
	Items []*TemplateListItemResponse `json:"items"`
	Total int64                       `json:"total"`
}

type TemplatePrefillResponse struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	ContractType string               `json:"contract_type"`
	Data         TemplateDataResponse `json:"data"`
}

// ---- Contract Responses ----

type ContractSigningResponse struct {
	ID              string    `json:"id"`
	Party           string    `json:"party"`
	SignedBy        string    `json:"signed_by"`
	SignedByName    string    `json:"signed_by_name"`
	SignedByTitle   string    `json:"signed_by_title"`
	Place           string    `json:"place"`
	SignatureBase64 string    `json:"signature_base64,omitempty"`
	SignedAt        time.Time `json:"signed_at"`
}

// ContractSigningListItemResponse is a redacted signing response for list views.
type ContractSigningListItemResponse struct {
	ID            string    `json:"id"`
	Party         string    `json:"party"`
	SignedByName  string    `json:"signed_by_name"`
	SignedAt      time.Time `json:"signed_at"`
}

type ContractResponse struct {
	ID                string                   `json:"id"`
	TemplateID        string                   `json:"template_id"`
	EmployeeID        string                   `json:"employee_id"`
	Number            string                   `json:"number"`
	StartDate         *time.Time               `json:"start_date"`
	EndDate           *time.Time               `json:"end_date,omitempty"`
	Salary            string                   `json:"salary"`
	DesignationTitle  string                   `json:"designation_title"`
	Status            string                   `json:"status"`
	Data              TemplateDataResponse      `json:"data"`
	Templates         TemplatePartialsResponse `json:"templates"`
	SentAt            *time.Time                `json:"sent_at,omitempty"`
	Signings          []ContractSigningResponse `json:"signings,omitempty"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
}

type ListContractResponse struct {
	Items []*ContractListItemResponse `json:"items"`
	Total int64                       `json:"total"`
}

type ContractListItemEmployeeBrief struct {
	Name            string `json:"name"`
	ProfilePhotoURL string `json:"profile_photo_url"`
}

type ContractListItemResponse struct {
	ID                string                       `json:"id"`
	TemplateID        string                       `json:"template_id"`
	EmployeeID        string                       `json:"employee_id"`
	Employee          *ContractListItemEmployeeBrief `json:"employee"`
	Number            string                       `json:"number"`
	DesignationID     *string                      `json:"designation_id"`
	DesignationTitle  string                       `json:"designation_title"`
	Status            string                       `json:"status"`
	ContractType      string                       `json:"contract_type"`
	StartDate         *time.Time                   `json:"start_date"`
	EndDate           *time.Time                   `json:"end_date,omitempty"`
	Salary            string                       `json:"salary"`
	DocumentID        *string                      `json:"document_id,omitempty"`
	FirstSignedAt     *time.Time                   `json:"first_signed_at,omitempty"`
	SecondSignedAt    *time.Time                   `json:"second_signed_at,omitempty"`
	SentAt            *time.Time                   `json:"sent_at,omitempty"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

type ContractDetailResponse struct {
	ID                string                     `json:"id"`
	TemplateID        string                     `json:"template_id"`
	TemplateName      string                     `json:"template_name"`
	EmployeeID        string                     `json:"employee_id"`
	Number            string                     `json:"number"`
	StartDate         *time.Time                 `json:"start_date"`
	EndDate           *time.Time                 `json:"end_date,omitempty"`
	Salary            string                     `json:"salary"`
	DesignationTitle  string                     `json:"designation_title"`
	Status            string                     `json:"status"`
	ContractType      string                     `json:"contract_type"`
	JobDuties         []string                   `json:"job_duties"`
	Signings          []ContractSigningResponse  `json:"signings,omitempty"`
	SentAt            *time.Time                 `json:"sent_at,omitempty"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

type ActiveContractResponse struct {
	ID                string     `json:"id"`
	Number            string     `json:"number"`
	ContractType      string     `json:"contract_type"`
	DesignationTitle  string     `json:"designation_title"`
	StartDate         *time.Time `json:"start_date"`
	EndDate           *time.Time `json:"end_date"`
}

type CountSoonExpiredResponse struct {
	SoonExpired int64 `json:"soon_expired"`
}

type PendingContractsResponse struct {
	Pending int64 `json:"pending"`
}

type SignContractsResponse struct {
	Signed    int                `json:"signed"`
	Contracts []*ContractResponse `json:"contracts"`
}

type GeneratePDFResponse struct {
	DocumentID  string `json:"document_id"`
	ContentHash string `json:"content_hash"`
}

type EmployeeContractResponse struct {
	ContractID     string     `json:"contract_id"`
	StartDate      *time.Time `json:"start_date"`
	Status         string     `json:"status"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	DesignationID  *string    `json:"designation_id"`
	ContractType   string     `json:"contract_type"`
	JobDuties      []string   `json:"job_duties"`
	InventoryItems []string   `json:"inventory_items"`
}

// ---- Converters ----

func toTemplateResponse(e *entity.ContractTemplate) *TemplateResponse {
	clauses := make([]SpecialClauseResponse, len(e.Data.SpecialClauses))
	for i, c := range e.Data.SpecialClauses {
		clauses[i] = SpecialClauseResponse{Title: c.Title, Description: c.Description}
	}
	blocks := make([]BlockResponse, len(e.Templates.Blocks))
	for i, b := range e.Templates.Blocks {
		blocks[i] = BlockResponse{Type: b.Type, Title: b.Title, Content: b.Content, Position: b.Position}
	}
	return &TemplateResponse{
		ID: e.ID, Name: e.Name, ContractType: string(e.ContractType),
		Description: e.Description, IsActive: e.IsActive,
		Data: TemplateDataResponse{
			DesignationID: e.Data.DesignationID, WorkingPatternID: e.Data.WorkingPatternID,
			JobDuties: e.Data.JobDuties, InventoryItems: e.Data.InventoryItems,
			SpecialClauses: clauses, EmployeeFields: e.Data.EmployeeFields,
		},
		Templates: TemplatePartialsResponse{Blocks: blocks},
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func toTemplatePrefillResponse(e *entity.ContractTemplate) *TemplatePrefillResponse {
	clauses := make([]SpecialClauseResponse, len(e.Data.SpecialClauses))
	for i, c := range e.Data.SpecialClauses {
		clauses[i] = SpecialClauseResponse{Title: c.Title, Description: c.Description}
	}
	return &TemplatePrefillResponse{
		ID: e.ID, Name: e.Name, ContractType: string(e.ContractType),
		Data: TemplateDataResponse{
			DesignationID: e.Data.DesignationID, WorkingPatternID: e.Data.WorkingPatternID,
			JobDuties: e.Data.JobDuties, InventoryItems: e.Data.InventoryItems,
			SpecialClauses: clauses, EmployeeFields: e.Data.EmployeeFields,
		},
	}
}

func toListItemResponse(e *entity.ContractTemplate) *TemplateListItemResponse {
	return &TemplateListItemResponse{
		ID: e.ID, Name: e.Name, ContractType: string(e.ContractType),
		Description: e.Description, IsActive: e.IsActive,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func toListResponse(entities []*entity.ContractTemplate, total int64) *ListTemplateResponse {
	items := make([]*TemplateListItemResponse, len(entities))
	for i, e := range entities {
		items[i] = toListItemResponse(e)
	}
	return &ListTemplateResponse{Items: items, Total: total}
}

func toContractResponse(e *entity.Contract) *ContractResponse {
	clauses := make([]SpecialClauseResponse, len(e.Data.SpecialClauses))
	for i, c := range e.Data.SpecialClauses {
		clauses[i] = SpecialClauseResponse{Title: c.Title, Description: c.Description}
	}
	blocks := make([]BlockResponse, len(e.Templates.Blocks))
	for i, b := range e.Templates.Blocks {
		blocks[i] = BlockResponse{Type: b.Type, Title: b.Title, Content: b.Content, Position: b.Position}
	}
	return &ContractResponse{
		ID: e.ID, TemplateID: e.TemplateID, EmployeeID: e.EmployeeID,
		Number: e.Number, StartDate: e.StartDate, EndDate: e.EndDate,
		Salary: e.Salary,
		DesignationTitle: e.DesignationTitle,
		Status: string(e.Status),
		Data: TemplateDataResponse{
			DesignationID: e.Data.DesignationID, WorkingPatternID: e.Data.WorkingPatternID,
			JobDuties: e.Data.JobDuties, InventoryItems: e.Data.InventoryItems,
			SpecialClauses: clauses, EmployeeFields: e.Data.EmployeeFields,
		},
		Templates: TemplatePartialsResponse{Blocks: blocks},
		SentAt: e.SentAt, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func toContractListItemResponse(e *entity.Contract) *ContractListItemResponse {
	return &ContractListItemResponse{
		ID: e.ID, TemplateID: e.TemplateID, EmployeeID: e.EmployeeID,
		Number: e.Number, DesignationID: e.DesignationID, DesignationTitle: e.DesignationTitle,
		Status: string(e.Status), ContractType: e.ContractType,
		StartDate: e.StartDate, EndDate: e.EndDate,
		Salary: e.Salary, DocumentID: e.DocumentID, SentAt: e.SentAt, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func toListContractResponse(entities []*entity.Contract, total int64, signingsByContract map[string][]*entity.ContractSigning, employeeBriefs map[string]usecase.EmployeeBrief) *ListContractResponse {
	if employeeBriefs == nil {
		employeeBriefs = make(map[string]usecase.EmployeeBrief)
	}
	items := make([]*ContractListItemResponse, len(entities))
	for i, e := range entities {
		item := toContractListItemResponse(e)
		if brief, ok := employeeBriefs[e.EmployeeID]; ok {
			item.Employee = &ContractListItemEmployeeBrief{
				Name:            brief.Name,
				ProfilePhotoURL: brief.ProfilePhotoURL,
			}
		}
		if signings, ok := signingsByContract[e.ID]; ok {
			for _, s := range signings {
				if s.Party == "first" {
					item.FirstSignedAt = &s.SignedAt
				}
				if s.Party == "second" {
					item.SecondSignedAt = &s.SignedAt
				}
			}
		}
		items[i] = item
	}
	return &ListContractResponse{Items: items, Total: total}
}

func toContractDetailResponse(e *entity.Contract, templateName, contractType string, signings []*entity.ContractSigning) *ContractDetailResponse {
	signingResponses := make([]ContractSigningResponse, len(signings))
	for i, s := range signings {
		signingResponses[i] = *toContractSigningResponse(s)
	}
	return &ContractDetailResponse{
		ID: e.ID, TemplateID: e.TemplateID, TemplateName: templateName, EmployeeID: e.EmployeeID,
		Number: e.Number, StartDate: e.StartDate, EndDate: e.EndDate,
		Salary: e.Salary, DesignationTitle: e.DesignationTitle,
		Status: string(e.Status), ContractType: contractType, JobDuties: e.Data.JobDuties,
		Signings: signingResponses, SentAt: e.SentAt, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

func toContractSigningResponse(e *entity.ContractSigning) *ContractSigningResponse {
	return &ContractSigningResponse{
		ID: e.ID, Party: e.Party,
		SignedBy: e.SignedBy, SignedByName: e.SignedByName, SignedByTitle: e.SignedByTitle,
		Place: e.Place, SignatureBase64: e.SignatureBase64, SignedAt: e.SignedAt,
	}
}

func toActiveContractResponse(e *entity.Contract) *ActiveContractResponse {
	return &ActiveContractResponse{
		ID:               e.ID,
		Number:           e.Number,
		ContractType:     e.ContractType,
		DesignationTitle: e.DesignationTitle,
		StartDate:        e.StartDate,
		EndDate:          e.EndDate,
	}
}

func toEmployeeContractResponse(e *entity.Contract) *EmployeeContractResponse {
	return &EmployeeContractResponse{
		ContractID:     e.ID,
		StartDate:      e.StartDate,
		Status:         string(e.Status),
		EndDate:        e.EndDate,
		DesignationID:  e.DesignationID,
		ContractType:   e.ContractType,
		JobDuties:      e.Data.JobDuties,
		InventoryItems: e.Data.InventoryItems,
	}
}
