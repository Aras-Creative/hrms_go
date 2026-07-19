package usecase

import (
	"context"
	"fmt"

	"hrms/internal/attendance/repository"
	"hrms/internal/pkg/timeutil"
	errors "hrms/internal/pkg/apperror"
)

type RecapRowData struct {
	EmployeeID      string
	EmployeeNumber  string
	FullName        string
	ProfilePhotoID  *string
	DesignationName *string
	WorkingDays     int
	Present         int
	Absent          int
	MissingClockOut int
	Late            int
	LateMinutes     int
	EarlyLeave      int
}

type RecapHeaderData struct {
	Key   string
	Label string
}

type RecapEmployeeData struct {
	ID                      string
	EmployeeNumber          string
	Name                    string
	ProfilePhotoID          *string
	Department              *string
	WorkingDays             int
	LateMinutes             int
	TotalAttendanceIncident int
	MetricValues            map[string]int
}

type RecapOutput struct {
	Headers   []RecapHeaderData
	Employees []RecapEmployeeData
}

type RecapResult struct {
	LeaveTypes []repository.LeaveTypeRow
	Rows       []*repository.RecapRow
}

func recapCacheKey(fromStr, toStr, designationID string) string {
	return fmt.Sprintf("%s:%s:%s", fromStr, toStr, designationID)
}

func (uc *DailyAttendanceUsecase) Recap(ctx context.Context, fromStr, toStr, designationID string) (*RecapResult, error) {
	from, to, err := timeutil.ParseDateRange(fromStr, toStr)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	key := recapCacheKey(fromStr, toStr, designationID)
	if cached, ok := uc.recapCache.Get(key); ok {
		return cached, nil
	}

	leaveTypes, err := uc.dailyRepo.FindActiveLeaveTypes(ctx)
	if err != nil {
		return nil, errors.WrapInternal("failed to fetch leave types", err)
	}

	rows, err := uc.dailyRepo.Recap(ctx, from, to, designationID)
	if err != nil {
		return nil, errors.WrapInternal("failed to fetch recap data", err)
	}

	result := &RecapResult{LeaveTypes: leaveTypes, Rows: rows}
	uc.recapCache.Set(key, result)
	return result, nil
}

// BuildRecapOutput takes a raw RecapResult and produces aggregated RecapOutput
// with headers, employee metrics, and incident counts.
func (uc *DailyAttendanceUsecase) BuildRecapOutput(result *RecapResult) *RecapOutput {
	// Build headers: present + each active leave type + absent + missing_clock_out + late + early_leave
	headers := []RecapHeaderData{
		{Key: "present", Label: "Hadir"},
	}
	ltKeys := make(map[string]string)
	for _, lt := range result.LeaveTypes {
		headers = append(headers, RecapHeaderData{Key: "leave_" + lt.ID, Label: lt.Name})
		ltKeys[lt.Name] = "leave_" + lt.ID
	}
	headers = append(headers,
		RecapHeaderData{Key: "absent", Label: "Absen"},
		RecapHeaderData{Key: "missing_clock_out", Label: "Tidak Absen Pulang"},
		RecapHeaderData{Key: "late", Label: "Terlambat"},
		RecapHeaderData{Key: "early_leave", Label: "Pulang Cepat"},
	)

	// Aggregate rows by employee
	type empAgg struct {
		ID              string
		EmployeeNumber  string
		Name            string
		ProfilePhotoID  *string
		Department      *string
		WorkingDays     int
		Present         int
		LeaveCounts     map[string]int
		Absent          int
		MissingClockOut int
		Late            int
		LateMinutes     int
		EarlyLeave      int
	}
	empMap := make(map[string]*empAgg)

	for _, r := range result.Rows {
		agg, ok := empMap[r.EmployeeID]
		if !ok {
			agg = &empAgg{
				ID:             r.EmployeeID,
				EmployeeNumber: r.EmployeeNumber,
				Name:           r.FullName,
				ProfilePhotoID: r.ProfilePhotoID,
				Department:     r.DesignationName,
				LeaveCounts:    make(map[string]int),
			}
			empMap[r.EmployeeID] = agg
		}
		agg.WorkingDays = r.WorkingDays
		agg.Present += r.Present
		agg.Absent += r.Absent
		agg.MissingClockOut += r.MissingClockOut
		agg.Late += r.Late
		agg.LateMinutes += r.LateMinutes
		agg.EarlyLeave += r.EarlyLeave

		if r.LeaveTypeName != nil {
			key := ltKeys[*r.LeaveTypeName]
			if key != "" {
				agg.LeaveCounts[key]++
			}
		}
	}

	// Build employees
	employees := make([]RecapEmployeeData, 0, len(empMap))
	for _, agg := range empMap {
		incidents := agg.Late + agg.EarlyLeave + agg.Absent + agg.MissingClockOut

		metricValues := map[string]int{
			"present": agg.Present,
		}
		for _, lt := range result.LeaveTypes {
			key := "leave_" + lt.ID
			metricValues[key] = agg.LeaveCounts[key]
		}
		metricValues["absent"] = agg.Absent
		metricValues["missing_clock_out"] = agg.MissingClockOut
		metricValues["late"] = agg.Late
		metricValues["early_leave"] = agg.EarlyLeave

		employees = append(employees, RecapEmployeeData{
			ID:                      agg.ID,
			EmployeeNumber:          agg.EmployeeNumber,
			Name:                    agg.Name,
			ProfilePhotoID:          agg.ProfilePhotoID,
			Department:              agg.Department,
			WorkingDays:             agg.WorkingDays,
			LateMinutes:             agg.LateMinutes,
			TotalAttendanceIncident: incidents,
			MetricValues:            metricValues,
		})
	}

	return &RecapOutput{
		Headers:   headers,
		Employees: employees,
	}
}
