package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

type CalcSalaryRow struct {
	EmployeeID string `db:"employee_id"`
	Amount     int64  `db:"amount"`
	Currency   string `db:"currency"`
}

type CalcCompRow struct {
	ID     string `db:"id"`
	Name   string `db:"name"`
	Amount int64  `db:"amount"`
}

type CalcDedRow struct {
	ID    string  `db:"id"`
	Name  string  `db:"name"`
	Type  string  `db:"deduction_type"`
	Value float64 `db:"value"`
}

func (d CalcDedRow) CalculateCents(salaryCents int64) int64 {
	if d.Type == "percentage" {
		return int64(math.Round(float64(salaryCents) * d.Value / 100))
	}
	return int64(math.Round(d.Value * 100))
}

type CalcAbsentResult struct {
	AbsentDays int `db:"absent_days"`
}

type PostgresCalculationRepo struct {
	db *sqlx.DB
}

func NewPostgresCalculationRepo(db *sqlx.DB) *PostgresCalculationRepo {
	return &PostgresCalculationRepo{db: db}
}

const qryCalcActiveSalaries = `
	SELECT employee_id, amount, currency FROM employee_base_salaries
	WHERE effective_date <= $2::date AND (end_date IS NULL OR end_date >= $1::date)
`

const qryCalcAbsentDays = `
	SELECT COUNT(*) AS absent_days FROM daily_attendances da
	LEFT JOIN leave_submissions ls ON ls.id = da.leave_submission_id AND ls.status = 'approved'
	LEFT JOIN leave_types lt ON lt.id = ls.leave_type_id
	WHERE da.employee_id = $1
	  AND da.date >= $2::date AND da.date <= $3::date
	  AND (da.status = 'absent' OR (da.status = 'on_leave' AND lt.is_paid = false))
`

const qryCalcCompensations = `
	SELECT ci.id, ci.name, ec.amount FROM employee_compensations ec
	JOIN compensation_items ci ON ci.id = ec.compensation_item_id
	WHERE ec.employee_id = $1
	  AND ec.frequency IN ('monthly', 'yearly')
	  AND ec.effective_date <= $3::date AND (ec.end_date IS NULL OR ec.end_date >= $2::date)
`

const qryCalcDeductions = `
	SELECT dt.id, dt.name, dt.deduction_type,
		COALESCE(ed.value, dt.default_value) AS value
	FROM employee_deductions ed
	JOIN deduction_types dt ON dt.id = ed.deduction_type_id
	WHERE ed.employee_id = $1
	  AND ed.effective_date <= $3::date AND (ed.end_date IS NULL OR ed.end_date >= $2::date)
	  AND dt.is_active = true AND (dt.slug IS NULL OR dt.slug != 'absent')
`

func (r *PostgresCalculationRepo) QueryActiveSalaries(ctx context.Context, startDate, endDate time.Time) ([]CalcSalaryRow, error) {
	var rows []CalcSalaryRow
	err := r.db.SelectContext(ctx, &rows, qryCalcActiveSalaries, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query active salaries: %w", err)
	}
	return rows, nil
}

func (r *PostgresCalculationRepo) QueryActiveSalariesByIDs(ctx context.Context, startDate, endDate time.Time, employeeIDs []string) ([]CalcSalaryRow, error) {
	query, args, err := sqlx.In(`
		SELECT employee_id, amount, currency FROM employee_base_salaries
		WHERE effective_date <= ? AND (end_date IS NULL OR end_date >= ?)
		  AND employee_id IN (?)
	`, endDate, startDate, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	query = r.db.Rebind(query)

	var rows []CalcSalaryRow
	err = r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query active salaries by ids: %w", err)
	}
	return rows, nil
}

func (r *PostgresCalculationRepo) QueryAbsentDays(ctx context.Context, employeeID string, startDate, endDate time.Time) (int, error) {
	var result CalcAbsentResult
	err := r.db.GetContext(ctx, &result, qryCalcAbsentDays, employeeID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("count absent days: %w", err)
	}
	return result.AbsentDays, nil
}

func (r *PostgresCalculationRepo) QueryEmployeeCompensations(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]CalcCompRow, error) {
	var rows []CalcCompRow
	err := r.db.SelectContext(ctx, &rows, qryCalcCompensations, employeeID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query employee compensations: %w", err)
	}
	return rows, nil
}

func (r *PostgresCalculationRepo) QueryEmployeeDeductions(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]CalcDedRow, error) {
	var rows []CalcDedRow
	err := r.db.SelectContext(ctx, &rows, qryCalcDeductions, employeeID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query employee deductions: %w", err)
	}
	return rows, nil
}

func (r *PostgresCalculationRepo) QueryEmployeeWorkingDaysBatch(ctx context.Context, employeeIDs []string, startDate, endDate time.Time) (map[string]int, error) {
	if len(employeeIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
		WITH ewp AS (
			SELECT ewp2.employee_id, ewp2.work_pattern_id
			FROM employee_work_patterns ewp2
			WHERE ewp2.employee_id IN (?)
			  AND ewp2.valid_from <= ?::date
			  AND (ewp2.valid_to IS NULL OR ewp2.valid_to >= ?::date)
			  AND ewp2.is_active = true
		),
		pattern_days AS (
			SELECT ewp.employee_id, wpd.day_of_week
			FROM ewp
			JOIN work_patterns wp ON wp.id = ewp.work_pattern_id AND wp.is_active = true
			JOIN work_pattern_details wpd ON wpd.work_pattern_id = wp.id
			WHERE wpd.start_time IS NOT NULL
		),
		overrides AS (
			SELECT eso.employee_id, eso.date, eso.is_working_day
			FROM employee_schedule_overrides eso
			WHERE eso.employee_id IN (?)
			  AND eso.date >= ?::date AND eso.date <= ?::date
		),
		base_days AS (
			SELECT pd.employee_id,
				COUNT(*) FILTER (
					WHERE d.dt >= ?::date AND d.dt <= ?::date
				) AS working_days
			FROM pattern_days pd
			JOIN generate_series(?::date, ?::date, '1 day'::interval) d(dt)
				ON EXTRACT(DOW FROM d.dt)::int = pd.day_of_week
			GROUP BY pd.employee_id
		),
		override_adjustments AS (
			SELECT o.employee_id,
				COUNT(*) FILTER (WHERE o.is_working_day = true) AS added,
				COUNT(*) FILTER (
					WHERE o.is_working_day = false
					  AND EXISTS (
						SELECT 1 FROM pattern_days pd2
						WHERE pd2.employee_id = o.employee_id
						  AND pd2.day_of_week = EXTRACT(DOW FROM o.date)::int
					)
				) AS removed
			FROM overrides o
			GROUP BY o.employee_id
		)
		SELECT COALESCE(bd.employee_id, oa.employee_id) AS employee_id,
			COALESCE(bd.working_days, 0)
				+ COALESCE(oa.added, 0)
				- COALESCE(oa.removed, 0) AS working_days
		FROM base_days bd
		FULL OUTER JOIN override_adjustments oa ON oa.employee_id = bd.employee_id
	`, employeeIDs, startDate, endDate, employeeIDs, startDate, endDate, startDate, endDate, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("build working days batch query: %w", err)
	}
	query = r.db.Rebind(query)

	type row struct {
		EmployeeID  string `db:"employee_id"`
		WorkingDays int    `db:"working_days"`
	}
	var rows []row
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("query working days batch: %w", err)
	}

	result := make(map[string]int, len(employeeIDs))
	for _, r := range rows {
		result[r.EmployeeID] = r.WorkingDays
	}
	return result, nil
}
