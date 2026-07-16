package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/models"
	"hrms/internal/contract/repository"
	errors "hrms/internal/pkg/apperror"
)

type EmployeeActivator interface {
	ActivateEmployee(ctx context.Context, employeeID string) error
}

type UserActivator interface {
	Activate(ctx context.Context, userID string) error
}

type EmployeeUserIDFinder interface {
	FindEmployeeUserID(ctx context.Context, employeeID string) (string, error)
}

type WorkPatternAssigner interface {
	AssignDefaultWorkPattern(ctx context.Context, employeeID string, validFrom time.Time) error
}

type SigningUsecase struct {
	db            *sqlx.DB
	contractRepo  repository.ContractRepository
	signingRepo   repository.SigningRepository
	docUC         *DocumentUsecase
	empFetcher    EmployeeFetcher
	empActivator  EmployeeActivator
	wpAssigner    WorkPatternAssigner
	userActivator UserActivator
	empFinder     EmployeeUserIDFinder
}

func NewSigningUsecase(db *sqlx.DB, contractRepo repository.ContractRepository, signingRepo repository.SigningRepository, docUC *DocumentUsecase, empFetcher EmployeeFetcher, empActivator EmployeeActivator, wpAssigner WorkPatternAssigner, userActivator UserActivator, empFinder EmployeeUserIDFinder) *SigningUsecase {
	return &SigningUsecase{db: db, contractRepo: contractRepo, signingRepo: signingRepo, docUC: docUC, empFetcher: empFetcher, empActivator: empActivator, wpAssigner: wpAssigner, userActivator: userActivator, empFinder: empFinder}
}

func (uc *SigningUsecase) BulkSign(ctx context.Context, input models.BulkSignContractInput) ([]*entity.Contract, error) {
	tx, err := uc.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to begin transaction: %v", err))
	}
	defer tx.Rollback()

	contractRepo := uc.contractRepo.WithTx(tx)
	signingRepo := uc.signingRepo.WithTx(tx)

	var results []*entity.Contract

	for _, contractID := range input.ContractIDs {
		e, err := contractRepo.FindContractByID(ctx, contractID)
		if err != nil {
			return nil, errors.NewInternal(fmt.Sprintf("failed to find contract %s: %v", contractID, err))
		}
		if e == nil {
			return nil, errors.NewNotFound("contract not found: " + contractID)
		}

		if err := e.CanSign(input.Party); err != nil {
			return nil, errors.NewInvalidInput(fmt.Sprintf("contract %s: %s", contractID, err.Error()))
		}

		// Stage 1: Record the signature
		signing := e.AddSignature(
			input.Party,
			input.SignedBy, input.SignedByName, input.SignedByTitle,
			input.Place, input.SignatureBase64,
		)

		if err := signingRepo.CreateContractSigning(ctx, signing); err != nil {
			return nil, errors.NewInternal(fmt.Sprintf("failed to create signing for contract %s: %v", contractID, err))
		}

		// Stage 2: Check signings and determine next status
		signings, err := signingRepo.FindSigningsByContractID(ctx, contractID)
		if err != nil {
			return nil, errors.NewInternal(fmt.Sprintf("failed to find signings for contract %s: %v", contractID, err))
		}

		shouldGeneratePDF, err := e.EvaluateSigningState(signings)
		if err != nil {
			return nil, fmt.Errorf("evaluate signing state for contract %s: %w", contractID, err)
		}

		// Stage 3: Save contract updates
		if err := contractRepo.UpdateContract(ctx, e); err != nil {
			return nil, errors.NewInternal(fmt.Sprintf("failed to update contract %s: %v", contractID, err))
		}

		// Stage 3b: When first-party signs (contract goes to "sent"),
		// reactivate the employee's user account so they can log in to sign.
		if !shouldGeneratePDF && e.Status == entity.ContractStatusSent && uc.userActivator != nil && uc.empFinder != nil {
			if empUserID, err := uc.empFinder.FindEmployeeUserID(ctx, e.EmployeeID); err == nil && empUserID != "" {
				_ = uc.userActivator.Activate(ctx, empUserID)
			}
		}

		// Stage 4: When new contract becomes active, expire any previous active contract
		if shouldGeneratePDF {
			if err := uc.expireOldActiveContractWithRepo(ctx, contractRepo, e); err != nil {
				return nil, errors.NewInternal(fmt.Sprintf("expire old active contract for %s: %v", contractID, err))
			}
		}

		// Stage 5: When new contract becomes active, activate the employee and assign default work pattern
		if shouldGeneratePDF && uc.empActivator != nil {
			if err := uc.empActivator.ActivateEmployee(ctx, e.EmployeeID); err != nil {
				return nil, errors.NewInternal(fmt.Sprintf("activate employee for contract %s: %v", contractID, err))
			}
		}
		if shouldGeneratePDF && uc.wpAssigner != nil && e.StartDate != nil {
			if err := uc.wpAssigner.AssignDefaultWorkPattern(ctx, e.EmployeeID, *e.StartDate); err != nil {
				return nil, errors.NewInternal(fmt.Sprintf("assign work pattern for contract %s: %v", contractID, err))
			}
		}

		// Stage 6: Generate and store the final PDF after both parties have signed
		if shouldGeneratePDF {
			if _, _, err := uc.docUC.StorePDFWithSignings(ctx, e.ID, e.Number, input.SignedByName, input.SignedByTitle, signings); err != nil {
				return nil, errors.NewInternal(fmt.Sprintf("failed to store signed PDF for contract %s: %v", contractID, err))
			}
			e.AttachDocument()
			if err := contractRepo.UpdateContract(ctx, e); err != nil {
				return nil, errors.NewInternal(fmt.Sprintf("failed to update contract %s after attaching doc: %v", contractID, err))
			}
		}

		results = append(results, e)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return results, nil
}

