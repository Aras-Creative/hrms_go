package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"hrms/internal/auth/entity"
)

// Named query constants for sessions.
const (
	queryInsertSession = `
		INSERT INTO sessions (id, user_id, device_id, refresh_token, is_active, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	querySelectSession = `
		SELECT id, user_id, device_id, refresh_token, is_active, expires_at, created_at
		FROM sessions
	`

	queryUpdateSession = `
		UPDATE sessions
		SET device_id = $1, refresh_token = $2, is_active = $3, expires_at = $4
		WHERE id = $5
	`
)

var (
	querySessionByRefreshToken = querySelectSession + ` WHERE refresh_token = $1 AND is_active = true`
	queryActiveByUserID        = querySelectSession + ` WHERE user_id = $1 AND is_active = true`
)

type PostgresSessionRepo struct {
	db *sqlx.DB
}

func NewPostgresSessionRepo(db *sqlx.DB) *PostgresSessionRepo {
	return &PostgresSessionRepo{db: db}
}

func (r *PostgresSessionRepo) CreateSession(ctx context.Context, session *entity.Session) error {
	var deviceID *string
	if session.DeviceID != "" {
		deviceID = &session.DeviceID
	}

	_, err := r.db.ExecContext(ctx, queryInsertSession,
		session.ID,
		session.UserID,
		deviceID,
		session.RefreshToken,
		session.IsActive,
		session.ExpiresAt,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepo) UpdateSession(ctx context.Context, session *entity.Session) error {
	var deviceID *string
	if session.DeviceID != "" {
		deviceID = &session.DeviceID
	}

	_, err := r.db.ExecContext(ctx, queryUpdateSession,
		deviceID,
		session.RefreshToken,
		session.IsActive,
		session.ExpiresAt,
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepo) FindByRefreshToken(ctx context.Context, refreshToken string) (*entity.Session, error) {
	var sess entity.Session
	var deviceID *string
	err := r.db.QueryRowxContext(ctx, querySessionByRefreshToken, refreshToken).Scan(
		&sess.ID,
		&sess.UserID,
		&deviceID,
		&sess.RefreshToken,
		&sess.IsActive,
		&sess.ExpiresAt,
		&sess.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	if deviceID != nil {
		sess.DeviceID = *deviceID
	}
	return &sess, nil
}

func (r *PostgresSessionRepo) FindActiveByUserID(ctx context.Context, userID string) ([]*entity.Session, error) {
	rows, err := r.db.QueryxContext(ctx, queryActiveByUserID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*entity.Session
	for rows.Next() {
		var sess entity.Session
		var deviceID *string
		if err := rows.Scan(
			&sess.ID,
			&sess.UserID,
			&deviceID,
			&sess.RefreshToken,
			&sess.IsActive,
			&sess.ExpiresAt,
			&sess.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		if deviceID != nil {
			sess.DeviceID = *deviceID
		}
		sessions = append(sessions, &sess)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return sessions, nil
}

func (r *PostgresSessionRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE sessions SET is_active = false WHERE user_id = $1 AND is_active = true`, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE is_active = false AND expires_at < NOW()`)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}
