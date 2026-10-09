package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/payroll/entity"
)

// ---------------------------------------------------------------------------
// PayrollPeriodRepository
// ---------------------------------------------------------------------------

type PostgresPayrollPeriodRepo struct {
	db *sqlx.DB
}

func NewPostgresPayrollPeriodRepo(db *sqlx.DB) *PostgresPayrollPeriodRepo {
	return &PostgresPayrollPeriodRepo{db: db}
}

const qryInsertPeriod = `
	INSERT INTO payroll_periods (id, name, start_date, end_date, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`

const qrySelectPeriod = `
	SELECT id, name, start_date, end_date, status, created_at, updated_at
	FROM payroll_periods
`

const qryUpdatePeriod = `
	UPDATE payroll_periods SET
		name = $1, start_date = $2, end_date = $3, status = $4, updated_at = $5
	WHERE id = $6
`

const qryDeletePeriod = `DELETE FROM payroll_periods WHERE id = $1`

func (r *PostgresPayrollPeriodRepo) Create(ctx context.Context, p *entity.PayrollPeriod) error {
	_, err := r.db.ExecContext(ctx, qryInsertPeriod,
		p.ID, p.Name, p.StartDate, p.EndDate, p.Status, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert period: %w", err)
	}
	return nil
}

func (r *PostgresPayrollPeriodRepo) FindByID(ctx context.Context, id string) (*entity.PayrollPeriod, error) {
	var m PayrollPeriodModel
	err := r.db.QueryRowxContext(ctx, qrySelectPeriod+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find period by id: %w", err)
	}
	return periodModelToEntity(&m), nil
}

