package adapter

import (
	"context"
	"time"

	attendanceUc "hrms/internal/attendance/usecase"
	scheduleRepo "hrms/internal/schedule/repository"
)

type ScheduleResolverAdapter struct {
	overrideRepo scheduleRepo.ScheduleOverrideRepository
}

func NewScheduleResolverAdapter(overrideRepo scheduleRepo.ScheduleOverrideRepository) *ScheduleResolverAdapter {
	return &ScheduleResolverAdapter{overrideRepo: overrideRepo}
}

func (a *ScheduleResolverAdapter) ResolveRange(ctx context.Context, employeeID string, from, to time.Time) (map[string]map[string]*attendanceUc.ResolvedSchedule, error) {
	rows, err := a.overrideRepo.QueryOverview(ctx, scheduleRepo.ScheduleOverviewParams{
		EmployeeID: employeeID,
		From:       from,
		To:         to,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]*attendanceUc.ResolvedSchedule)
	for _, r := range rows {
		dateKey := r.Date.Format("2006-01-02")
		rs := &attendanceUc.ResolvedSchedule{}

		isPatternOff := r.PatternType != nil && *r.PatternType == "off"
		isPatternDynamic := r.PatternType != nil && *r.PatternType == "dynamic"

		switch {
		case r.OverrideIsWorking != nil && !*r.OverrideIsWorking:
			rs.Source = "override"
			rs.OverrideIsWorking = r.OverrideIsWorking
			rs.ScheduleOverrideID = r.OverrideID
		case r.OverrideStartTime != nil && *r.OverrideStartTime != "":
			rs.ExpectedStartTime = r.OverrideStartTime
			rs.ExpectedEndTime = r.OverrideEndTime
			rs.Source = "override"
			rs.ScheduleOverrideID = r.OverrideID
			rs.OverrideIsWorking = r.OverrideIsWorking
			rs.WorkingType = "fixed"
		case r.OverrideIsWorking != nil && *r.OverrideIsWorking:
			rs.ExpectedStartTime = r.PatternStartTime
			rs.ExpectedEndTime = r.PatternEndTime
			rs.Source = "override"
			rs.ScheduleOverrideID = r.OverrideID
			rs.OverrideIsWorking = r.OverrideIsWorking
		case isPatternDynamic && !isPatternOff:
			rs.ExpectedStartTime = r.PatternStartTime
			rs.ExpectedEndTime = r.PatternEndTime
			rs.Source = "working_pattern"
			rs.WorkingType = "dynamic"
		case r.PatternStartTime != nil && *r.PatternStartTime != "" && !isPatternOff:
			rs.ExpectedStartTime = r.PatternStartTime
			rs.ExpectedEndTime = r.PatternEndTime
			rs.Source = "working_pattern"
			if r.PatternType != nil {
				rs.WorkingType = *r.PatternType
			}
		default:
			rs.Source = "none"
		}

		if result[r.EmployeeID] == nil {
			result[r.EmployeeID] = make(map[string]*attendanceUc.ResolvedSchedule)
		}
		result[r.EmployeeID][dateKey] = rs
	}
	return result, nil
}

var _ attendanceUc.ScheduleResolver = (*ScheduleResolverAdapter)(nil)
