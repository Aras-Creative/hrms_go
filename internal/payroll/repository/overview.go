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

func (r *PostgresOverviewRepo) QueryTotalCompensationsBatch(ctx context.Context, employeeIDs []string, startDate, endDate interface{}) (map[string]float64, error) {
	query, args, err := sqlx.In(`
		SELECT employee_id, COALESCE(SUM(amount), 0) AS total FROM employee_compensations
		WHERE employee_id IN (?) AND frequency IN ('monthly', 'yearly')
		  AND effective_date <= ? AND (end_date IS NULL OR end_date >= ?)
		GROUP BY employee_id
	`, employeeIDs, endDate, startDate)
	if err != nil {
		return nil, fmt.Errorf("build compensations batch query: %w", err)
	}
	query = r.db.Rebind(query)

	var rows []struct {
		EmployeeID string  `db:"employee_id"`
		Total      float64 `db:"total"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("query compensations batch: %w", err)
	}
	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		result[row.EmployeeID] = row.Total / 100
	}
	return result, nil
}

func (r *PostgresOverviewRepo) QueryTotalDeductionsBatch(ctx context.Context, employeeIDs []string, startDate, endDate interface{}, salaryCentsMap map[string]int64) (map[string]float64, error) {
	query, args, err := sqlx.In(`
		SELECT ed.employee_id,
			dt.deduction_type,
			COALESCE(ed.value, dt.default_value) AS ded_value
		FROM employee_deductions ed
		JOIN deduction_types dt ON dt.id = ed.deduction_type_id
		WHERE ed.employee_id IN (?)
		  AND ed.effective_date <= ? AND (ed.end_date IS NULL OR ed.end_date >= ?)
		  AND dt.is_active = true AND (dt.slug IS NULL OR dt.slug != 'absent')
	`, employeeIDs, endDate, startDate)
	if err != nil {
		return nil, fmt.Errorf("build deductions batch query: %w", err)
	}
	query = r.db.Rebind(query)

	var rows []struct {
		EmployeeID   string  `db:"employee_id"`
		DeductionType string `db:"deduction_type"`
		DedValue     float64 `db:"ded_value"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("query deductions batch: %w", err)
	}
	totals := make(map[string]float64, len(salaryCentsMap))
	for _, row := range rows {
		var amount float64
		if row.DeductionType == "percentage" {
			amount = float64(salaryCentsMap[row.EmployeeID]) * row.DedValue / 100
		} else {
			amount = row.DedValue * 100
		}
		totals[row.EmployeeID] += amount
	}
	result := make(map[string]float64, len(salaryCentsMap))
	for empID, total := range totals {
		result[empID] = total / 100
	}
	return result, nil
}

func (r *PostgresOverviewRepo) QueryAbsentDaysBatch(ctx context.Context, employeeIDs []string, startDate, endDate interface{}) (map[string]int, error) {
	if len(employeeIDs) == 0 {
		return make(map[string]int), nil
	}
	query, args, err := sqlx.In(`
		SELECT da.employee_id, COUNT(*)::int AS absent_days
		FROM daily_attendances da
		LEFT JOIN leave_submissions ls ON ls.id = da.leave_submission_id AND ls.status = 'approved'
		LEFT JOIN leave_types lt ON lt.id = ls.leave_type_id
		WHERE da.employee_id IN (?)
		  AND da.date >= ?::date AND da.date <= ?::date
		  AND (da.status = 'absent' OR (da.status = 'on_leave' AND lt.is_paid = false))
		GROUP BY da.employee_id
	`, employeeIDs, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("build absent days batch query: %w", err)
	}
	query = r.db.Rebind(query)

	var rows []struct {
		EmployeeID string `db:"employee_id"`
		AbsentDays int    `db:"absent_days"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("query absent days batch: %w", err)
	}
	result := make(map[string]int, len(rows))
	for _, row := range rows {
		result[row.EmployeeID] = row.AbsentDays
	}
	return result, nil
}
