package delivery

import (
	"time"

	"hrms/internal/attendance/entity"
	"hrms/internal/attendance/models"
	"hrms/internal/pkg/timeutil"
)

type PunchResponse struct {
	ID         string    `json:"id"`
	EmployeeID string    `json:"employee_id"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	CreatedAt  time.Time `json:"created_at"`
}

func punchToResponse(p *entity.Punch) PunchResponse {
	return PunchResponse{ID: p.ID, EmployeeID: p.EmployeeID, Type: string(p.Type), Timestamp: p.Timestamp, CreatedAt: p.CreatedAt}
}

func punchListToResponse(list []*entity.Punch) []PunchResponse {
	resp := make([]PunchResponse, 0, len(list))
	for _, p := range list {
		resp = append(resp, punchToResponse(p))
	}
	return resp
}

type DailyResponse struct {
	ID                 string     `json:"id"`
	EmployeeID         string     `json:"employee_id"`
	Date               time.Time  `json:"date"`
	Status             string     `json:"status"`
	IsLate             bool       `json:"is_late"`
	IsEarlyLeave       bool       `json:"is_early_leave"`
	ExpectedStartTime  *string    `json:"expected_start_time,omitempty"`
	ExpectedEndTime    *string    `json:"expected_end_time,omitempty"`
	Source             string     `json:"source"`
	FirstPunchIn       *time.Time `json:"first_punch_in,omitempty"`
	LastPunchOut       *time.Time `json:"last_punch_out,omitempty"`
	TotalWorkSeconds   *int       `json:"total_work_seconds,omitempty"`
	LateMinutes        int        `json:"late_minutes"`
	LeaveSubmissionID  *string    `json:"leave_submission_id,omitempty"`
	LeaveTypeName      *string    `json:"leave_type_name,omitempty"`
	ScheduleOverrideID *string    `json:"schedule_override_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func dailyToResponse(da *entity.DailyAttendance) DailyResponse {
	return DailyResponse{
		ID: da.ID, EmployeeID: da.EmployeeID, Date: da.Date, Status: string(da.Status),
		IsLate: da.IsLate, IsEarlyLeave: da.IsEarlyLeave,
		ExpectedStartTime: da.ExpectedStartTime, ExpectedEndTime: da.ExpectedEndTime,
		Source: da.Source, FirstPunchIn: da.FirstPunchIn, LastPunchOut: da.LastPunchOut,
		TotalWorkSeconds: da.TotalWorkSeconds, LeaveSubmissionID: da.LeaveSubmissionID,
		LeaveTypeName: da.LeaveTypeName, ScheduleOverrideID: da.ScheduleOverrideID,
		CreatedAt: da.CreatedAt, UpdatedAt: da.UpdatedAt, LateMinutes: da.LateMinutes(),
	}
}

// ---- Admin Attendance ----

