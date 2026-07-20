package usecase

import (
	"context"
	"fmt"
	"sort"
	"time"

	"hrms/internal/attendance/entity"
	"hrms/internal/attendance/models"
	"hrms/internal/attendance/repository"
	"hrms/internal/pkg/timeutil"
	errors "hrms/internal/pkg/apperror"
)

type MeUsecase struct {
	fetcher          EmployeeFetcher
	processor        *DailyProcessor
	dailyRepo        repository.DailyAttendanceRepository
	correctionFetcher CorrectionAuditFetcher
}

func NewMeUsecase(fetcher EmployeeFetcher, processor *DailyProcessor, dailyRepo repository.DailyAttendanceRepository, correctionFetcher CorrectionAuditFetcher) *MeUsecase {
	return &MeUsecase{fetcher: fetcher, processor: processor, dailyRepo: dailyRepo, correctionFetcher: correctionFetcher}
}

func (uc *MeUsecase) GetMyAttendance(ctx context.Context, userID string) (*models.MyAttendance, error) {
	employeeID, employeeName, err := uc.fetcher.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find employee", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("employee not found for user")
	}

	now := time.Now().In(timeutil.LoadDefaultLocation())
	today := entity.LocalDate(now)

	da, err := uc.processor.ComputeDaily(ctx, employeeID, today)
	if err != nil {
		return nil, errors.WrapInternal("failed to get attendance", err)
	}

	return toMyAttendance(da, employeeName), nil
}

func (uc *MeUsecase) GetMyStats(ctx context.Context, userID string) ([]models.MonthlyStats, error) {
	employeeID, _, err := uc.fetcher.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find employee", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("employee not found for user")
	}

	now := time.Now().In(timeutil.LoadDefaultLocation())
	from := time.Date(now.Year()-1, now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	records, err := uc.dailyRepo.FindByEmployeeAndDateRange(ctx, employeeID, from, to)
	if err != nil {
		return nil, errors.WrapInternal("failed to find records", err)
	}

	// Index persisted records by date so we can detect missing today.
	today := entity.LocalDate(now)
	seen := make(map[string]bool)
	monthMap := make(map[string]*models.MonthlyStats)
	for _, r := range records {
		dateKey := r.Date.Format("2006-01-02")
		seen[dateKey] = true

		key := r.Date.Format("2006-01")
		ms, ok := monthMap[key]
		if !ok {
			ms = &models.MonthlyStats{Month: key}
			monthMap[key] = ms
		}

		switch r.Status {
		case entity.AttendancePresent:
			ms.Present++
		case entity.AttendanceOnLeave:
			ms.OnLeave++
		case entity.AttendanceAbsent:
			ms.Absent++
		case entity.AttendanceDayOff:
			ms.DayOff++
		case entity.AttendanceNoPunch:
			// skip — no_punch is a transient state, not counted as absent
		}
		ms.LateMinutes += r.LateMinutes()
	}

	// If today isn't in the persisted records, compute it on the fly so the
	// stats reflect the current state (e.g. before the first punch or sweep).
	todayKey := today.Format("2006-01-02")
	if !seen[todayKey] {
		da, err := uc.processor.ComputeDaily(ctx, employeeID, today)
		if err == nil && da != nil {
			key := today.Format("2006-01")
			ms, ok := monthMap[key]
			if !ok {
				ms = &models.MonthlyStats{Month: key}
				monthMap[key] = ms
			}
			switch da.Status {
			case entity.AttendancePresent:
				ms.Present++
			case entity.AttendanceOnLeave:
				ms.OnLeave++
			case entity.AttendanceAbsent:
				ms.Absent++
			case entity.AttendanceDayOff:
				ms.DayOff++
			case entity.AttendanceNoPunch:
				// skip
			}
			ms.LateMinutes += da.LateMinutes()
		}
	}

	result := make([]models.MonthlyStats, 0, len(monthMap))
	var keys []string
	for k := range monthMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result = append(result, *monthMap[k])
	}
	return result, nil
}

func (uc *MeUsecase) GetMyAttendanceHistory(ctx context.Context, userID, fromStr, toStr string) ([]models.MyAttendanceHistoryItem, error) {
	employeeID, _, err := uc.fetcher.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find employee", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("employee not found for user")
	}

	from, to, err := timeutil.ParseDateRange(fromStr, toStr)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	records, err := uc.processor.ComputeDailyRange(ctx, employeeID, from, to)
	if err != nil {
		return nil, err
	}

	var correctionMap map[string]*CorrectionAuditInfo
	if uc.correctionFetcher != nil {
		correctionMap, err = uc.correctionFetcher.FetchCorrectionLogs(ctx, employeeID, from, to)
		if err != nil {
			return nil, fmt.Errorf("fetch correction logs: %w", err)
		}
	}

	items := make([]models.MyAttendanceHistoryItem, len(records))
	for i, da := range records {
		var correction *CorrectionAuditInfo
		if correctionMap != nil {
			dateKey := da.Date.Format("2006-01-02")
			correction = correctionMap[dateKey]
		}
		items[i] = toHistoryItem(da, correction)
	}
	return items, nil
}
