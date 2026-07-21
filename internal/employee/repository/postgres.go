package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/employee/entity"
	"hrms/internal/employee/models"
	"hrms/internal/employee/numbergen"
)

const (
	queryInsertEmployee = `
		INSERT INTO employees (
			id, user_id, full_name, employee_number, phone, personal_email,
			emergency_contact_name, emergency_contact_phone, place_of_birth,
			date_of_birth, join_date, gender, education, status, address,
			designation_id, national_id, religion, profile_photo_id,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22
		)
	`

	querySelectEmployee = `
		SELECT
			id, user_id, full_name, employee_number, phone, personal_email,
			emergency_contact_name, emergency_contact_phone, place_of_birth,
			date_of_birth, join_date, gender, education, status, address,
			designation_id, national_id, religion, profile_photo_id,
			bank_holder, bank_name, bank_number,
			is_active, termination_date, created_at, updated_at
		FROM employees
	`

	queryUpdateEmployee = `
		UPDATE employees SET
			user_id = $1, full_name = $2, employee_number = $3, phone = $4,
			personal_email = $5, emergency_contact_name = $6,
			emergency_contact_phone = $7, place_of_birth = $8,
			date_of_birth = $9, join_date = $10, gender = $11,
			education = $12, status = $13, address = $14,
			designation_id = $15, national_id = $16, religion = $17,
			profile_photo_id = $18, is_active = $19, updated_at = $20,
			bank_holder = $21, bank_name = $22, bank_number = $23
		WHERE id = $24
	`

	queryDeleteEmployee = `DELETE FROM employees WHERE id = $1`

	querySelectNumberSequence = `
		SELECT designation_code, prefix, last_sequence, updated_at
		FROM number_sequences
	`

	queryInsertNumberSequence = `
		INSERT INTO number_sequences (designation_code, prefix, last_sequence, updated_at)
		VALUES ($1, $2, $3, NOW())
	`

	queryUpdateNumberSequence = `
		UPDATE number_sequences SET last_sequence = $1, updated_at = NOW()
		WHERE designation_code = $2
	`
)

const querySelectEmployeeWithDetails = `
	SELECT DISTINCT ON (e.id)
		e.id, e.user_id, e.full_name, e.employee_number, e.phone, e.personal_email,
		e.emergency_contact_name, e.emergency_contact_phone, e.place_of_birth,
		e.date_of_birth, e.join_date, e.gender, e.education, e.status, e.address,
		e.designation_id, e.national_id, e.religion, e.profile_photo_id,
		e.bank_holder, e.bank_name, e.bank_number,
		e.is_active, e.termination_date, e.created_at, e.updated_at,
		d.name AS designation_name,
		dv.id AS device_id,
		dv.platform AS device_platform,
		dv.user_agent AS device_name,
		dv.is_active AS device_is_active,
		dv.last_used_at AS device_last_used,
		dv.created_at AS device_created_at
	FROM employees e
	LEFT JOIN designations d ON d.id = e.designation_id
	LEFT JOIN devices dv ON dv.user_id = e.user_id
`

var (
	queryEmployeeByID                   = querySelectEmployee + ` WHERE id = $1`
	queryEmployeeByIDWithDetails        = querySelectEmployeeWithDetails + ` WHERE e.id = $1`
	queryEmployeeByUserID               = querySelectEmployee + ` WHERE user_id = $1`
	queryEmployeeByUserIDWithDetails    = querySelectEmployeeWithDetails + ` WHERE e.user_id = $1`
	queryEmployeeAllWithDetailsBase     = querySelectEmployeeWithDetails
	queryNumberSequenceByCode           = querySelectNumberSequence + ` WHERE designation_code = $1`
	queryNumberSequenceByCodeForUpdate  = querySelectNumberSequence + ` WHERE designation_code = $1 FOR UPDATE`
	queryInsertNumberSequenceOnConflict = queryInsertNumberSequence + ` ON CONFLICT (designation_code) DO UPDATE SET last_sequence = EXCLUDED.last_sequence, updated_at = NOW()`
)

