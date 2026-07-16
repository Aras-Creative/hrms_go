package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/notification/entity"
)

type notificationModel struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	Type       string    `db:"type"`
	Title      string    `db:"title"`
	Body       string    `db:"body"`
	Resource   string    `db:"resource"`
	ResourceID string    `db:"resource_id"`
	IsRead     bool      `db:"is_read"`
	CreatedAt  time.Time `db:"created_at"`
}

type PostgresNotificationRepo struct {
	db *sqlx.DB
}

func NewPostgresNotificationRepo(db *sqlx.DB) *PostgresNotificationRepo {
	return &PostgresNotificationRepo{db: db}
}

func (r *PostgresNotificationRepo) Create(ctx context.Context, n *entity.Notification) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, title, body, resource, resource_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, n.ID, n.UserID, n.Type, n.Title, n.Body, n.Resource, n.ResourceID, n.CreatedAt)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *PostgresNotificationRepo) List(ctx context.Context, userID string, page, perPage int) ([]*entity.Notification, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var models []notificationModel
	if err := r.db.SelectContext(ctx, &models, `
		SELECT id, user_id, type, title, body, resource, resource_id, is_read, created_at
		FROM notifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, perPage, offset); err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}

	items := make([]*entity.Notification, len(models))
	for i, m := range models {
		items[i] = entity.ReconstituteNotification(m.ID, m.UserID, m.Type, m.Title, m.Body, m.Resource, m.ResourceID, m.IsRead, m.CreatedAt)
	}
	return items, total, nil
}

func (r *PostgresNotificationRepo) UnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userID); err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return count, nil
}

func (r *PostgresNotificationRepo) MarkRead(ctx context.Context, id, userID string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresNotificationRepo) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`, userID)
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

func (r *PostgresNotificationRepo) DeleteOlderThan(ctx context.Context, ttl string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notifications WHERE created_at < NOW() - $1::interval`, ttl)
	if err != nil {
		return fmt.Errorf("delete old notifications: %w", err)
	}
	return nil
}
