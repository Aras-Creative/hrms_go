package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type OverviewEmployee struct {
	EmployeeID      string  `db:"employee_id"`
	EmployeeName    string  `db:"full_name"`
	EmployeeNumber  string  `db:"employee_number"`
	DesignationName *string `db:"designation_name"`
	ProfilePhotoID  *string `db:"profile_photo_id"`
}

type PostgresOverviewRepo struct {
	db *sqlx.DB
}

func NewPostgresOverviewRepo(db *sqlx.DB) *PostgresOverviewRepo {
	return &PostgresOverviewRepo{db: db}
}

func (r *PostgresOverviewRepo) QueryEmployees(ctx context.Context, startDate, endDate interface{}) ([]OverviewEmployee, error) {
	var employees []OverviewEmployee
	err := r.db.SelectContext(ctx, &employees, `
		SELECT DISTINCT ebs.employee_id, e.full_name, e.employee_number,
			d.name AS designation_name, e.profile_photo_id
		FROM employee_base_salaries ebs
		JOIN employees e ON e.id = ebs.employee_id
		LEFT JOIN designations d ON d.id = e.designation_id
		WHERE ebs.effective_date <= $1::date
		  AND (ebs.end_date IS NULL OR ebs.end_date >= $2::date)
	`, endDate, startDate)
	if err != nil {
		return nil, fmt.Errorf("query employees: %w", err)
	}
	return employees, nil
}

// QueryCompensationRowsBatch returns the compensation rows per employee rather than a
// pre-summed total, so the caller can apply the shared per-row calculation (which
// attendance-derived allowances need) instead of a second, divergent formula.
func (r *PostgresOverviewRepo) QueryCompensationRowsBatch(ctx context.Context, employeeIDs []string, startDate, endDate interface{}) (map[string][]CalcCompRow, error) {
	result := make(map[string][]CalcCompRow, len(employeeIDs))
	if len(employeeIDs) == 0 {
		return result, nil
	}
	query, args, err := sqlx.In(`
		SELECT ec.employee_id, ci.id, ci.name, ec.amount, ec.frequency, ec.calc_type
		FROM employee_compensations ec
		JOIN compensation_items ci ON ci.id = ec.compensation_item_id
		WHERE ec.employee_id IN (?)
		  AND ec.frequency IN ('monthly', 'yearly')
		  AND ec.effective_date <= ? AND (ec.end_date IS NULL OR ec.end_date >= ?)
	`, employeeIDs, endDate, startDate)
	if err != nil {
		return nil, fmt.Errorf("build compensations batch query: %w", err)
	}
	query = r.db.Rebind(query)

	var rows []struct {
		EmployeeID string `db:"employee_id"`
		ID         string `db:"id"`
		Name       string `db:"name"`
		Amount     int64  `db:"amount"`
		Frequency  string `db:"frequency"`
		CalcType   string `db:"calc_type"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("query compensations batch: %w", err)
	}
	for _, row := range rows {
		result[row.EmployeeID] = append(result[row.EmployeeID], CalcCompRow{
			ID: row.ID, Name: row.Name, Amount: row.Amount, CalcType: row.CalcType,
		})
	}
	return result, nil
}

// QueryDeductionRowsBatch returns deduction rows per employee, mirroring
// QueryCompensationRowsBatch so the caller shares the processor's formula.
func (r *PostgresOverviewRepo) QueryDeductionRowsBatch(ctx context.Context, employeeIDs []string, startDate, endDate interface{}) (map[string][]CalcDedRow, error) {
	result := make(map[string][]CalcDedRow, len(employeeIDs))
	if len(employeeIDs) == 0 {
		return result, nil
	}
	query, args, err := sqlx.In(`
		SELECT ed.employee_id, dt.id, dt.name, dt.deduction_type,
			COALESCE(ed.value, dt.default_value) AS value,
			dt.value_source,
			COALESCE(ed.unit_amount, dt.unit_amount) AS unit_amount,
			(ed.unit_amount IS NOT NULL) AS unit_amount_overridden
		FROM employee_deductions ed
		JOIN deduction_types dt ON dt.id = ed.deduction_type_id
		WHERE ed.employee_id IN (?)
		  AND ed.effective_date <= ? AND (ed.end_date IS NULL OR ed.end_date >= ?)
		  AND dt.is_active = true
	`, employeeIDs, endDate, startDate)
	if err != nil {
		return nil, fmt.Errorf("build deductions batch query: %w", err)
	}
	query = r.db.Rebind(query)

	var rows []struct {
		EmployeeID   string  `db:"employee_id"`
		ID           string  `db:"id"`
		Name         string  `db:"name"`
		DeductionTyp string  `db:"deduction_type"`
		Value        float64 `db:"value"`
		ValueSource  string  `db:"value_source"`
		UnitAmount   int64   `db:"unit_amount"`
		Overridden   bool    `db:"unit_amount_overridden"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("query deductions batch: %w", err)
	}
	for _, row := range rows {
		result[row.EmployeeID] = append(result[row.EmployeeID], CalcDedRow{
			ID: row.ID, Name: row.Name, Type: row.DeductionTyp,
			Value: row.Value, ValueSource: row.ValueSource, UnitAmount: row.UnitAmount,
			UnitAmountOverridden: row.Overridden,
		})
	}
	return result, nil
}