type PostgresEmployeeRepo struct {
	db *sqlx.DB
}

func NewPostgresEmployeeRepo(db *sqlx.DB) *PostgresEmployeeRepo {
	return &PostgresEmployeeRepo{db: db}
}

func (r *PostgresEmployeeRepo) Create(ctx context.Context, e *entity.Employee) error {
	var personalEmail, dateOfBirth, joinDate *string
	if e.PersonalEmail != "" {
		personalEmail = &e.PersonalEmail
	}
	if e.DateOfBirth != nil {
		s := e.DateOfBirth.String()
		dateOfBirth = &s
	}
	if e.JoinDate != nil {
		s := e.JoinDate.String()
		joinDate = &s
	}

	_, err := r.db.ExecContext(ctx, queryInsertEmployee,
		e.ID,
		e.UserID,
		e.FullName,
		e.EmployeeNumber.String(),
		e.Phone.String(),
		personalEmail,
		e.EmergencyContactName,
		e.EmergencyContactPhone.String(),
		e.PlaceOfBirth,
		dateOfBirth,
		joinDate,
		string(e.Gender),
		e.Education,
		string(e.Status),
		e.Address,
		e.DesignationID,
		e.NationalID,
		string(e.Religion),
		e.ProfilePhotoID,
		e.IsActive,
		e.CreatedAt,
		e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}
	return nil
}

func (r *PostgresEmployeeRepo) FindByID(ctx context.Context, id string) (*entity.Employee, error) {
	var m EmployeeModel
	err := r.db.QueryRowxContext(ctx, queryEmployeeByID, id).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find employee: %w", err)
	}
	return modelToEntity(&m), nil
}

func (r *PostgresEmployeeRepo) FindByIDs(ctx context.Context, ids []string) ([]*entity.Employee, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(querySelectEmployee+` WHERE id IN (?)`, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	query = r.db.Rebind(query)
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find employees by ids: %w", err)
	}
	defer rows.Close()

	var list []*entity.Employee
	for rows.Next() {
		var m EmployeeModel
		if err := rows.StructScan(&m); err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}
		list = append(list, modelToEntity(&m))
	}
	return list, rows.Err()
}

func (r *PostgresEmployeeRepo) FindByIDWithDetails(ctx context.Context, id string) (*models.EmployeeResult, error) {
	var m EmployeeWithDetailsRow
	err := r.db.QueryRowxContext(ctx, queryEmployeeByIDWithDetails, id).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find employee with details: %w", err)
	}
	return toEmployeeResult(&m), nil
}

func (r *PostgresEmployeeRepo) FindByUserID(ctx context.Context, userID string) (*entity.Employee, error) {
	var m EmployeeModel
	err := r.db.QueryRowxContext(ctx, queryEmployeeByUserID, userID).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find employee by user id: %w", err)
	}
	return modelToEntity(&m), nil
}

func (r *PostgresEmployeeRepo) FindByUserIDWithDetails(ctx context.Context, userID string) (*models.MeResult, error) {
	var m EmployeeWithDetailsRow
	err := r.db.QueryRowxContext(ctx, queryEmployeeByUserIDWithDetails, userID).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find employee with designation: %w", err)
	}
	return rowToMeResult(&m), nil
}

