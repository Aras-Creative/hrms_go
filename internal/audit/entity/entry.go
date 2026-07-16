package entity

import (
	"encoding/json"
	"time"
)

type AuditEntry struct {
	ID         string
	Action     string
	ActorID    string
	Resource   string
	ResourceID string
	TargetID   *string
	Payload    map[string]any
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
}

func NewAuditEntry(action, actorID, resource, resourceID string, targetID *string, payload map[string]any, ip, userAgent string) *AuditEntry {
	now := time.Now()
	return &AuditEntry{
		ID:         "",
		Action:     action,
		ActorID:    actorID,
		Resource:   resource,
		ResourceID: resourceID,
		TargetID:   targetID,
		Payload:    payload,
		IPAddress:  ip,
		UserAgent:  userAgent,
		CreatedAt:  now,
	}
}

func (e *AuditEntry) PayloadJSON() []byte {
	if e.Payload == nil {
		return nil
	}
	b, _ := json.Marshal(e.Payload)
	return b
}
