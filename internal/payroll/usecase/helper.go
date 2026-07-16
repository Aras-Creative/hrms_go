package usecase

import (
	"context"

	"github.com/jmoiron/sqlx"

	"hrms/internal/payroll/entity"
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
		INSERT INTO employee_compensations (id, employee_id, compensation_item_id, amount, frequency, effective_date, end_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		ec.ID, ec.EmployeeID, ec.CompensationItemID, ec.Amount.Cents(), string(ec.Frequency), ec.EffectiveDate, ec.EndDate, ec.CreatedAt, ec.UpdatedAt)
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
		INSERT INTO employee_deductions (id, employee_id, deduction_type_id, value, effective_date, end_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		ed.ID, ed.EmployeeID, ed.DeductionTypeID, ed.Value, ed.EffectiveDate, ed.EndDate, ed.CreatedAt, ed.UpdatedAt)
	return err
}
