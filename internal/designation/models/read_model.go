package models

import "time"

type MemberRow struct {
	ID             string
	FullName       string
	EmployeeNumber string
	ProfilePhotoID *string
	DesignationID  string
}

type DesignationReadModel struct {
	ID          string
	Code        string
	Name        string
	Members     []MemberRow
	MemberCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
