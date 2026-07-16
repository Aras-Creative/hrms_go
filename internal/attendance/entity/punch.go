package entity

import (
	"time"

	"github.com/google/uuid"
	"hrms/internal/pkg/timeutil"
)

type PunchType string

const (
	PunchIn  PunchType = "in"
	PunchOut PunchType = "out"
)

type Punch struct {
	ID         string
	EmployeeID string
	Type       PunchType
	Timestamp  time.Time
	Date       time.Time
	CreatedAt  time.Time
}

func LocalDate(t time.Time) time.Time {
	loc := timeutil.LoadDefaultLocation()
	local := t.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func NewPunch(employeeID string, punchType PunchType, timestamp time.Time) *Punch {
	return &Punch{
		ID:         uuid.New().String(),
		EmployeeID: employeeID,
		Type:       punchType,
		Timestamp:  timestamp,
		Date:       LocalDate(timestamp),
		CreatedAt:  time.Now(),
	}
}

func ReconstitutePunch(id, employeeID string, punchType PunchType, timestamp time.Time, date time.Time, createdAt time.Time) *Punch {
	return &Punch{
		ID:         id,
		EmployeeID: employeeID,
		Type:       punchType,
		Timestamp:  timestamp,
		Date:       date,
		CreatedAt:  createdAt,
	}
}

func (p *Punch) IsToday() bool {
	return time.Now().In(timeutil.LoadDefaultLocation()).Format("2006-01-02") == p.Date.Format("2006-01-02")
}


