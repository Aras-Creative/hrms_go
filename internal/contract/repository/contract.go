package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/models"
	"hrms/internal/employee/numbergen"
)

// txContext is satisfied by both *sqlx.DB and *sqlx.Tx.
type txContext interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
}

type ContractModel struct {
	ID                string     `db:"id"`
	TemplateID        string     `db:"template_id"`
	EmployeeID        string     `db:"employee_id"`
	Number            string     `db:"number"`
	StartDate         *time.Time `db:"start_date"`
	EndDate           *time.Time `db:"end_date"`
	Salary            string     `db:"salary"`
	DesignationID     *string    `db:"designation_id"`
	DesignationTitle  string     `db:"designation_title"`
	Status            string     `db:"status"`
	ContractType      string     `db:"contract_type"`
	Data              []byte     `db:"data"`
	Templates         []byte     `db:"templates"`
	SentAt            *time.Time `db:"sent_at"`
	DocumentID        *string    `db:"document_id"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type ContractListRow struct {
	ID                string     `db:"id"`
	TemplateID        string     `db:"template_id"`
	EmployeeID        string     `db:"employee_id"`
	Number            string     `db:"number"`
	StartDate         *time.Time `db:"start_date"`
	EndDate           *time.Time `db:"end_date"`
	Salary            string     `db:"salary"`
	DesignationID     *string    `db:"designation_id"`
	DesignationTitle  string     `db:"designation_title"`
	Status            string     `db:"status"`
	ContractType      string     `db:"contract_type"`
	Data              []byte     `db:"data"`
	Templates         []byte     `db:"templates"`
	SentAt            *time.Time `db:"sent_at"`
	DocumentID        *string    `db:"document_id"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type NumberSequenceModel struct {
	DesignationCode string    `db:"designation_code"`
	Prefix          string    `db:"prefix"`
	LastSequence    int       `db:"last_sequence"`
	UpdatedAt       time.Time `db:"updated_at"`
}

const (
	queryInsertContract = `
		INSERT INTO contracts (id, template_id, employee_id, number,
			start_date, end_date,
			salary, designation_id, designation_title,
			status, data, templates, sent_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`
	querySelectContract = `
		SELECT id, template_id, employee_id, number,
		       start_date, end_date,
		       salary, designation_id, designation_title,
		       status, data, templates, sent_at, created_at, updated_at
		FROM contracts
	`
	queryUpdateContract = `UPDATE contracts SET status=$1, sent_at=$2, updated_at=$3 WHERE id=$4`

	querySelectNumberSequenceByCode          = `SELECT designation_code, prefix, last_sequence, updated_at FROM number_sequences WHERE designation_code=$1`
	querySelectNumberSequenceByCodeForUpdate = `SELECT designation_code, prefix, last_sequence, updated_at FROM number_sequences WHERE designation_code=$1 FOR UPDATE`
	queryInsertNumberSequenceOnConflict      = `INSERT INTO number_sequences (designation_code, prefix, last_sequence, updated_at) VALUES ($1,$2,$3,NOW()) ON CONFLICT (designation_code) DO UPDATE SET last_sequence=EXCLUDED.last_sequence, updated_at=NOW()`
	queryUpdateNumberSequence                = `UPDATE number_sequences SET last_sequence=$1, updated_at=NOW() WHERE designation_code=$2`
	querySetMinimumSequence                  = `UPDATE number_sequences SET last_sequence=GREATEST(last_sequence,$1), updated_at=NOW() WHERE designation_code=$2`
)

var (
	queryContractByID = querySelectContractWithType + ` WHERE c.id=$1`
)

const querySelectContractWithType = `
	SELECT c.id, c.template_id, c.employee_id, c.number,
	       c.start_date, c.end_date,
	       c.salary, c.designation_id, c.designation_title,
	       c.status, ct.contract_type, c.data, c.templates,
	       c.sent_at, cd.document_id, c.created_at, c.updated_at
	FROM contracts c
	LEFT JOIN contract_templates ct ON ct.id = c.template_id
	LEFT JOIN contract_documents cd ON cd.contract_id = c.id
`

type PostgresContractRepo struct {
	db txContext
}

func NewPostgresContractRepo(db *sqlx.DB) *PostgresContractRepo {
	return &PostgresContractRepo{db: db}
}

func (r *PostgresContractRepo) WithTx(tx *sqlx.Tx) ContractRepository {
	return &PostgresContractRepo{db: tx}
}

func (r *PostgresContractRepo) CreateContract(ctx context.Context, e *entity.Contract) error {
	dataJSON, err := json.Marshal(e.Data)
	if err != nil {
		return fmt.Errorf("marshal contract data: %w", err)
	}
	templatesJSON, err := json.Marshal(e.Templates)
	if err != nil {
		return fmt.Errorf("marshal contract templates: %w", err)
	}
	_, err = r.db.ExecContext(ctx, queryInsertContract,
		e.ID, e.TemplateID, e.EmployeeID, e.Number,
		e.StartDate, e.EndDate,
		e.Salary, e.DesignationID, e.DesignationTitle,
		string(e.Status), dataJSON, templatesJSON, e.SentAt, e.CreatedAt, e.UpdatedAt,
	)
	return err
}

func (r *PostgresContractRepo) BulkCreateContracts(ctx context.Context, contracts []*entity.Contract) error {
	if len(contracts) == 0 {
		return nil
	}

	query := `INSERT INTO contracts (id, template_id, employee_id, number,
		start_date, end_date,
		salary, designation_id, designation_title,
		status, data, templates, sent_at, created_at, updated_at) VALUES `
	args := make([]interface{}, 0, 15*len(contracts))
	paramIdx := 1

	for _, e := range contracts {
		if paramIdx > 1 {
			query += ", "
		}
		query += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			paramIdx, paramIdx+1, paramIdx+2, paramIdx+3, paramIdx+4, paramIdx+5, paramIdx+6, paramIdx+7, paramIdx+8,
			paramIdx+9, paramIdx+10, paramIdx+11, paramIdx+12, paramIdx+13, paramIdx+14)
		paramIdx += 15

		dataJSON, err := json.Marshal(e.Data)
		if err != nil {
			return fmt.Errorf("marshal contract data: %w", err)
		}
		templatesJSON, err := json.Marshal(e.Templates)
		if err != nil {
			return fmt.Errorf("marshal contract templates: %w", err)
		}
		args = append(args, e.ID, e.TemplateID, e.EmployeeID, e.Number,
			e.StartDate, e.EndDate,
			e.Salary, e.DesignationID, e.DesignationTitle,
			string(e.Status), dataJSON, templatesJSON, e.SentAt, e.CreatedAt, e.UpdatedAt)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresContractRepo) FindContractByID(ctx context.Context, id string) (*entity.Contract, error) {
	var m ContractModel
	err := r.db.QueryRowxContext(ctx, queryContractByID, id).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find contract: %w", err)
	}
	return modelToContract(&m), nil
}

