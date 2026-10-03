package usecase

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/payroll/entity"
	errors "hrms/internal/pkg/apperror"
)

// --- transactional helpers ---

func deleteCurrentSalaryTx(ctx context.Context, tx *sqlx.Tx, employeeID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM employee_base_salaries WHERE employee_id = $1 AND end_date IS NULL`, employeeID)
	return err
}

func insertSalaryTx(ctx context.Context, tx *sqlx.Tx, s *entity.EmployeeBaseSalary) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO employee_base_salaries (id, employee_id, amount, currency, effective_date, end_date, notes, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		s.ID, s.EmployeeID, s.Amount.Cents(), s.Currency.String(), s.EffectiveDate, s.EndDate, s.Notes, s.CreatedAt, s.UpdatedAt)
	return err
}

func deleteEmployeeCompensationsTx(ctx context.Context, tx *sqlx.Tx, employeeID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM employee_compensations WHERE employee_id = $1`, employeeID)
	return err
}

func insertEmpCompTx(ctx context.Context, tx *sqlx.Tx, ec *entity.EmployeeCompensation) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO employee_compensations (id, employee_id, compensation_item_id, amount, frequency, calc_type, effective_date, end_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		ec.ID, ec.EmployeeID, ec.CompensationItemID, ec.Amount.Cents(), string(ec.Frequency), string(ec.CalcType), ec.EffectiveDate, ec.EndDate, ec.CreatedAt, ec.UpdatedAt)
	return err
}

func deleteEmployeeBenefitsTx(ctx context.Context, tx *sqlx.Tx, employeeID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM employee_benefits WHERE employee_id = $1`, employeeID)
	return err
}

func insertEmpBenefitTx(ctx context.Context, tx *sqlx.Tx, eb *entity.EmployeeBenefit) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO employee_benefits (id, employee_id, benefit_type_id, participant_number, effective_date, end_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		eb.ID, eb.EmployeeID, eb.BenefitTypeID, eb.ParticipantNumber, eb.EffectiveDate, eb.EndDate, eb.CreatedAt, eb.UpdatedAt)
	return err
}

func deleteEmployeeDeductionsTx(ctx context.Context, tx *sqlx.Tx, employeeID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM employee_deductions WHERE employee_id = $1`, employeeID)
	return err
}

func insertEmpDeductionTx(ctx context.Context, tx *sqlx.Tx, ed *entity.EmployeeDeduction) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO employee_deductions (id, employee_id, deduction_type_id, value, unit_amount, effective_date, end_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		ed.ID, ed.EmployeeID, ed.DeductionTypeID, ed.Value, ed.UnitAmount,
		ed.EffectiveDate, ed.EndDate, ed.CreatedAt, ed.UpdatedAt)
	return err
}

// qryDedTypeForValidation reads only what ValidateAssignmentValue needs. The accepted
// range of a value depends on the master's deduction_type, so the check cannot run until
// this row is resolved.
const qryDedTypeForValidation = `
	SELECT name, deduction_type FROM deduction_types WHERE id = $1`

// deductionTypeForValidation is a local row shape rather than entity.DeductionType: that
// struct is a domain type with no db tags, and mapping it here would mean adding tags to
// domain code for the sake of one validation query.
type deductionTypeForValidation struct {
	Name          string                   `db:"name"`
	DeductionType entity.DeductionCalcType `db:"deduction_type"`
}

// validateDeductionAssignment rejects a value that the resolved deduction type could not
// legally read, before any row is written. The caller runs it inside the setup
// transaction, which has already removed the employee's previous deductions, so failing
// here rolls the whole replacement back rather than leaving the employee with none.
func validateDeductionAssignment(ctx context.Context, tx *sqlx.Tx, typeID string, value *float64) error {
	if value == nil {
		return nil
	}
	var row deductionTypeForValidation
	if err := tx.GetContext(ctx, &row, qryDedTypeForValidation, typeID); err != nil {
		return fmt.Errorf("load deduction type %s: %w", typeID, err)
	}
	dt := entity.DeductionType{Name: row.Name, DeductionType: row.DeductionType}
	// Wrapped as a domain error so the caller sees 400 with the explanation, rather than
	// the generic 500 that response.Error falls back to for unclassified errors. The value
	// came from the request, so reporting it as a server fault would misdirect the client.
	if err := dt.ValidateAssignmentValue(*value); err != nil {
		return errors.NewInvalidInput(err.Error())
	}
	return nil
}
