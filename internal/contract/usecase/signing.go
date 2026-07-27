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
	AssignWorkPattern(ctx context.Context, employeeID string, workPatternID *string, validFrom time.Time) error
}

type DesignationAssigner interface {
	AssignDesignation(ctx context.Context, employeeID string, designationID *string) error
}

type SigningUsecase struct {
	db            *sqlx.DB
	contractRepo  repository.ContractRepository
	signingRepo   repository.SigningRepository
	docUC         *DocumentUsecase
	empFetcher    EmployeeFetcher
	empActivator  EmployeeActivator
	wpAssigner    WorkPatternAssigner
	desAssigner   DesignationAssigner
	userActivator UserActivator
	empFinder     EmployeeUserIDFinder
}

func NewSigningUsecase(db *sqlx.DB, contractRepo repository.ContractRepository, signingRepo repository.SigningRepository, docUC *DocumentUsecase, empFetcher EmployeeFetcher, empActivator EmployeeActivator, wpAssigner WorkPatternAssigner, desAssigner DesignationAssigner, userActivator UserActivator, empFinder EmployeeUserIDFinder) *SigningUsecase {
	return &SigningUsecase{db: db, contractRepo: contractRepo, signingRepo: signingRepo, docUC: docUC, empFetcher: empFetcher, empActivator: empActivator, wpAssigner: wpAssigner, desAssigner: desAssigner, userActivator: userActivator, empFinder: empFinder}
}

func (uc *SigningUsecase) BulkSign(ctx context.Context, input models.BulkSignContractInput) ([]*entity.Contract, error) {
	// Deduplicate contract IDs
	seen := make(map[string]struct{}, len(input.ContractIDs))
	uniqueIDs := make([]string, 0, len(input.ContractIDs))
	for _, id := range input.ContractIDs {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			uniqueIDs = append(uniqueIDs, id)
		}
	}
	input.ContractIDs = uniqueIDs

	tx, err := uc.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.WrapInternal("failed to begin transaction", err)
	}
	defer tx.Rollback()

	contractRepo := uc.contractRepo.WithTx(tx)
	signingRepo := uc.signingRepo.WithTx(tx)

	var results []*entity.Contract
	// Collect contracts that need post-commit side effects
	type pendingSideEffect struct {
		contract *entity.Contract
		signings []*entity.ContractSigning
	}
	var sideEffects []pendingSideEffect

	for _, contractID := range input.ContractIDs {
		e, err := contractRepo.FindContractByID(ctx, contractID)
		if err != nil {
			return nil, errors.WrapInternal(fmt.Sprintf("failed to find contract %s", contractID), err)
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
			return nil, errors.WrapInternal(fmt.Sprintf("failed to create signing for contract %s", contractID), err)
		}

		// Stage 2: Check signings and determine next status
		signings, err := signingRepo.FindSigningsByContractID(ctx, contractID)
		if err != nil {
			return nil, errors.WrapInternal(fmt.Sprintf("failed to find signings for contract %s", contractID), err)
		}

		shouldGeneratePDF, err := e.EvaluateSigningState(signings)
		if err != nil {
			return nil, fmt.Errorf("evaluate signing state for contract %s: %w", contractID, err)
		}

		// Stage 3: Save contract updates
		if err := contractRepo.UpdateContract(ctx, e); err != nil {
			return nil, errors.WrapInternal(fmt.Sprintf("failed to update contract %s", contractID), err)
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
				return nil, errors.WrapInternal(fmt.Sprintf("expire old active contract for %s", contractID), err)
			}
		}

		results = append(results, e)

		// Collect side effects to run after commit
		if shouldGeneratePDF {
			sideEffects = append(sideEffects, pendingSideEffect{contract: e, signings: signings})
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.WrapInternal("failed to commit transaction", err)
	}

	// Stage 5-6: Side effects after commit — cannot be rolled back
	for _, se := range sideEffects {
		if uc.empActivator != nil {
			_ = uc.empActivator.ActivateEmployee(ctx, se.contract.EmployeeID)
		}
		if uc.desAssigner != nil {
			_ = uc.desAssigner.AssignDesignation(ctx, se.contract.EmployeeID, se.contract.DesignationID)
		}
		if uc.wpAssigner != nil && se.contract.StartDate != nil {
			_ = uc.wpAssigner.AssignWorkPattern(ctx, se.contract.EmployeeID, se.contract.Data.WorkingPatternID, *se.contract.StartDate)
		}
		if _, _, err := uc.docUC.StorePDFWithSignings(ctx, se.contract.ID, se.contract.Number, input.SignedByName, input.SignedByTitle, se.signings); err != nil {
			return nil, errors.WrapInternal(fmt.Sprintf("failed to store signed PDF for contract %s", se.contract.ID), err)
		}
		se.contract.AttachDocument()
		if err := uc.contractRepo.UpdateContract(ctx, se.contract); err != nil {
			return nil, errors.WrapInternal(fmt.Sprintf("failed to update contract %s after attaching doc", se.contract.ID), err)
		}
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
			return nil, errors.WrapInternal(fmt.Sprintf("failed to find contract %s", contractID), err)
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
