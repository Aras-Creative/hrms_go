package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"hrms/internal/audit/entity"
	"hrms/internal/audit/models"
)

const (
	queryInsertAuditLog = `
		INSERT INTO audit_logs (id, action, actor_id, resource, resource_id, target_id, payload, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	querySelectAuditLogs = `
		SELECT id, action, actor_id, resource, resource_id, target_id, payload, ip_address, user_agent, created_at
		FROM audit_logs
	`

	queryCountAuditLogs = `SELECT COUNT(*) FROM audit_logs`
)

type AuditLogModel struct {
	ID         string          `db:"id"`
	Action     string          `db:"action"`
	ActorID    string          `db:"actor_id"`
	Resource   string          `db:"resource"`
	ResourceID string          `db:"resource_id"`
	TargetID   *string         `db:"target_id"`
	Payload    json.RawMessage `db:"payload"`
	IPAddress  string          `db:"ip_address"`
	UserAgent  string          `db:"user_agent"`
	CreatedAt  time.Time       `db:"created_at"`
}

type PostgresAuditRepo struct {
	db *sqlx.DB
}

func NewPostgresAuditRepo(db *sqlx.DB) *PostgresAuditRepo {
	return &PostgresAuditRepo{db: db}
}

func (r *PostgresAuditRepo) Create(ctx context.Context, e *entity.AuditEntry) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}

	var payload []byte
	if e.Payload != nil {
		var err error
		payload, err = json.Marshal(e.Payload)
		if err != nil {
			return fmt.Errorf("marshal audit payload: %w", err)
		}
	}

	_, err := r.db.ExecContext(ctx, queryInsertAuditLog,
		e.ID, e.Action, e.ActorID, e.Resource, e.ResourceID,
		e.TargetID, payload, e.IPAddress, e.UserAgent, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func (r *PostgresAuditRepo) List(ctx context.Context, filter models.AuditFilter) ([]*entity.AuditEntry, int64, error) {
	where := " WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.Action != "" {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, filter.Action)
		argIdx++
	}
	if filter.ActorID != "" {
		where += fmt.Sprintf(" AND actor_id = $%d", argIdx)
		args = append(args, filter.ActorID)
		argIdx++
	}
	if filter.Resource != "" {
		where += fmt.Sprintf(" AND resource = $%d", argIdx)
		args = append(args, filter.Resource)
		argIdx++
	}
	if filter.ResourceID != "" {
		where += fmt.Sprintf(" AND resource_id = $%d", argIdx)
		args = append(args, filter.ResourceID)
		argIdx++
	}
	if filter.DateFrom != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *filter.DateTo)
		argIdx++
	}

	var total int64
	countQuery := queryCountAuditLogs + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
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

	dataQuery := querySelectAuditLogs + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var list []*entity.AuditEntry
	for rows.Next() {
		var m AuditLogModel
		if err := rows.StructScan(&m); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		list = append(list, modelToEntry(&m))
	}
	return list, total, rows.Err()
}

func modelToEntry(m *AuditLogModel) *entity.AuditEntry {
	var payload map[string]any
	if len(m.Payload) > 0 {
		json.Unmarshal(m.Payload, &payload)
	}
	return &entity.AuditEntry{
		ID:         m.ID,
		Action:     m.Action,
		ActorID:    m.ActorID,
		Resource:   m.Resource,
		ResourceID: m.ResourceID,
		TargetID:   m.TargetID,
		Payload:    payload,
		IPAddress:  m.IPAddress,
		UserAgent:  m.UserAgent,
		CreatedAt:  m.CreatedAt,
	}
}

func (r *PostgresAuditRepo) ListByResourceWithActor(ctx context.Context, resource, resourceID string) ([]*models.AuditEntryWithActor, error) {
	type auditEntryRow struct {
		ID         string          `db:"id"`
		Action     string          `db:"action"`
		ActorID    string          `db:"actor_id"`
		ActorName  string          `db:"actor_name"`
		Resource   string          `db:"resource"`
		ResourceID string          `db:"resource_id"`
		TargetID   *string         `db:"target_id"`
		Payload    json.RawMessage `db:"payload"`
		IPAddress  string          `db:"ip_address"`
		UserAgent  string          `db:"user_agent"`
		CreatedAt  time.Time       `db:"created_at"`
	}

	query := `
		SELECT a.id, a.action, a.actor_id, COALESCE(u.full_name, '') AS actor_name,
		       a.resource, a.resource_id, a.target_id, a.payload,
		       a.ip_address, a.user_agent, a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.actor_id
		WHERE a.resource = $1 AND a.resource_id = $2
		ORDER BY a.created_at DESC
	`
	rows, err := r.db.QueryxContext(ctx, query, resource, resourceID)
	if err != nil {
		return nil, fmt.Errorf("query audit logs with actor: %w", err)
	}
	defer rows.Close()

	var list []*models.AuditEntryWithActor
	for rows.Next() {
		var r auditEntryRow
		if err := rows.StructScan(&r); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		var payload map[string]any
		if len(r.Payload) > 0 {
			json.Unmarshal(r.Payload, &payload)
		}
		list = append(list, &models.AuditEntryWithActor{
			ID: r.ID, Action: r.Action, ActorID: r.ActorID, ActorName: r.ActorName,
			Resource: r.Resource, ResourceID: r.ResourceID, TargetID: r.TargetID,
			Payload: payload, IPAddress: r.IPAddress, UserAgent: r.UserAgent, CreatedAt: r.CreatedAt,
		})
	}
	return list, rows.Err()
}