func (r *PostgresEmployeeRepo) FindAllActiveIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryxContext(ctx, `SELECT id FROM employees WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to list active employee IDs: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan employee ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PostgresEmployeeRepo) FindAllWithDetails(ctx context.Context, filter models.ListEmployeeInput) ([]*models.EmployeeListItem, int64, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.SearchName != "" {
		where += fmt.Sprintf(" AND e.full_name ILIKE $%d", argIdx)
		args = append(args, "%"+filter.SearchName+"%")
		argIdx++
	}
	if filter.Status != "" {
		where += fmt.Sprintf(" AND e.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Gender != "" {
		where += fmt.Sprintf(" AND e.gender = $%d", argIdx)
		args = append(args, filter.Gender)
		argIdx++
	}
	if filter.DesignationID != "" {
		where += fmt.Sprintf(" AND e.designation_id = $%d", argIdx)
		args = append(args, filter.DesignationID)
		argIdx++
	}

	countQuery := `SELECT COUNT(*) FROM employees e` + where
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := queryEmployeeAllWithDetailsBase + where + fmt.Sprintf(" ORDER BY e.full_name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list employees with designation: %w", err)
	}
	defer rows.Close()

	var list []*models.EmployeeListItem
	for rows.Next() {
		var m EmployeeWithDetailsRow
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee with designation: %w", err)
		}
		list = append(list, rowToEmployeeListItem(&m))
	}
	return list, total, rows.Err()
}

func (r *PostgresEmployeeRepo) Update(ctx context.Context, e *entity.Employee) error {
	var personalEmail, dateOfBirth, joinDate *string
	if e.PersonalEmail != "" {
		personalEmail = &e.PersonalEmail
	}
	if e.DateOfBirth != nil {
		s := e.DateOfBirth.String()
		dateOfBirth = &s
	}
	if e.JoinDate != nil {
		s := e.JoinDate.String()
		joinDate = &s
	}

	result, err := r.db.ExecContext(ctx, queryUpdateEmployee,
		e.UserID,
		e.FullName,
		e.EmployeeNumber.String(),
		e.Phone.String(),
		personalEmail,
		e.EmergencyContactName,
		e.EmergencyContactPhone.String(),
		e.PlaceOfBirth,
		dateOfBirth,
		joinDate,
		string(e.Gender),
		e.Education,
		string(e.Status),
		e.Address,
		e.DesignationID,
		e.NationalID,
		string(e.Religion),
		e.ProfilePhotoID,
		e.IsActive,
		e.UpdatedAt,
		e.BankAccount.Holder(),
		e.BankAccount.Name(),
		e.BankAccount.Number(),
		e.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update employee: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("employee with id %s not found", e.ID)
	}
	return nil
}

func (r *PostgresEmployeeRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, queryDeleteEmployee, id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("employee with id %s not found", id)
	}
	return nil
}

func (r *PostgresEmployeeRepo) GetCurrent(ctx context.Context, designationCode string) (*numbergen.Sequence, error) {
	var m NumberSequenceModel
	err := r.db.QueryRowxContext(ctx, queryNumberSequenceByCode, designationCode).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current sequence: %w", err)
	}
	return &numbergen.Sequence{
		DesignationCode: m.DesignationCode,
		Prefix:          m.Prefix,
		LastSequence:    m.LastSequence,
	}, nil
}

func (r *PostgresEmployeeRepo) GetForUpdate(ctx context.Context, designationCode string) (*numbergen.Sequence, error) {
	var m NumberSequenceModel
	err := r.db.QueryRowxContext(ctx, queryNumberSequenceByCodeForUpdate, designationCode).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sequence for update: %w", err)
	}
	return &numbergen.Sequence{
		DesignationCode: m.DesignationCode,
		Prefix:          m.Prefix,
		LastSequence:    m.LastSequence,
	}, nil
}

func (r *PostgresEmployeeRepo) CreateSequence(ctx context.Context, designationCode, prefix string, lastSequence int) error {
	_, err := r.db.ExecContext(ctx, queryInsertNumberSequenceOnConflict, designationCode, prefix, lastSequence)
	if err != nil {
		return fmt.Errorf("failed to create number sequence: %w", err)
	}
	return nil
}

func (r *PostgresEmployeeRepo) Increment(ctx context.Context, designationCode string, nextSequence int) error {
	result, err := r.db.ExecContext(ctx, queryUpdateNumberSequence, nextSequence, designationCode)
	if err != nil {
		return fmt.Errorf("failed to increment sequence: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("number sequence not found for %s", designationCode)
	}
	return nil
}

func (r *PostgresEmployeeRepo) SetMinimumSequence(ctx context.Context, designationCode string, minSequence int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE number_sequences SET last_sequence = GREATEST(last_sequence, $1), updated_at = NOW() WHERE designation_code = $2`, minSequence, designationCode)
	if err != nil {
		return fmt.Errorf("failed to set minimum sequence: %w", err)
	}
	return nil
}

