package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/attendance/entity"
	"hrms/internal/attendance/repository"
	errors "hrms/internal/pkg/apperror"
)

type DailyProcessor struct {
	dailyRepo repository.DailyAttendanceRepository
	resolver  ScheduleResolver
}

func NewDailyProcessor(dailyRepo repository.DailyAttendanceRepository, resolver ScheduleResolver) *DailyProcessor {
	return &DailyProcessor{dailyRepo: dailyRepo, resolver: resolver}
}

// ProcessDailyRange iterates over a date range calling ProcessDaily for each day.
func (p *DailyProcessor) ProcessDailyRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.DailyAttendance, error) {
	var records []*entity.DailyAttendance
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		da, err := p.ProcessDaily(ctx, employeeID, d)
		if err != nil {
			return nil, errors.NewInternal(fmt.Sprintf("failed to process daily attendance for %s: %v", d.Format("2006-01-02"), err))
		}
		records = append(records, da)
	}
	return records, nil
}

// ComputeDailyRange iterates over a date range calling ComputeDaily for each day.
func (p *DailyProcessor) ComputeDailyRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.DailyAttendance, error) {
	var records []*entity.DailyAttendance
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		da, err := p.ComputeDaily(ctx, employeeID, d)
		if err != nil {
			return nil, errors.NewInternal(fmt.Sprintf("failed to get attendance for %s: %v", d.Format("2006-01-02"), err))
		}
		records = append(records, da)
	}
	return records, nil
}

func (p *DailyProcessor) ProcessDaily(ctx context.Context, employeeID string, date time.Time) (*entity.DailyAttendance, error) {
	existing, err := p.dailyRepo.FindByEmployeeAndDate(ctx, employeeID, date)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to check existing attendance: %v", err))
	}

	if existing != nil && !existing.CanBeOverwritten() {
		return existing, nil
	}

	da, err := p.ComputeDaily(ctx, employeeID, date)
	if err != nil {
		return nil, err
	}

	if err := p.dailyRepo.Upsert(ctx, da); err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to save attendance: %v", err))
	}

	return da, nil
}

func (p *DailyProcessor) ComputeDaily(ctx context.Context, employeeID string, date time.Time) (*entity.DailyAttendance, error) {
	existing, err := p.dailyRepo.FindByEmployeeAndDate(ctx, employeeID, date)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to check existing attendance: %v", err))
	}
	if existing != nil && !existing.CanBeOverwritten() {
		return existing, nil
	}

	schedules, err := p.resolver.ResolveRange(ctx, employeeID, date, date)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to resolve schedule: %v", err))
	}

	row, err := p.dailyRepo.ComputeDaily(ctx, employeeID, date)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to compute daily: %v", err))
	}

	dateKey := date.Format("2006-01-02")
	if s, ok := schedules[employeeID][dateKey]; ok {
		row.ExpectedStartTime = s.ExpectedStartTime
		row.ExpectedEndTime = s.ExpectedEndTime
		row.Source = s.Source
		row.ScheduleOverrideID = s.ScheduleOverrideID
		row.OverrideIsWorking = s.OverrideIsWorking
		row.WorkingType = s.WorkingType
	}

	da := entity.NewDailyAttendance(row.EmployeeID, row.Date)
	da.ApplyScheduleAndPunches(
		row.ExpectedStartTime, row.ExpectedEndTime, row.Source,
		row.FirstPunchIn, row.LastPunchOut, row.TotalWorkSeconds,
		row.ScheduleOverrideID,
		row.LeaveSubmissionID, row.LeaveTypeName, row.LeaveIsHalfDay, row.OverrideIsWorking,
		row.WorkingType,
	)

	return da, nil
}

func (p *DailyProcessor) ProcessRange(ctx context.Context, from, to time.Time) error {
	schedules, err := p.resolver.ResolveRange(ctx, "", from, to)
	if err != nil {
		return errors.NewInternal(fmt.Sprintf("failed to resolve schedules: %v", err))
	}

	rows, err := p.dailyRepo.ComputeRange(ctx, from, to)
	if err != nil {
		return errors.NewInternal(fmt.Sprintf("failed to compute range: %v", err))
	}

	seen := make(map[string]bool)
	var errs []error

	for _, row := range rows {
		key := row.EmployeeID + "|" + row.Date.Format("2006-01-02")
		if seen[key] {
			continue
		}
		seen[key] = true

		existing, err := p.dailyRepo.FindByEmployeeAndDate(ctx, row.EmployeeID, row.Date)
		if err == nil && existing != nil && !existing.CanBeOverwritten() {
			continue
		}

		dateKey := row.Date.Format("2006-01-02")
		if s, ok := schedules[row.EmployeeID][dateKey]; ok {
			row.ExpectedStartTime = s.ExpectedStartTime
			row.ExpectedEndTime = s.ExpectedEndTime
			row.Source = s.Source
			row.ScheduleOverrideID = s.ScheduleOverrideID
			row.OverrideIsWorking = s.OverrideIsWorking
			row.WorkingType = s.WorkingType
		}

		da := entity.NewDailyAttendance(row.EmployeeID, row.Date)
		da.ApplyScheduleAndPunches(
			row.ExpectedStartTime, row.ExpectedEndTime, row.Source,
			row.FirstPunchIn, row.LastPunchOut, row.TotalWorkSeconds,
			row.ScheduleOverrideID,
			row.LeaveSubmissionID, row.LeaveTypeName, row.LeaveIsHalfDay, row.OverrideIsWorking,
			row.WorkingType,
		)

		if err := p.dailyRepo.Upsert(ctx, da); err != nil {
			errs = append(errs, errors.NewInternal(fmt.Sprintf("failed to upsert %s on %s: %v", row.EmployeeID, row.Date.Format("2006-01-02"), err)))
		}
	}

	if len(errs) > 0 {
		return errors.NewInternal(fmt.Sprintf("process range had %d errors: %v", len(errs), errs[0]))
	}
	return nil
}