func (r *PostgresContractRepo) FindActiveByEmployeeID(ctx context.Context, employeeID string) (*entity.Contract, error) {
	var m ContractModel
	err := r.db.QueryRowxContext(ctx, querySelectContractWithType+` WHERE c.employee_id = $1 AND c.status = 'active' ORDER BY c.created_at DESC LIMIT 1`, employeeID).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find active contract by employee: %w", err)
	}
	return modelToContract(&m), nil
}

func (r *PostgresContractRepo) FindCurrentByEmployeeID(ctx context.Context, employeeID string) (*entity.Contract, error) {
	var m ContractModel
	err := r.db.QueryRowxContext(ctx, querySelectContractWithType+` WHERE c.employee_id = $1 ORDER BY c.created_at DESC LIMIT 1`, employeeID).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find current contract by employee: %w", err)
	}
	return modelToContract(&m), nil
}

func (r *PostgresContractRepo) FindActiveContractEmployeeIDs(ctx context.Context, employeeIDs []string) (map[string]*time.Time, error) {
	if len(employeeIDs) == 0 {
		return map[string]*time.Time{}, nil
	}
	query, args, err := sqlx.In(`SELECT employee_id, end_date FROM contracts WHERE employee_id IN (?) AND status='active'`, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find active contracts: %w", err)
	}
	defer rows.Close()
	result := make(map[string]*time.Time)
	for rows.Next() {
		var id string
		var endDate *time.Time
		if err := rows.Scan(&id, &endDate); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		result[id] = endDate
	}
	return result, rows.Err()
}