type AdminAttendanceResponse struct {
	ID                 *string    `json:"id,omitempty"`
	EmployeeID         string     `json:"employee_id"`
	EmployeeName       string     `json:"employee_name"`
	EmployeeNumber     string     `json:"employee_number"`
	DesignationName    *string    `json:"designation_name,omitempty"`
	ProfilePhotoID     *string    `json:"profile_photo_id,omitempty"`
	ProfilePhotoURL    string     `json:"profile_photo_url,omitempty"`
	Date               time.Time  `json:"date"`
	Status             string     `json:"status"`
	IsLate             bool       `json:"is_late"`
	IsEarlyLeave       bool       `json:"is_early_leave"`
	IsOverdue          bool       `json:"is_overdue"`
	ExpectedStartTime  *string    `json:"expected_start_time,omitempty"`
	ExpectedEndTime    *string    `json:"expected_end_time,omitempty"`
	Source             string     `json:"source"`
	FirstPunchIn       *time.Time `json:"first_punch_in,omitempty"`
	LastPunchOut       *time.Time `json:"last_punch_out,omitempty"`
	TotalWorkSeconds   *int       `json:"total_work_seconds,omitempty"`
	LateMinutes        int        `json:"late_minutes"`
	LeaveSubmissionID  *string    `json:"leave_submission_id,omitempty"`
	LeaveTypeName      *string    `json:"leave_type_name,omitempty"`
	ScheduleOverrideID *string    `json:"schedule_override_id,omitempty"`
	CreatedAt          *time.Time `json:"created_at,omitempty"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
}

func computeLateMinutes(date time.Time, firstPunchIn *time.Time, expectedStartTime *string) int {
	if firstPunchIn == nil || expectedStartTime == nil || *expectedStartTime == "" {
		return 0
	}
	expected, err := time.Parse("15:04", *expectedStartTime)
	if err != nil {
		return 0
	}
	loc := timeutil.LoadDefaultLocation()
	punch := firstPunchIn.In(loc)
	ref := time.Date(punch.Year(), punch.Month(), punch.Day(), expected.Hour(), expected.Minute(), 0, 0, loc)
	if !punch.After(ref) {
		return 0
	}
	return int(punch.Sub(ref).Minutes())
}

// computeIsOverdue returns true when the employee has not punched in and the
// current time is past the expected_start_time.  This is computed on-the-fly
// in the response layer (not stored) so admins see a real-time "not punched,
// past entry time" signal without destroying the no_punch status that is
// awaiting the finalize cutoff.
func computeIsOverdue(status string, date time.Time, expectedStartTime *string, firstPunchIn *time.Time) bool {
	if status != "no_punch" {
		return false
	}
	if firstPunchIn != nil {
		return false
	}
	if expectedStartTime == nil || *expectedStartTime == "" {
		return false
	}
	startParsed, err := time.Parse("15:04", *expectedStartTime)
	if err != nil {
		return false
	}
	loc := timeutil.LoadDefaultLocation()
	ref := time.Date(date.Year(), date.Month(), date.Day(), startParsed.Hour(), startParsed.Minute(), 0, 0, loc)
	return time.Now().In(loc).After(ref)
}

func adminAttendanceToResponse(row *models.AdminAttendanceItem) AdminAttendanceResponse {
	return AdminAttendanceResponse{
		ID:                 row.ID,
		EmployeeID:         row.EmployeeID,
		EmployeeName:       row.EmployeeName,
		EmployeeNumber:     row.EmployeeNumber,
		DesignationName:    row.DesignationName,
		ProfilePhotoID:     row.ProfilePhotoID,
		Date:               row.Date,
		Status:             row.Status,
		IsLate:             row.IsLate,
		IsEarlyLeave:       row.IsEarlyLeave,
		IsOverdue:          computeIsOverdue(row.Status, row.Date, row.ExpectedStartTime, row.FirstPunchIn),
		ExpectedStartTime:  row.ExpectedStartTime,
		ExpectedEndTime:    row.ExpectedEndTime,
		Source:             row.Source,
		FirstPunchIn:       row.FirstPunchIn,
		LastPunchOut:       row.LastPunchOut,
		TotalWorkSeconds:   row.TotalWorkSeconds,
		LateMinutes:        computeLateMinutes(row.Date, row.FirstPunchIn, row.ExpectedStartTime),
		LeaveSubmissionID:  row.LeaveSubmissionID,
		LeaveTypeName:      row.LeaveTypeName,
		ScheduleOverrideID: row.ScheduleOverrideID,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

// ---- Correction ----

type CorrectionResponse struct {
	ID               string     `json:"id"`
	EmployeeID       string     `json:"employee_id"`
	Date             time.Time  `json:"date"`
	ClockIn          *time.Time `json:"clock_in,omitempty"`
	ClockOut         *time.Time `json:"clock_out,omitempty"`
	Status           *string    `json:"status,omitempty"`
	IsLate           *bool      `json:"is_late,omitempty"`
	IsEarlyLeave     *bool      `json:"is_early_leave,omitempty"`
	LeaveTypeName    *string    `json:"leave_type_name,omitempty"`
	LeaveSubmissionID *string   `json:"leave_submission_id,omitempty"`
	Reason           string     `json:"reason"`
	CorrectedBy      string     `json:"corrected_by"`
	CreatedAt        time.Time  `json:"created_at"`
}

func correctionToResponse(c *entity.AttendanceCorrection) CorrectionResponse {
	return CorrectionResponse{
		ID: c.ID, EmployeeID: c.EmployeeID, Date: c.Date,
		ClockIn: c.ClockIn, ClockOut: c.ClockOut, Status: c.Status,
		IsLate: c.IsLate, IsEarlyLeave: c.IsEarlyLeave,
		LeaveTypeName: c.LeaveTypeName, LeaveSubmissionID: c.LeaveSubmissionID,
		Reason: c.Reason, CorrectedBy: c.CorrectedBy, CreatedAt: c.CreatedAt,
	}
}

type CorrectionViewResponse struct {
	ID               string     `json:"id"`
	EmployeeID       string     `json:"employee_id"`
	EmployeeName     string     `json:"employee_name"`
	Date             time.Time  `json:"date"`
	ClockIn          *time.Time `json:"clock_in,omitempty"`
	ClockOut         *time.Time `json:"clock_out,omitempty"`
	Status           *string    `json:"status,omitempty"`
	IsLate           *bool      `json:"is_late,omitempty"`
	IsEarlyLeave     *bool      `json:"is_early_leave,omitempty"`
	LeaveTypeName    *string    `json:"leave_type_name,omitempty"`
	LeaveSubmissionID *string   `json:"leave_submission_id,omitempty"`
	Reason           string     `json:"reason"`
	CorrectedBy      string     `json:"corrected_by"`
	CreatedAt        time.Time  `json:"created_at"`
}

func correctionViewToResponse(c *models.CorrectionViewItem) CorrectionViewResponse {
	return CorrectionViewResponse{
		ID: c.ID, EmployeeID: c.EmployeeID, EmployeeName: c.EmployeeName,
		Date: c.Date, ClockIn: c.ClockIn, ClockOut: c.ClockOut,
		Status: c.Status, IsLate: c.IsLate, IsEarlyLeave: c.IsEarlyLeave,
		LeaveTypeName: c.LeaveTypeName, LeaveSubmissionID: c.LeaveSubmissionID,
		Reason: c.Reason, CorrectedBy: c.CorrectedBy, CreatedAt: c.CreatedAt,
	}
}

func correctionViewListToResponse(list []*models.CorrectionViewItem) []CorrectionViewResponse {
	resp := make([]CorrectionViewResponse, 0, len(list))
	for _, c := range list {
		resp = append(resp, correctionViewToResponse(c))
	}
	return resp
}

// ---- Attendance Detail ----

type AttendanceDetailResponse struct {
	Attendance  DailyResponse          `json:"attendance"`
	Corrections []CorrectionResponse   `json:"corrections"`
	Punches     []PunchResponse        `json:"punches"`
	AuditLogs   []AuditLogEntryResponse `json:"audit_logs"`
}

type AuditLogEntryResponse struct {
	ID        string         `json:"id"`
	Action    string         `json:"action"`
	ActorID   string         `json:"actor_id"`
	ActorName string         `json:"actor_name"`
	Payload   map[string]any `json:"payload"`
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"user_agent"`
	CreatedAt time.Time      `json:"created_at"`
}

// ---- Recap ----

type RecapHeader struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type RecapMetric struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

type RecapEmployee struct {
	ID                      string        `json:"id"`
	EmployeeNumber          string        `json:"employee_number"`
	Name                    string        `json:"name"`
	ProfilePictureURL       *string       `json:"profile_picture_url"`
	Department              *string       `json:"department"`
	WorkingDays             int           `json:"working_days"`
	LateMinutes             int           `json:"late_minutes"`
	AttendanceMetrics       []RecapMetric `json:"attendance_metrics"`
	TotalAttendanceIncident int           `json:"total_attendance_incident"`
}

type RecapResponse struct {
	Headers   []RecapHeader    `json:"headers"`
	Employees []RecapEmployee  `json:"employees"`
}
