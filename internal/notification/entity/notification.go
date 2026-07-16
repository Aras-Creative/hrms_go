package entity

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

const (
	TypeEmployment = "employment"
	TypeLeave      = "leave"
	TypeSecurity   = "security"
	TypePayroll    = "payroll"
	TypeGeneral    = "general"
)

func NewNotification(userID, ntype, title, body, resource, resourceID string) *Notification {
	return &Notification{
		ID:         uuid.New().String(),
		UserID:     userID,
		Type:       ntype,
		Title:      title,
		Body:       body,
		Resource:   resource,
		ResourceID: resourceID,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}
}

func ReconstituteNotification(id, userID, ntype, title, body, resource, resourceID string, isRead bool, createdAt time.Time) *Notification {
	return &Notification{
		ID:         id,
		UserID:     userID,
		Type:       ntype,
		Title:      title,
		Body:       body,
		Resource:   resource,
		ResourceID: resourceID,
		IsRead:     isRead,
		CreatedAt:  createdAt,
	}
}