func (r *PostgresContractRepo) FindAllContracts(ctx context.Context, filter models.ListContractInput) ([]*entity.Contract, int64, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argIdx := 1
	if filter.Status != "" {
		where += fmt.Sprintf(" AND c.status=$%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.DesignationID != "" {
		where += fmt.Sprintf(" AND c.designation_id=$%d", argIdx)
		args = append(args, filter.DesignationID)
		argIdx++
	}
	if filter.ContractType != "" {
		where += fmt.Sprintf(" AND ct.contract_type=$%d", argIdx)
		args = append(args, filter.ContractType)
		argIdx++
	}
	if filter.EmployeeID != "" {
		where += fmt.Sprintf(" AND c.employee_id=$%d", argIdx)
		args = append(args, filter.EmployeeID)
		argIdx++
	}
	if filter.ExcludeDraft {
		where += " AND c.status != 'draft'"
	}

	countQuery := `SELECT COUNT(*) FROM contracts c LEFT JOIN contract_templates ct ON ct.id=c.template_id` + where
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count contracts: %w", err)
	}

	page, perPage := filter.Page, filter.PerPage
	if page < 1 { page = 1 }
	if perPage < 1 || perPage > 100 { perPage = 20 }
	offset := (page - 1) * perPage

	dataQuery := querySelectContractWithType + where + fmt.Sprintf(" ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query contracts: %w", err)
	}
	defer rows.Close()

	var list []*entity.Contract
	for rows.Next() {
		var m ContractListRow
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("scan contract: %w", err)
		}
		list = append(list, rowToContract(&m))
	}
	return list, total, rows.Err()
}