func (r *PostgresPayrollPeriodRepo) FindAll(ctx context.Context, page, perPage int) ([]*entity.PayrollPeriod, int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM payroll_periods`)
	if err != nil {
		return nil, 0, fmt.Errorf("count periods: %w", err)
	}

	var models []PayrollPeriodModel
	err = r.db.SelectContext(ctx, &models, qrySelectPeriod+` ORDER BY start_date DESC LIMIT $1 OFFSET $2`, perPage, (page-1)*perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("list periods: %w", err)
	}
	result := make([]*entity.PayrollPeriod, len(models))
	for i := range models {
		result[i] = periodModelToEntity(&models[i])
	}
	return result, total, nil
}

func (r *PostgresPayrollPeriodRepo) FindByOverlap(ctx context.Context, startDate, endDate time.Time, excludeID string) (*entity.PayrollPeriod, error) {
	var m PayrollPeriodModel
	err := r.db.QueryRowxContext(ctx, `
		`+qrySelectPeriod+`
		WHERE start_date <= $2::date AND end_date >= $1::date
		  AND ($3 = '' OR id != $3::uuid)
		LIMIT 1
	`, startDate, endDate, excludeID).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find overlapping period: %w", err)
	}
	return periodModelToEntity(&m), nil
}

func (r *PostgresPayrollPeriodRepo) Update(ctx context.Context, p *entity.PayrollPeriod) error {
	res, err := r.db.ExecContext(ctx, qryUpdatePeriod,
		p.Name, p.StartDate, p.EndDate, p.Status, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update period: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return nil
	}
	return nil
}

func (r *PostgresPayrollPeriodRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, qryDeletePeriod, id)
	if err != nil {
		return fmt.Errorf("delete period: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return nil
	}
	return nil
}

func periodModelToEntity(m *PayrollPeriodModel) *entity.PayrollPeriod {
	return entity.ReconstitutePayrollPeriod(
		m.ID, m.Name, m.StartDate, m.EndDate, entity.PeriodStatus(m.Status), m.CreatedAt, m.UpdatedAt,
	)
}

// ---------------------------------------------------------------------------
// PaySlipRepository
// ---------------------------------------------------------------------------

type PostgresPaySlipRepo struct {
	db *sqlx.DB
}

func NewPostgresPaySlipRepo(db *sqlx.DB) *PostgresPaySlipRepo {
	return &PostgresPaySlipRepo{db: db}
}

const qryUpsertPaySlip = `
	INSERT INTO pay_slips (id, period_id, employee_id, base_salary, total_compensations, total_deductions,
		absent_days, net_salary, currency, source, base_salary_label, compensations_breakdown, deductions_breakdown, income_inputs, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (period_id, employee_id) DO UPDATE SET
		base_salary = EXCLUDED.base_salary,
		total_compensations = EXCLUDED.total_compensations,
		total_deductions = EXCLUDED.total_deductions,
		absent_days = EXCLUDED.absent_days,
		net_salary = EXCLUDED.net_salary,
		currency = EXCLUDED.currency,
		source = EXCLUDED.source,
		base_salary_label = EXCLUDED.base_salary_label,
		compensations_breakdown = EXCLUDED.compensations_breakdown,
		deductions_breakdown = EXCLUDED.deductions_breakdown,
		income_inputs = EXCLUDED.income_inputs,
		updated_at = EXCLUDED.updated_at
`

// qryUpsertPaySlipIfNotManual is the processor's upsert. The WHERE clause on the DO
// UPDATE turns the conflict into a no-op when a manual override already exists, so
// re-processing a period cannot silently revert a deliberate human correction.
// RowsAffected == 0 therefore means "skipped, manual override preserved".
const qryUpsertPaySlipIfNotManual = `
	-- income_inputs and base_salary_label are deliberately left out of the column list and
	-- of the DO UPDATE SET. Re-processing recalculates money from master data; it has no
	-- source for a figure an admin typed by hand, so overwriting it would erase their work.
	-- Keeping them out of the DO UPDATE also means a slip that skipped this insert keeps its
	-- existing values.
	INSERT INTO pay_slips (id, period_id, employee_id, base_salary, total_compensations, total_deductions,
		absent_days, net_salary, currency, source, compensations_breakdown, deductions_breakdown, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT (period_id, employee_id) DO UPDATE SET
		base_salary = EXCLUDED.base_salary,
		total_compensations = EXCLUDED.total_compensations,
		total_deductions = EXCLUDED.total_deductions,
		absent_days = EXCLUDED.absent_days,
		net_salary = EXCLUDED.net_salary,
		currency = EXCLUDED.currency,
		source = EXCLUDED.source,
		compensations_breakdown = EXCLUDED.compensations_breakdown,
		deductions_breakdown = EXCLUDED.deductions_breakdown,
		updated_at = EXCLUDED.updated_at
	WHERE pay_slips.source <> 'manual'
`

// qryUpdatePaySlip persists an in-place edit of an existing slip by primary key.
const qryUpdatePaySlip = `
	UPDATE pay_slips SET
		base_salary = $2,
		total_compensations = $3,
		total_deductions = $4,
		absent_days = $5,
		net_salary = $6,
		currency = $7,
		source = $8,
		base_salary_label = $9,
		compensations_breakdown = $10,
		deductions_breakdown = $11,
		income_inputs = $12,
		updated_at = $13
	WHERE id = $1
`

const qrySelectPaySlip = `
	SELECT id, period_id, employee_id, base_salary, total_compensations, total_deductions,
		absent_days, net_salary, currency, source, base_salary_label, compensations_breakdown, deductions_breakdown, income_inputs, created_at, updated_at
	FROM pay_slips
`

func (r *PostgresPaySlipRepo) Upsert(ctx context.Context, ps *entity.PaySlip) error {
	compJSON, _ := json.Marshal(ps.CompensationsBreakdown)
	dedJSON, _ := json.Marshal(ps.DeductionsBreakdown)
	incomeJSON := marshalIncomeInputs(ps.IncomeInputs)

	_, err := r.db.ExecContext(ctx, qryUpsertPaySlip,
		ps.ID, ps.PeriodID, ps.EmployeeID,
		ps.BaseSalary.Cents(), ps.TotalCompensations.Cents(), ps.TotalDeductions.Cents(),
		ps.AbsentDays, ps.NetSalary.Cents(), ps.Currency.String(), string(ps.Source),
		ps.BaseSalaryLabel, compJSON, dedJSON, incomeJSON, ps.CreatedAt, ps.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert pay slip: %w", err)
	}
	return nil
}

func (r *PostgresPaySlipRepo) UpsertIfNotManual(ctx context.Context, ps *entity.PaySlip) (bool, error) {
	compJSON, _ := json.Marshal(ps.CompensationsBreakdown)
	dedJSON, _ := json.Marshal(ps.DeductionsBreakdown)

	res, err := r.db.ExecContext(ctx, qryUpsertPaySlipIfNotManual,
		ps.ID, ps.PeriodID, ps.EmployeeID,
		ps.BaseSalary.Cents(), ps.TotalCompensations.Cents(), ps.TotalDeductions.Cents(),
		ps.AbsentDays, ps.NetSalary.Cents(), ps.Currency.String(), string(ps.Source),
		compJSON, dedJSON, ps.CreatedAt, ps.UpdatedAt,
	)
	if err != nil {
		return false, fmt.Errorf("upsert pay slip: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return n == 0, nil
}

// Update replaces an existing payslip in place, keyed by its primary key. Unlike
// Upsert it never touches the (period_id, employee_id) identity of the slip.
func (r *PostgresPaySlipRepo) Update(ctx context.Context, ps *entity.PaySlip) error {
	compJSON, _ := json.Marshal(ps.CompensationsBreakdown)
	dedJSON, _ := json.Marshal(ps.DeductionsBreakdown)
	incomeJSON := marshalIncomeInputs(ps.IncomeInputs)

	res, err := r.db.ExecContext(ctx, qryUpdatePaySlip,
		ps.ID,
		ps.BaseSalary.Cents(), ps.TotalCompensations.Cents(), ps.TotalDeductions.Cents(),
		ps.AbsentDays, ps.NetSalary.Cents(), ps.Currency.String(), string(ps.Source),
		ps.BaseSalaryLabel, compJSON, dedJSON, incomeJSON, ps.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update pay slip: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresPaySlipRepo) FindByPeriodID(ctx context.Context, periodID string) ([]*entity.PaySlip, error) {
	var models []PaySlipModel
	err := r.db.SelectContext(ctx, &models, qrySelectPaySlip+` WHERE period_id = $1 ORDER BY employee_id`, periodID)
	if err != nil {
		return nil, fmt.Errorf("find pay slips by period: %w", err)
	}
	result := make([]*entity.PaySlip, len(models))
	for i := range models {
		result[i] = paySlipModelToEntity(&models[i])
	}
	return result, nil
}

func (r *PostgresPaySlipRepo) FindByID(ctx context.Context, id string) (*entity.PaySlip, error) {
	var m PaySlipModel
	err := r.db.QueryRowxContext(ctx, qrySelectPaySlip+` WHERE id = $1`, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find pay slip by id: %w", err)
	}
	return paySlipModelToEntity(&m), nil
}

func (r *PostgresPaySlipRepo) FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.PaySlip, error) {
	var models []PaySlipModel
	if err := r.db.SelectContext(ctx, &models, qrySelectPaySlip+` WHERE employee_id = $1 ORDER BY created_at DESC`, employeeID); err != nil {
		return nil, fmt.Errorf("find pay slips by employee: %w", err)
	}
	result := make([]*entity.PaySlip, len(models))
	for i := range models {
		result[i] = paySlipModelToEntity(&models[i])
	}
	return result, nil
}

func (r *PostgresPaySlipRepo) FindByEmployeeAndPeriod(ctx context.Context, employeeID, periodID string) (*entity.PaySlip, error) {
	var m PaySlipModel
	err := r.db.QueryRowxContext(ctx, qrySelectPaySlip+` WHERE employee_id = $1 AND period_id = $2`, employeeID, periodID).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find pay slip by employee and period: %w", err)
	}
	return paySlipModelToEntity(&m), nil
}

func (r *PostgresPaySlipRepo) CountByPeriodID(ctx context.Context, periodID string) (int64, error) {
	var n int64
	if err := r.db.GetContext(ctx, &n, `SELECT COUNT(*) FROM pay_slips WHERE period_id = $1`, periodID); err != nil {
		return 0, fmt.Errorf("count pay slips by period: %w", err)
	}
	return n, nil
}

func (r *PostgresPaySlipRepo) DeleteByPeriodID(ctx context.Context, periodID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM pay_slips WHERE period_id = $1`, periodID)
	if err != nil {
		return fmt.Errorf("delete payslips by period: %w", err)
	}
	return nil
}

// marshalIncomeInputs renders the descriptive figures for storage. A nil or empty slice
// becomes an empty JSON array rather than "null", so the column never holds a null document
// that a reader has to guard against.
func marshalIncomeInputs(inputs []entity.IncomeInput) []byte {
	if inputs == nil {
		inputs = []entity.IncomeInput{}
	}
	b, err := json.Marshal(inputs)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func paySlipModelToEntity(m *PaySlipModel) *entity.PaySlip {
	return entity.ReconstitutePaySlip(
		m.ID, m.PeriodID, m.EmployeeID,
		m.BaseSalary, m.TotalCompensations, m.TotalDeductions,
		m.AbsentDays, m.NetSalary, m.Currency, m.Source, m.BaseSalaryLabel,
		[]byte(m.CompensationsBreakdown), []byte(m.DeductionsBreakdown), []byte(m.IncomeInputs),
		m.CreatedAt, m.UpdatedAt,
	)
}

// ---------------------------------------------------------------------------
// DB model structs
// ---------------------------------------------------------------------------

type PayrollPeriodModel struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PaySlipModel struct {
	ID                     string          `db:"id"`
	PeriodID               string          `db:"period_id"`
	EmployeeID             string          `db:"employee_id"`
	BaseSalary             int64           `db:"base_salary"`
	TotalCompensations     int64           `db:"total_compensations"`
	TotalDeductions        int64           `db:"total_deductions"`
	AbsentDays             int             `db:"absent_days"`
	NetSalary              int64           `db:"net_salary"`
	Currency               string          `db:"currency"`
	Source                 string          `db:"source"`
	BaseSalaryLabel        string          `db:"base_salary_label"`
	CompensationsBreakdown json.RawMessage `db:"compensations_breakdown"`
	DeductionsBreakdown    json.RawMessage `db:"deductions_breakdown"`
	IncomeInputs           json.RawMessage `db:"income_inputs"`
	CreatedAt              time.Time       `db:"created_at"`
	UpdatedAt              time.Time       `db:"updated_at"`
}
