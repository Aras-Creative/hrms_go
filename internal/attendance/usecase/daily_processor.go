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
	dailyRepo      repository.DailyAttendanceRepository
	correctionRepo repository.CorrectionRepository
	resolver       ScheduleResolver
}

// applyCorrectionOverlay fetches an existing correction for the employee+date
// and overlays it on top of the freshly computed attendance. This ensures
// corrections always win as the top priority layer while underlying events
// (punches, schedule changes) are never stale.
func (p *DailyProcessor) applyCorrectionOverlay(ctx context.Context, da *entity.DailyAttendance, employeeID string, date time.Time) {
	correction, err := p.correctionRepo.FindByEmployeeAndDate(ctx, employeeID, date)
	if err == nil && correction != nil {
		correction.ApplyTo(da)
	}
}

func NewDailyProcessor(
	dailyRepo repository.DailyAttendanceRepository,
	correctionRepo repository.CorrectionRepository,
	resolver ScheduleResolver,
) *DailyProcessor {
	return &DailyProcessor{
		dailyRepo:      dailyRepo,
		correctionRepo: correctionRepo,
		resolver:       resolver,
	}
}

// ComputeDailyRange iterates over a date range calling ComputeDaily for each day.
func (p *DailyProcessor) ComputeDailyRange(ctx context.Context, employeeID string, from, to time.Time) ([]*entity.DailyAttendance, error) {
	var records []*entity.DailyAttendance
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		da, err := p.ComputeDaily(ctx, employeeID, d)
		if err != nil {
			return nil, errors.WrapInternal(fmt.Sprintf("failed to get attendance for %s", d.Format("2006-01-02")), err)
		}
		records = append(records, da)
	}
	return records, nil
}

func (p *DailyProcessor) ProcessDaily(ctx context.Context, employeeID string, date time.Time) (*entity.DailyAttendance, error) {
	da, err := p.ComputeDaily(ctx, employeeID, date)
	if err != nil {
		return nil, err
	}

	if err := p.dailyRepo.Upsert(ctx, da); err != nil {
		return nil, errors.WrapInternal("failed to save attendance", err)
	}

	return da, nil
}

// computeBase computes attendance from raw events (schedule, punches, leave)
// without applying any correction overlay.
func (p *DailyProcessor) computeBase(ctx context.Context, employeeID string, date time.Time) (*entity.DailyAttendance, error) {
	schedules, err := p.resolver.ResolveRange(ctx, employeeID, date, date)
	if err != nil {
		return nil, errors.WrapInternal("failed to resolve schedule", err)
	}

	row, err := p.dailyRepo.ComputeDaily(ctx, employeeID, date)
	if err != nil {
		return nil, errors.WrapInternal("failed to compute daily", err)
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

// ComputeDaily computes attendance from raw events and overlays any existing
// correction on top.
func (p *DailyProcessor) ComputeDaily(ctx context.Context, employeeID string, date time.Time) (*entity.DailyAttendance, error) {
	da, err := p.computeBase(ctx, employeeID, date)
	if err != nil {
		return nil, err
	}

	p.applyCorrectionOverlay(ctx, da, employeeID, date)
	return da, nil
}

// ProcessRange processes all employees in the date range.
func (p *DailyProcessor) ProcessRange(ctx context.Context, from, to time.Time) error {
	schedules, err := p.resolver.ResolveRange(ctx, "", from, to)
	if err != nil {
		return errors.WrapInternal("failed to resolve schedules", err)
	}

	rows, err := p.dailyRepo.ComputeRange(ctx, from, to)
	if err != nil {
		return errors.WrapInternal("failed to compute range", err)
	}

	seen := make(map[string]bool)
	var errs []error

	for _, row := range rows {
		key := row.EmployeeID + "|" + row.Date.Format("2006-01-02")
		if seen[key] {
			continue
		}
		seen[key] = true

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

		p.applyCorrectionOverlay(ctx, da, row.EmployeeID, row.Date)

		if err := p.dailyRepo.Upsert(ctx, da); err != nil {
			errs = append(errs, errors.WrapInternal(fmt.Sprintf("failed to upsert %s on %s", row.EmployeeID, row.Date.Format("2006-01-02")), err))
		}
	}

	if len(errs) > 0 {
		return errors.WrapInternal(fmt.Sprintf("process range had %d errors", len(errs)), errs[0])
	}
	return nil
}
