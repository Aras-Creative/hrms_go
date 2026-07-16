package models

import "time"

type AuditFilter struct {
	Action     string
	ActorID    string
	Resource   string
	ResourceID string
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	PerPage    int
}

type AuditEntryWithActor struct {
	ID         string
	Action     string
	ActorID    string
	ActorName  string
	Resource   string
	ResourceID string
	TargetID   *string
	Payload    map[string]any
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
}

type AuditLogData struct {
	Action     string
	ActorID    string
	Resource   string
	ResourceID string
	TargetID   *string
	Payload    map[string]any
	IP         string
	UserAgent  string
}
