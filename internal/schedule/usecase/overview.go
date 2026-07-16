package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/schedule/entity"
	"hrms/internal/schedule/repository"
)

type ScheduleOverviewUsecase struct {
	repo repository.ScheduleOverrideRepository
}

func NewScheduleOverviewUsecase(repo repository.ScheduleOverrideRepository) *ScheduleOverviewUsecase {
	return &ScheduleOverviewUsecase{repo: repo}
}

func (uc *ScheduleOverviewUsecase) Query(ctx context.Context, p ScheduleOverviewParams) ([]ScheduleOverviewEmployee, error) {
	from, to, err := parseOverviewDates(p.From, p.To)
	if err != nil {
		return nil, err
	}

	rows, err := uc.repo.QueryOverview(ctx, repository.ScheduleOverviewParams{
		EmployeeID:    p.EmployeeID,
		DesignationID: p.DesignationID,
		Search:        p.Search,
		From:          from,
		To:            to,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query overview: %w", err)
	}

	return groupOverviewRows(rows), nil
}

func parseOverviewDates(fromStr, toStr string) (time.Time, time.Time, error) {
	from := time.Now().AddDate(0, -1, 0)
	if fromStr != "" {
		parsed, err := tryParseDate(fromStr)
		if err != nil {
			return from, from, fmt.Errorf("invalid from date: %w", err)
		}
		from = parsed
	}

	to := time.Now().AddDate(0, 1, 0)
	if toStr != "" {
		parsed, err := tryParseDate(toStr)
		if err != nil {
			return from, from, fmt.Errorf("invalid to date: %w", err)
		}
		to = parsed
	}

	return from, to, nil
}

func groupOverviewRows(rows []repository.ScheduleOverviewRow) []ScheduleOverviewEmployee {
	type empKey struct {
		id   string
		name string
		num  string
	}

	empMap := make(map[empKey]*ScheduleOverviewEmployee)
	var order []empKey

	for _, row := range rows {
		key := empKey{id: row.EmployeeID, name: row.FullName, num: row.EmployeeNumber}
		emp, ok := empMap[key]
		if !ok {
			emp = &ScheduleOverviewEmployee{
				EmployeeID:         row.EmployeeID,
				EmployeeName:       row.FullName,
				EmployeeNumber:     row.EmployeeNumber,
				DesignationName:    row.DesignationName,
				WorkingPatternID:   row.WorkPatternID,
				WorkingPatternName: row.WorkingPatternName,
			}
			empMap[key] = emp
			order = append(order, key)
		}

		date := buildOverviewDate(row)
		emp.Dates = append(emp.Dates, date)
	}

	result := make([]ScheduleOverviewEmployee, 0, len(order))
	for _, key := range order {
		result = append(result, *empMap[key])
	}
	return result
}

func buildOverviewDate(row repository.ScheduleOverviewRow) ScheduleOverviewDate {
	var source ScheduleSource
	var isWorkingDay bool
	var startTime, endTime, notes, overrideID, workingType *string

	if row.OverrideID != nil {
		source = SourceOverride
		if row.OverrideIsWorking != nil {
			isWorkingDay = *row.OverrideIsWorking
		}
		if row.OverrideStartTime != nil {
			startTime = row.OverrideStartTime
		} else if row.PatternStartTime != nil {
			startTime = row.PatternStartTime
		}
		if row.OverrideEndTime != nil {
			endTime = row.OverrideEndTime
		} else if row.PatternEndTime != nil {
			endTime = row.PatternEndTime
		}
		notes = row.OverrideNotes
		overrideID = row.OverrideID
		workingType = row.PatternType
	} else if row.PatternDetailID != nil {
		source = SourceWorkingPattern
		workingType = row.PatternType
		if row.PatternType != nil && *row.PatternType == string(entity.WorkingTypeOff) {
			isWorkingDay = false
		} else if row.PatternType != nil && *row.PatternType == string(entity.WorkingTypeDynamic) {
			isWorkingDay = true
		} else {
			isWorkingDay = row.PatternStartTime != nil && row.PatternEndTime != nil
		}
		startTime = row.PatternStartTime
		endTime = row.PatternEndTime
	} else if row.WorkPatternID != nil {
		source = SourceNoPattern
	} else {
		source = SourceNone
	}

	return ScheduleOverviewDate{
		Date:         row.Date.Format("2006-01-02"),
		Source:       source,
		IsWorkingDay: isWorkingDay,
		WorkingType:  workingType,
		StartTime:    startTime,
		EndTime:      endTime,
		Notes:        notes,
		OverrideID:   overrideID,
	}
}