func modelToEntity(m *EmployeeModel) *entity.Employee {
	var dateOfBirth *entity.Date
	if m.DateOfBirth != nil {
		d, _ := entity.ParseDate(*m.DateOfBirth)
		dateOfBirth = &d
	}
	var joinDate *entity.Date
	if m.JoinDate != nil {
		d, _ := entity.ParseDate(*m.JoinDate)
		joinDate = &d
	}

	e := entity.ReconstituteEmployee(
		m.ID,
		m.UserID,
		m.FullName,
		entity.FromString(m.EmployeeNumber),
		entity.PhoneFromDB(m.Phone),
		coalesceStr(m.PersonalEmail),
		m.EmergencyContactName,
		entity.PhoneFromDB(m.EmergencyContactPhone),
		m.PlaceOfBirth,
		dateOfBirth,
		joinDate,
		entity.Gender(m.Gender),
		m.Education,
		entity.Status(m.Status),
		m.Address,
		m.DesignationID,
		m.NationalID,
		entity.Religion(m.Religion),
		m.ProfilePhotoID,
		entity.BankAccountFromDB(m.BankHolder, m.BankName, m.BankNumber),
		m.IsActive,
		m.CreatedAt,
		m.UpdatedAt,
	)
	e.TerminationDate = m.TerminationDate
	return e
}

func coalesceStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func coalesceStrPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func rowToMeResult(r *EmployeeWithDetailsRow) *models.MeResult {
	result := &models.MeResult{
		ID:                    r.ID,
		UserID:                r.UserID,
		FullName:              r.FullName,
		EmployeeNumber:        r.EmployeeNumber,
		Phone:                 r.Phone,
		PersonalEmail:         coalesceStrPtr(r.PersonalEmail),
		EmergencyContactName:  r.EmergencyContactName,
		EmergencyContactPhone: r.EmergencyContactPhone,
		PlaceOfBirth:          r.PlaceOfBirth,
		DateOfBirth:           r.DateOfBirth,
		JoinDate:              r.JoinDate,
		Gender:                r.Gender,
		Education:             r.Education,
		Status:                r.Status,
		Address:               r.Address,
		DesignationID:         r.DesignationID,
		DesignationName:       r.DesignationName,
		NationalID:            r.NationalID,
		Religion:              r.Religion,
		ProfilePhotoID:        r.ProfilePhotoID,
		IsActive:              r.IsActive,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
		BankHolder:            r.BankHolder,
		BankName:              r.BankName,
		BankNumber:            r.BankNumber,
	}

	if r.DeviceID != nil {
		result.Device = &models.DeviceInfo{
			ID:         *r.DeviceID,
			Platform:   coalesceStrPtr(r.DevicePlatform),
			Name:       coalesceStrPtr(r.DeviceName),
			IsActive:   r.DeviceIsActive != nil && *r.DeviceIsActive,
			LastUsedAt: derefTime(r.DeviceLastUsed),
			CreatedAt:  derefTime(r.DeviceCreatedAt),
		}
	}

	return result
}

func rowToEmployeeListItem(r *EmployeeWithDetailsRow) *models.EmployeeListItem {
	return &models.EmployeeListItem{
		ID:                    r.ID,
		FullName:              r.FullName,
		EmployeeNumber:        r.EmployeeNumber,
		Phone:                 r.Phone,
		PersonalEmail:         coalesceStrPtr(r.PersonalEmail),
		EmergencyContactName:  r.EmergencyContactName,
		EmergencyContactPhone: r.EmergencyContactPhone,
		JoinDate:              r.JoinDate,
		Gender:                r.Gender,
		Status:                r.Status,
		DesignationID:         r.DesignationID,
		DesignationName:       r.DesignationName,
		ProfilePhotoID:        r.ProfilePhotoID,
		IsActive:              r.IsActive,
	}
}