func (r *PostgresContractRepo) UpdateContract(ctx context.Context, e *entity.Contract) error {
	result, err := r.db.ExecContext(ctx, queryUpdateContract, string(e.Status), e.SentAt, e.UpdatedAt, e.ID)
	if err != nil {
		return fmt.Errorf("update contract: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return nil
	}
	return nil
}

func (r *PostgresContractRepo) CountByEmployeeIDAndStatus(ctx context.Context, employeeID string, status string) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM contracts WHERE employee_id=$1 AND status=$2`, employeeID, status)
	if err != nil {
		return 0, fmt.Errorf("count contracts: %w", err)
	}
	return count, nil
}

func modelToContract(m *ContractModel) *entity.Contract {
	var data entity.ContractTemplateData
	if len(m.Data) > 0 {
		_ = json.Unmarshal(m.Data, &data)
	}
	var templates entity.ContractTemplatePartials
	if len(m.Templates) > 0 {
		_ = json.Unmarshal(m.Templates, &templates)
	}
	c := entity.ReconstituteContract(m.ID, m.TemplateID, m.EmployeeID, m.Number,
		m.StartDate, m.EndDate, m.Salary,
		m.DesignationID, m.DesignationTitle,
		entity.ContractStatus(m.Status), m.ContractType, data, templates,
		m.SentAt, m.CreatedAt, m.UpdatedAt)
	c.DocumentID = m.DocumentID
	return c
}

func rowToContract(r *ContractListRow) *entity.Contract {
	var data entity.ContractTemplateData
	if len(r.Data) > 0 {
		json.Unmarshal(r.Data, &data)
	}
	var templates entity.ContractTemplatePartials
	if len(r.Templates) > 0 {
		json.Unmarshal(r.Templates, &templates)
	}
	c := entity.ReconstituteContract(r.ID, r.TemplateID, r.EmployeeID, r.Number,
		r.StartDate, r.EndDate, r.Salary,
		r.DesignationID, r.DesignationTitle,
		entity.ContractStatus(r.Status), r.ContractType, data, templates,
		r.SentAt, r.CreatedAt, r.UpdatedAt)
	c.DocumentID = r.DocumentID
	return c
}

// ---- Number Sequence (implements numbergen.SequenceRepository) ----

func (r *PostgresContractRepo) GetCurrent(ctx context.Context, designationCode string) (*numbergen.Sequence, error) {
	var m NumberSequenceModel
	err := r.db.QueryRowxContext(ctx, querySelectNumberSequenceByCode, designationCode).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sequence: %w", err)
	}
	return &numbergen.Sequence{DesignationCode: m.DesignationCode, Prefix: m.Prefix, LastSequence: m.LastSequence}, nil
}

func (r *PostgresContractRepo) GetForUpdate(ctx context.Context, designationCode string) (*numbergen.Sequence, error) {
	var m NumberSequenceModel
	err := r.db.QueryRowxContext(ctx, querySelectNumberSequenceByCodeForUpdate, designationCode).StructScan(&m)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sequence for update: %w", err)
	}
	return &numbergen.Sequence{DesignationCode: m.DesignationCode, Prefix: m.Prefix, LastSequence: m.LastSequence}, nil
}

func (r *PostgresContractRepo) CreateSequence(ctx context.Context, designationCode, prefix string, lastSequence int) error {
	_, err := r.db.ExecContext(ctx, queryInsertNumberSequenceOnConflict, designationCode, prefix, lastSequence)
	return err
}

func (r *PostgresContractRepo) Increment(ctx context.Context, designationCode string, nextSequence int) error {
	result, err := r.db.ExecContext(ctx, queryUpdateNumberSequence, nextSequence, designationCode)
	if err != nil {
		return fmt.Errorf("increment sequence: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("sequence not found for %s", designationCode)
	}
	return nil
}

func (r *PostgresContractRepo) SetMinimumSequence(ctx context.Context, designationCode string, minSequence int) error {
	_, err := r.db.ExecContext(ctx, querySetMinimumSequence, minSequence, designationCode)
	return err
}

func (r *PostgresContractRepo) FindActiveContractsPastEndDate(ctx context.Context, asOf time.Time) ([]*entity.Contract, error) {
	rows, err := r.db.QueryxContext(ctx,
		`SELECT c.id, c.template_id, c.employee_id, c.number,
		        c.start_date, c.end_date,
		        c.salary, c.designation_id, c.designation_title,
		        c.status, c.data, c.templates, c.sent_at, c.created_at, c.updated_at
		 FROM contracts c
		 WHERE c.status = 'active'
		   AND c.end_date IS NOT NULL
		   AND c.end_date < $1`, asOf)
	if err != nil {
		return nil, fmt.Errorf("find expired active contracts: %w", err)
	}
	defer rows.Close()

	var result []*entity.Contract
	for rows.Next() {
		var m ContractModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		result = append(result, modelToContract(&m))
	}
	return result, rows.Err()
}

func (r *PostgresContractRepo) CountSoonExpired(ctx context.Context, withinDays int) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM contracts
		 WHERE status = 'active'
		   AND end_date IS NOT NULL
		   AND end_date >= CURRENT_DATE
		   AND end_date < CURRENT_DATE + $1::int`,
		withinDays,
	)
	if err != nil {
		return 0, fmt.Errorf("count soon expired: %w", err)
	}
	return count, nil
}

func (r *PostgresContractRepo) HasOtherActiveContract(ctx context.Context, employeeID, excludeContractID string) (bool, error) {
	var count int64
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM contracts
		 WHERE employee_id = $1 AND status = 'active' AND id != $2`,
		employeeID, excludeContractID)
	if err != nil {
		return false, fmt.Errorf("count other active contracts: %w", err)
	}
	return count > 0, nil
}

func (r *PostgresContractRepo) DeleteContract(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM contracts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete contract: %w", err)
	}
	return nil
}