// expireOldActiveContractWithRepo is the transaction-aware version of expireOldActiveContract.
func (uc *SigningUsecase) expireOldActiveContractWithRepo(ctx context.Context, contractRepo repository.ContractRepository, newContract *entity.Contract) error {
	oldContract, err := contractRepo.FindActiveByEmployeeID(ctx, newContract.EmployeeID)
	if err != nil {
		return fmt.Errorf("find previous active contract: %w", err)
	}
	if oldContract == nil || oldContract.ID == newContract.ID {
		return nil
	}
	if err := oldContract.Expire(); err != nil {
		return fmt.Errorf("expire previous contract: %w", err)
	}
	return contractRepo.UpdateContract(ctx, oldContract)
}

func (uc *SigningUsecase) expireOldActiveContract(ctx context.Context, newContract *entity.Contract) error {
	oldContract, err := uc.contractRepo.FindActiveByEmployeeID(ctx, newContract.EmployeeID)
	if err != nil {
		return fmt.Errorf("find previous active contract: %w", err)
	}
	if oldContract == nil || oldContract.ID == newContract.ID {
		return nil
	}
	if err := oldContract.Expire(); err != nil {
		return fmt.Errorf("expire previous contract: %w", err)
	}
	return uc.contractRepo.UpdateContract(ctx, oldContract)
}

func (uc *SigningUsecase) BulkSignAsSecondParty(ctx context.Context, input models.BulkSignContractInput, userID string) ([]*entity.Contract, error) {
	employeeID, err := uc.empFetcher.FindEmployeeIDByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NewInvalidInput("failed to resolve user to employee: " + err.Error())
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("employee not found for authenticated user")
	}

	// Validate all contracts belong to this employee before signing any of them
	for _, contractID := range input.ContractIDs {
		e, err := uc.contractRepo.FindContractByID(ctx, contractID)
		if err != nil {
			return nil, errors.NewInternal(fmt.Sprintf("failed to find contract %s: %v", contractID, err))
		}
		if e == nil {
			return nil, errors.NewNotFound("contract not found: " + contractID)
		}
		if e.EmployeeID != employeeID {
			return nil, errors.NewInvalidInput(fmt.Sprintf("contract %s does not belong to authenticated employee", contractID))
		}
	}

	// Ownership validated — delegate to BulkSign with party forced to "second"
	input.Party = "second"
	return uc.BulkSign(ctx, input)
}
