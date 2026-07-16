package usecase

import (
	"fmt"

	"hrms/internal/attendance/entity"
	"hrms/internal/attendance/models"
	"hrms/internal/pkg/timeutil"
)

func toHistoryItem(da *entity.DailyAttendance) models.MyAttendanceHistoryItem {
	day := fmt.Sprintf("%d", da.Date.Day())
	month := da.Date.Format("Jan")

	loc := timeutil.LoadDefaultLocation()

	var checkIn, checkOut *string
	if da.FirstPunchIn != nil {
		s := da.FirstPunchIn.In(loc).Format("15:04")
		checkIn = &s
	}
	if da.LastPunchOut != nil {
		s := da.LastPunchOut.In(loc).Format("15:04")
		checkOut = &s
	}

	var workingHours *string
	if da.TotalWorkSeconds != nil {
		h := *da.TotalWorkSeconds / 3600
		m := (*da.TotalWorkSeconds % 3600) / 60
		s := fmt.Sprintf("%02dh %02dm", h, m)
		workingHours = &s
	}

	attType := string(da.Status)
	if da.Status == entity.AttendanceOnLeave {
		attType = "leave"
	}

	var reason *string
	if da.LeaveTypeName != nil && *da.LeaveTypeName != "" {
		reason = da.LeaveTypeName
	}

	return models.MyAttendanceHistoryItem{
		Day:          day,
		Month:        month,
		Type:         attType,
		CheckIn:      checkIn,
		CheckOut:     checkOut,
		WorkingHours: workingHours,
		Reason:       reason,
		LateMinutes:  da.LateMinutes(),
	}
}

func toMyAttendance(da *entity.DailyAttendance, employeeName string) *models.MyAttendance {
	date := da.Date.Format("2006-01-02")

	loc := timeutil.LoadDefaultLocation()

	var clockIn, clockOut *string
	if da.FirstPunchIn != nil {
		s := da.FirstPunchIn.In(loc).Format("15:04:05")
		clockIn = &s
	}
	if da.LastPunchOut != nil {
		s := da.LastPunchOut.In(loc).Format("15:04:05")
		clockOut = &s
	}

	var totalHours float64
	if da.FirstPunchIn != nil && da.LastPunchOut != nil {
		totalHours = da.LastPunchOut.Sub(*da.FirstPunchIn).Hours()
	}

	lateMinutes := da.LateMinutes()

	isPresent := da.IsPresent()
	isLate := da.IsLate
	isAbsent := da.Status == entity.AttendanceAbsent
	isComplete := da.IsComplete()

	return &models.MyAttendance{
		EmployeeID:   da.EmployeeID,
		EmployeeName: employeeName,
		Date:         date,
		ClockIn:      clockIn,
		ClockOut:     clockOut,
		TotalHours:   totalHours,
		Status:       string(da.Status),
		LateMinutes:  lateMinutes,
		IsPresent:    isPresent,
		IsLate:       isLate,
		IsAbsent:     isAbsent,
		IsComplete:   isComplete,
	}
}
