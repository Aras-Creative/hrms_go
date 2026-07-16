package adapter

import (
	"context"
	"fmt"
	"time"

	attendanceEntity "hrms/internal/attendance/entity"
	punchRepo "hrms/internal/attendance/repository"
	attendanceUc "hrms/internal/attendance/usecase"
	employeeRepo "hrms/internal/employee/repository"
	leaveUc "hrms/internal/leave/usecase"
)

type AttendanceProcessorAdapter struct {
	processor  *attendanceUc.DailyProcessor
	punchRepo  punchRepo.PunchRepository
	empRepo    employeeRepo.EmployeeRepository
}

func NewAttendanceProcessorAdapter(processor *attendanceUc.DailyProcessor, punchRepo punchRepo.PunchRepository, empRepo employeeRepo.EmployeeRepository) *AttendanceProcessorAdapter {
	return &AttendanceProcessorAdapter{processor: processor, punchRepo: punchRepo, empRepo: empRepo}
}

func (a *AttendanceProcessorAdapter) ReprocessDay(ctx context.Context, employeeID string, date time.Time) (skippedDueToCorrection bool, err error) {
	da, err := a.processor.ProcessDaily(ctx, employeeID, date)
	if err != nil {
		return false, fmt.Errorf("process daily: %w", err)
	}
	// Correction always wins — if the record has source="correction", leave changes won't apply
	if da.Source == "correction" {
		return true, nil
	}
	return false, nil
}

func (a *AttendanceProcessorAdapter) EnsureHalfDayPunches(ctx context.Context, employeeID string, date time.Time) (createdIn, createdOut bool, err error) {
	da, err := a.processor.ComputeDaily(ctx, employeeID, date)
	if err != nil {
		return false, false, fmt.Errorf("compute daily: %w", err)
	}

	now := time.Now()

	if da.FirstPunchIn == nil {
		p := attendanceEntity.NewPunch(employeeID, attendanceEntity.PunchIn, now)
		if err := a.punchRepo.Create(ctx, p); err != nil {
			return false, false, fmt.Errorf("create clock-in: %w", err)
		}
		createdIn = true
	} else if da.LastPunchOut == nil {
		punchOut := time.Date(date.Year(), date.Month(), date.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
		p := attendanceEntity.NewPunch(employeeID, attendanceEntity.PunchOut, punchOut)
		if err := a.punchRepo.Create(ctx, p); err != nil {
			return false, false, fmt.Errorf("create clock-out: %w", err)
		}
		createdOut = true
	}

	return createdIn, createdOut, nil
}

var _ leaveUc.AttendanceReprocessor = (*AttendanceProcessorAdapter)(nil)
var _ leaveUc.HalfDayPunchHandler = (*AttendanceProcessorAdapter)(nil)
