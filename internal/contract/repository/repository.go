package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/models"
	"hrms/internal/employee/numbergen"
)

type TemplateRepository interface {
	Create(ctx context.Context, e *entity.ContractTemplate) error
	FindByID(ctx context.Context, id string) (*entity.ContractTemplate, error)
	FindByName(ctx context.Context, name string) (*entity.ContractTemplate, error)
	FindAll(ctx context.Context, filter models.ListTemplateInput) ([]*entity.ContractTemplate, int64, error)
	Update(ctx context.Context, e *entity.ContractTemplate, expectedUpdatedAt time.Time) error
	Delete(ctx context.Context, id string) error
	CountByTemplateID(ctx context.Context, templateID string) (int64, error)
}

type ContractRepository interface {
	WithTx(tx *sqlx.Tx) ContractRepository
	CreateContract(ctx context.Context, e *entity.Contract) error
	BulkCreateContracts(ctx context.Context, contracts []*entity.Contract) error
	FindContractByID(ctx context.Context, id string) (*entity.Contract, error)
	FindActiveByEmployeeID(ctx context.Context, employeeID string) (*entity.Contract, error)
	FindCurrentByEmployeeID(ctx context.Context, employeeID string) (*entity.Contract, error)
	FindActiveContractEmployeeIDs(ctx context.Context, employeeIDs []string) (map[string]*time.Time, error)
	FindAllContracts(ctx context.Context, filter models.ListContractInput) ([]*entity.Contract, int64, error)
	UpdateContract(ctx context.Context, e *entity.Contract) error
	DeleteContract(ctx context.Context, id string) error
	CountByEmployeeIDAndStatus(ctx context.Context, employeeID string, status string) (int64, error)
	FindActiveContractsPastEndDate(ctx context.Context, asOf time.Time) ([]*entity.Contract, error)
	CountSoonExpired(ctx context.Context, withinDays int) (int64, error)
	HasOtherActiveContract(ctx context.Context, employeeID, excludeContractID string) (bool, error)
	numbergen.SequenceRepository
}

type SigningRepository interface {
	WithTx(tx *sqlx.Tx) SigningRepository
	CreateContractSigning(ctx context.Context, e *entity.ContractSigning) error
	FindSigningsByContractID(ctx context.Context, contractID string) ([]*entity.ContractSigning, error)
	FindSigningsByContractIDs(ctx context.Context, contractIDs []string) (map[string][]*entity.ContractSigning, error)
}

type DocumentRepository interface {
	CreateContractDocument(ctx context.Context, e *entity.ContractDocument) error
	FindContractDocumentByContractID(ctx context.Context, contractID string) (*entity.ContractDocument, error)
}
