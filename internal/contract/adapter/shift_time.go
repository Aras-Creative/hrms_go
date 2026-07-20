package adapter

import (
	"context"
	"fmt"

	scheduleRepo "hrms/internal/schedule/repository"
	contractUc "hrms/internal/contract/usecase"
)

type ShiftTimeFetcherAdapter struct {
	wpRepo  scheduleRepo.WorkPatternRepository
	ewpRepo scheduleRepo.EmployeeWorkPatternRepository
}

func NewShiftTimeFetcherAdapter(wpRepo scheduleRepo.WorkPatternRepository, ewpRepo scheduleRepo.EmployeeWorkPatternRepository) *ShiftTimeFetcherAdapter {
	return &ShiftTimeFetcherAdapter{wpRepo: wpRepo, ewpRepo: ewpRepo}
}

// FindShiftTimesByEmployeeID returns the shift start and end time for an employee's
// active work pattern. It checks weekday details (Monday=1..Friday=5) and
// returns the times of the first matching detail.
func (a *ShiftTimeFetcherAdapter) FindShiftTimesByEmployeeID(ctx context.Context, employeeID string) (start, end string, err error) {
	ewp, err := a.ewpRepo.FindActiveByEmployee(ctx, employeeID)
	if err != nil {
		return "", "", fmt.Errorf("find active work pattern: %w", err)
	}
	if ewp == nil || ewp.WorkPatternID == "" {
		return "", "", nil
	}

	wp, err := a.wpRepo.FindByID(ctx, ewp.WorkPatternID)
	if err != nil {
		return "", "", fmt.Errorf("find work pattern: %w", err)
	}
	if wp == nil {
		return "", "", nil
	}

	for _, d := range wp.Details {
		if string(d.Type) == "off" {
			continue
		}
		if int(d.DayOfWeek) >= 1 && int(d.DayOfWeek) <= 5 {
			start = ""
			end = ""
			if d.StartTime != nil {
				start = *d.StartTime
			}
			if d.EndTime != nil {
				end = *d.EndTime
			}
			return start, end, nil
		}
	}

	// fallback: return the first non-off detail's times
	for _, d := range wp.Details {
		if string(d.Type) != "off" {
			if d.StartTime != nil {
				start = *d.StartTime
			}
			if d.EndTime != nil {
				end = *d.EndTime
			}
			return start, end, nil
		}
	}

	return "", "", nil
}

var _ contractUc.ShiftTimeFetcher = (*ShiftTimeFetcherAdapter)(nil)
