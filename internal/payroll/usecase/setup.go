package usecase

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/models"
	errors "hrms/internal/pkg/apperror"
)

type SetupUsecase struct {
	db              *sqlx.DB
	employeeFetcher EmployeeFetcher
}

func NewSetupUsecase(db *sqlx.DB, employeeFetcher EmployeeFetcher) *SetupUsecase {
	return &SetupUsecase{db: db, employeeFetcher: employeeFetcher}
}

func (uc *SetupUsecase) SetupEmployeePayroll(ctx context.Context, input models.SetupEmployeePayrollInput) error {
	exists, err := uc.employeeFetcher.ExistsByID(ctx, input.EmployeeID)
	if err != nil {
		return fmt.Errorf("check employee: %w", err)
	}
	if !exists {
		return errors.NewNotFound("employee not found")
	}

	tx, err := uc.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if input.BaseSalary != nil {
		if err := deleteCurrentSalaryTx(ctx, tx, input.EmployeeID); err != nil {
			return fmt.Errorf("delete current salary: %w", err)
		}
		amount, err := entity.NewAmount(input.BaseSalary.Amount)
		if err != nil {
			return errors.NewInvalidInput(err.Error())
		}
		currency, err := entity.NewCurrency(input.BaseSalary.Currency)
		if err != nil {
			return errors.NewInvalidInput(err.Error())
		}
		s := entity.NewEmployeeBaseSalary(input.EmployeeID, amount, currency, input.BaseSalary.EffectiveDate, input.BaseSalary.EndDate, input.BaseSalary.Notes)
		if err := insertSalaryTx(ctx, tx, s); err != nil {
			return fmt.Errorf("insert salary: %w", err)
		}
	}

	if err := deleteEmployeeCompensationsTx(ctx, tx, input.EmployeeID); err != nil {
		return fmt.Errorf("delete compensations: %w", err)
	}
	for _, c := range input.Compensations {
		amount, err := entity.NewAmount(c.Amount)
		if err != nil {
			return errors.WrapInvalidInput(fmt.Sprintf("compensation %s", c.CompensationItemID), err)
		}
		freq, err := entity.ParseFrequency(c.Frequency)
		if err != nil {
			return errors.WrapInvalidInput(fmt.Sprintf("compensation %s", c.CompensationItemID), err)
		}
		ec := entity.NewEmployeeCompensation(input.EmployeeID, c.CompensationItemID, amount, freq, c.EffectiveDate, c.EndDate)
		if err := insertEmpCompTx(ctx, tx, ec); err != nil {
			return fmt.Errorf("insert compensation: %w", err)
		}
	}

	if err := deleteEmployeeBenefitsTx(ctx, tx, input.EmployeeID); err != nil {
		return fmt.Errorf("delete benefits: %w", err)
	}
	for _, b := range input.Benefits {
		eb := entity.NewEmployeeBenefit(input.EmployeeID, b.BenefitTypeID, b.ParticipantNumber, b.EffectiveDate, b.EndDate)
		if err := insertEmpBenefitTx(ctx, tx, eb); err != nil {
			return fmt.Errorf("insert benefit: %w", err)
		}
	}

	if err := deleteEmployeeDeductionsTx(ctx, tx, input.EmployeeID); err != nil {
		return fmt.Errorf("delete deductions: %w", err)
	}
	for _, d := range input.Deductions {
		ed := entity.NewEmployeeDeduction(input.EmployeeID, d.DeductionTypeID, d.Value, d.EffectiveDate, d.EndDate)
		if err := insertEmpDeductionTx(ctx, tx, ed); err != nil {
			return fmt.Errorf("insert deduction: %w", err)
		}
	}

	return tx.Commit()
}
