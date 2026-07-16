package entity

import (
	"fmt"
	"time"

	"hrms/internal/pkg/timeutil"

	"github.com/google/uuid"
)

type AttendanceStatus string

const (
	AttendancePresent AttendanceStatus = "present"
	AttendanceAbsent  AttendanceStatus = "absent"
	AttendanceNoPunch AttendanceStatus = "no_punch"
	AttendanceOnLeave AttendanceStatus = "on_leave"
	AttendanceDayOff  AttendanceStatus = "day_off"
)

var validStatuses = map[AttendanceStatus]bool{
	AttendancePresent: true,
	AttendanceAbsent:  true,
	AttendanceNoPunch: true,
	AttendanceOnLeave: true,
	AttendanceDayOff:  true,
}

func ParseAttendanceStatus(s string) (AttendanceStatus, error) {
	st := AttendanceStatus(s)
	if !validStatuses[st] {
		return "", fmt.Errorf("invalid attendance status: %s", s)
	}
	return st, nil
}

type DailyAttendance struct {
	ID                 string
	EmployeeID         string
	Date               time.Time
	Status             AttendanceStatus
	IsLate             bool
	IsEarlyLeave       bool
	ExpectedStartTime  *string
	ExpectedEndTime    *string
	Source             string
	FirstPunchIn       *time.Time
	LastPunchOut       *time.Time
	TotalWorkSeconds   *int
	LeaveSubmissionID  *string
	LeaveTypeName      *string
	ScheduleOverrideID *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewDailyAttendance(employeeID string, date time.Time) *DailyAttendance {
	now := time.Now()
	return &DailyAttendance{
		ID:           uuid.New().String(),
		EmployeeID:   employeeID,
		Date:         date,
		IsLate:       false,
		IsEarlyLeave: false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func ReconstituteDailyAttendance(
	id, employeeID string,
	date time.Time,
	status AttendanceStatus,
	isLate, isEarlyLeave bool,
	expectedStartTime, expectedEndTime *string,
	source string,
	firstPunchIn, lastPunchOut *time.Time,
	totalWorkSeconds *int,
	leaveSubmissionID, leaveTypeName *string,
	scheduleOverrideID *string,
	createdAt, updatedAt time.Time,
) *DailyAttendance {
	return &DailyAttendance{
		ID:                 id,
		EmployeeID:         employeeID,
		Date:               date,
		Status:             status,
		IsLate:             isLate,
		IsEarlyLeave:       isEarlyLeave,
		ExpectedStartTime:  expectedStartTime,
		ExpectedEndTime:    expectedEndTime,
		Source:             source,
		FirstPunchIn:       firstPunchIn,
		LastPunchOut:       lastPunchOut,
		TotalWorkSeconds:   totalWorkSeconds,
		LeaveSubmissionID:  leaveSubmissionID,
		LeaveTypeName:      leaveTypeName,
		ScheduleOverrideID: scheduleOverrideID,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
}

func (d *DailyAttendance) SetStatus(s AttendanceStatus) error {
	if !validStatuses[s] {
		return fmt.Errorf("invalid attendance status: %s", s)
	}
	d.Status = s
	d.UpdatedAt = time.Now()
	return nil
}

func (d *DailyAttendance) MarkOnLeave(submissionID, leaveTypeName string) {
	d.Status = AttendanceOnLeave
	d.LeaveSubmissionID = &submissionID
	d.LeaveTypeName = &leaveTypeName
	d.UpdatedAt = time.Now()
}

func (d *DailyAttendance) MarkPresent() {
	d.Status = AttendancePresent
	d.IsLate = false
	d.IsEarlyLeave = false
	d.UpdatedAt = time.Now()
}

func (d *DailyAttendance) MarkAbsent() {
	d.Status = AttendanceAbsent
	d.UpdatedAt = time.Now()
}

func (d *DailyAttendance) MarkNoPunch() {
	d.Status = AttendanceNoPunch
	d.UpdatedAt = time.Now()
}

func (d *DailyAttendance) MarkDayOff() {
	d.Status = AttendanceDayOff
	d.UpdatedAt = time.Now()
}

// IsPresent returns true if the employee has clocked in (has a first punch).
func (d *DailyAttendance) IsPresent() bool {
	return d.FirstPunchIn != nil
}

// IsComplete returns true if the employee has both clocked in and clocked out.
func (d *DailyAttendance) IsComplete() bool {
	return d.FirstPunchIn != nil && d.LastPunchOut != nil
}

// ApplyScheduleAndPunches sets computed schedule and punch data onto this entity,
// then runs EvaluateAndDetermineStatus to derive the final attendance status.
func (d *DailyAttendance) ApplyScheduleAndPunches(
	expectedStart, expectedEnd *string,
	source string,
	firstPunchIn, lastPunchOut *time.Time,
	totalWorkSeconds *int,
	scheduleOverrideID *string,
	leaveSubmissionID, leaveTypeName *string,
	leaveIsHalfDay, overrideIsWorking *bool,
	workingType string,
) {
	d.ExpectedStartTime = expectedStart
	d.ExpectedEndTime = expectedEnd
	d.Source = source
	d.FirstPunchIn = firstPunchIn
	d.LastPunchOut = lastPunchOut
	d.TotalWorkSeconds = totalWorkSeconds
	d.ScheduleOverrideID = scheduleOverrideID
	d.EvaluateAndDetermineStatus(leaveSubmissionID, leaveTypeName, leaveIsHalfDay, overrideIsWorking, workingType)
}

func (d *DailyAttendance) LateMinutes() int {
	if d.FirstPunchIn == nil || d.ExpectedStartTime == nil || *d.ExpectedStartTime == "" {
		return 0
	}
	expected, err := time.Parse("15:04", *d.ExpectedStartTime)
	if err != nil {
		return 0
	}
	loc := timeutil.LoadDefaultLocation()
	punch := d.FirstPunchIn.In(loc)
	ref := time.Date(punch.Year(), punch.Month(), punch.Day(), expected.Hour(), expected.Minute(), 0, 0, loc)
	if !punch.After(ref) {
		return 0
	}
	return int(punch.Sub(ref).Minutes())
}

func (d *DailyAttendance) EarlyLeaveMinutes() int {
	if d.LastPunchOut == nil || d.ExpectedEndTime == nil || *d.ExpectedEndTime == "" {
		return 0
	}
	expected, err := time.Parse("15:04", *d.ExpectedEndTime)
	if err != nil {
		return 0
	}
	loc := timeutil.LoadDefaultLocation()
	punch := d.LastPunchOut.In(loc)
	ref := time.Date(punch.Year(), punch.Month(), punch.Day(), expected.Hour(), expected.Minute(), 0, 0, loc)
	if !punch.Before(ref) {
		return 0
	}
	return int(ref.Sub(punch).Minutes())
}

func (d *DailyAttendance) CanBeOverwritten() bool {
	if d == nil {
		return true
	}
	return d.Source != "correction" && d.Source != "legacy"
}

func (d *DailyAttendance) EvaluateAndDetermineStatus(
	leaveSubmissionID *string,
	leaveTypeName *string,
	leaveIsHalfDay *bool,
	overrideIsWorking *bool,
	workingType string,
) {
	if d.evaluateLeaveCase(leaveSubmissionID, leaveTypeName, leaveIsHalfDay) {
		return
	}
	if d.evaluateOverrideDayOff(overrideIsWorking) {
		return
	}
	if d.evaluateWorkCase(workingType) {
		return
	}
	if d.evaluateAbsentCase(workingType) {
		return
	}
	// Default: no pattern, no punch, no leave — treat as day off
	d.MarkDayOff()
}

// evaluateLeaveCase handles leave (full-day or half-day) scenarios.
// Returns true if leave logic was applied.
func (d *DailyAttendance) evaluateLeaveCase(leaveSubmissionID, leaveTypeName *string, leaveIsHalfDay *bool) bool {
	if leaveSubmissionID == nil {
		return false
	}
	if leaveIsHalfDay != nil && *leaveIsHalfDay && d.FirstPunchIn != nil {
		// Half-day leave but employee still clocked in
		d.LeaveSubmissionID = leaveSubmissionID
		d.LeaveTypeName = leaveTypeName
		d.MarkPresent()
		if d.LateMinutes() > 0 {
			d.IsLate = true
		}
		if d.LastPunchOut != nil && d.EarlyLeaveMinutes() > 0 {
			d.IsEarlyLeave = true
		}
		return true
	}
	// Full-day leave
	d.MarkOnLeave(*leaveSubmissionID, *leaveTypeName)
	return true
}

// evaluateOverrideDayOff handles scheduled day-off overrides.
// Returns true if the day-off was applied.
func (d *DailyAttendance) evaluateOverrideDayOff(overrideIsWorking *bool) bool {
	if overrideIsWorking != nil && !*overrideIsWorking {
		d.MarkDayOff()
		return true
	}
	return false
}

// evaluateWorkCase handles scenarios where the employee has clocked in.
// For "dynamic" working type, late/early checks are skipped (flexible hours).
// Returns true if work-case logic was applied.
func (d *DailyAttendance) evaluateWorkCase(workingType string) bool {
	if d.FirstPunchIn == nil {
		return false
	}
	d.MarkPresent()
	if workingType != "dynamic" {
		if d.LateMinutes() > 0 {
			d.IsLate = true
		}
		if d.LastPunchOut != nil && d.EarlyLeaveMinutes() > 0 {
			d.IsEarlyLeave = true
		}
	}
	return true
}

// evaluateAbsentCase handles scenarios where the employee is expected to work
// but did not clock in. For "dynamic" type without expected times, mark as no_punch
// (flexible hours, no shift end to compare against).
// For "fixed" type, if the shift has not ended yet, mark as no_punch;
// otherwise mark as absent.
// Returns true if absent/no_punch logic was applied.
func (d *DailyAttendance) evaluateAbsentCase(workingType string) bool {
	if d.Source != "working_pattern" && d.Source != "override" {
		return false
	}

	if d.ExpectedStartTime == nil || *d.ExpectedStartTime == "" {
		if workingType == "dynamic" {
			d.MarkNoPunch()
			return true
		}
		d.MarkDayOff()
		return true
	}

	if d.ExpectedEndTime != nil && *d.ExpectedEndTime != "" {
		loc := timeutil.LoadDefaultLocation()
		now := time.Now().In(loc)
		endParsed, err := time.Parse("15:04", *d.ExpectedEndTime)
		if err == nil {
			ref := time.Date(d.Date.Year(), d.Date.Month(), d.Date.Day(), endParsed.Hour(), endParsed.Minute(), 0, 0, loc)
			if now.Before(ref) {
				d.MarkNoPunch()
				return true
			}
		}
	}

	d.MarkAbsent()
	return true
}

// AdminScheduleFields holds raw schedule data from the admin attendance query
// for computing the resolved status when no daily_attendances row exists.
type AdminScheduleFields struct {
	Status             string
	Source             string
	Date               time.Time
	ExpectedStartTime  *string
	ExpectedEndTime    *string
	ScheduleOverrideID *string
	OverrideIsWorking  *bool
	OverrideStartTime  *string
	OverrideEndTime    *string
	PatternStartTime   *string
	PatternEndTime     *string
	PatternType        *string
}

// ResolveAdminAttendance fills in Status, Source, ExpectedStartTime,
// ExpectedEndTime, and ScheduleOverrideID from schedule fields when the
// daily_attendances row is absent.
func ResolveAdminAttendance(f *AdminScheduleFields) {
	if f.Status != "" {
		return
	}

	isPatternOff := f.PatternType != nil && *f.PatternType == "off"
	isPatternDynamic := f.PatternType != nil && *f.PatternType == "dynamic"

	if f.Source == "" {
		switch {
		case f.OverrideIsWorking != nil:
			f.Source = "override"
		case isPatternDynamic && !isPatternOff:
			f.Source = "working_pattern"
		case f.PatternStartTime != nil && !isPatternOff:
			f.Source = "working_pattern"
		default:
			f.Source = "none"
		}
	}

	if f.ExpectedStartTime == nil {
		if f.OverrideStartTime != nil && *f.OverrideStartTime != "" {
			f.ExpectedStartTime = f.OverrideStartTime
		} else if f.PatternStartTime != nil && !isPatternOff {
			f.ExpectedStartTime = f.PatternStartTime
		}
	}
	if f.ExpectedEndTime == nil {
		if f.OverrideEndTime != nil && *f.OverrideEndTime != "" {
			f.ExpectedEndTime = f.OverrideEndTime
		} else if f.PatternEndTime != nil && !isPatternOff {
			f.ExpectedEndTime = f.PatternEndTime
		}
	}

	switch {
	case f.OverrideIsWorking != nil && !*f.OverrideIsWorking:
		f.Status = string(AttendanceDayOff)
	case isPatternDynamic && !isPatternOff:
		f.Status = string(AttendanceNoPunch)
	case f.ExpectedStartTime != nil && *f.ExpectedStartTime != "" && !isPatternOff:
		if f.ExpectedEndTime != nil && *f.ExpectedEndTime != "" {
			loc := timeutil.LoadDefaultLocation()
			now := time.Now().In(loc)
			endParsed, err := time.Parse("15:04", *f.ExpectedEndTime)
			if err == nil {
				ref := time.Date(f.Date.Year(), f.Date.Month(), f.Date.Day(), endParsed.Hour(), endParsed.Minute(), 0, 0, loc)
				if now.After(ref) {
					f.Status = string(AttendanceAbsent)
				} else {
					f.Status = string(AttendanceNoPunch)
				}
			} else {
				f.Status = string(AttendanceNoPunch)
			}
		} else {
			f.Status = string(AttendanceNoPunch)
		}
	case f.OverrideIsWorking != nil && *f.OverrideIsWorking:
		f.Status = string(AttendanceNoPunch)
	default:
		f.Status = string(AttendanceDayOff)
	}
}
