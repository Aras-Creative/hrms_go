package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PayrollPeriod struct {
	ID        string
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    PeriodStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPayrollPeriod(name string, startDate, endDate time.Time) *PayrollPeriod {
	now := time.Now()
	return &PayrollPeriod{
		ID:        uuid.New().String(),
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    PeriodStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func ReconstitutePayrollPeriod(id, name string, startDate, endDate time.Time, status PeriodStatus, createdAt, updatedAt time.Time) *PayrollPeriod {
	return &PayrollPeriod{
		ID:        id,
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (p *PayrollPeriod) MarkDraft() error {
	if !p.Status.CanTransitionTo(PeriodStatusDraft) {
		return fmt.Errorf("cannot reset period to draft from %s status", p.Status)
	}
	p.Status = PeriodStatusDraft
	p.UpdatedAt = time.Now()
	return nil
}

func (p *PayrollPeriod) MarkProcessed() error {
	if !p.Status.CanTransitionTo(PeriodStatusProcessed) {
		return fmt.Errorf("cannot process period in %s status", p.Status)
	}
	p.Status = PeriodStatusProcessed
	p.UpdatedAt = time.Now()
	return nil
}

func (p *PayrollPeriod) MarkClosed() error {
	if !p.Status.CanTransitionTo(PeriodStatusClosed) {
		return fmt.Errorf("cannot close period in %s status", p.Status)
	}
	p.Status = PeriodStatusClosed
	p.UpdatedAt = time.Now()
	return nil
}
