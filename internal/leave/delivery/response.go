package delivery

import "time"

type AuditEntryResponse struct {
	ID        string         `json:"id"`
	Action    string         `json:"action"`
	ActorID   string         `json:"actor_id"`
	ActorName string         `json:"actor_name"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type LeaveBalanceResponse struct {
	ID               string    `json:"id"`
	EmployeeID       string    `json:"employee_id"`
	EmployeeName     string    `json:"employee_name,omitempty"`
	EmployeeNumber   string    `json:"employee_number,omitempty"`
	ProfilePhotoID   *string   `json:"profile_photo_id,omitempty"`
	ProfilePhotoURL  string    `json:"profile_photo_url,omitempty"`
	LeaveTypeID      string    `json:"leave_type_id"`
	LeaveTypeName    string    `json:"leave_type_name,omitempty"`
	Year             int       `json:"year"`
	TotalDays        int       `json:"total_days"`
	UsedDays         int       `json:"used_days"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type LeaveSubmissionResponse struct {
	ID              string     `json:"id"`
	EmployeeID      string     `json:"employee_id"`
	EmployeeName    string     `json:"employee_name,omitempty"`
	EmployeeNumber  string     `json:"employee_number,omitempty"`
	ProfilePhotoID  *string    `json:"profile_photo_id,omitempty"`
	ProfilePhotoURL string     `json:"profile_photo_url,omitempty"`
	LeaveTypeID     string     `json:"leave_type_id"`
	LeaveTypeName   string     `json:"leave_type_name,omitempty"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	Days            int        `json:"days"`
	Reason          string     `json:"reason"`
	AttachmentID    *string    `json:"attachment_id,omitempty"`
	Status          string     `json:"status"`
	ApprovedBy      *string    `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type AttachmentResponse struct {
	URL string `json:"url"`
}

type SubmissionDetailResponse struct {
	Submission LeaveSubmissionResponse `json:"submission"`
	History    []AuditEntryResponse    `json:"history"`
}

type LeaveTypeResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DefaultDays int       `json:"default_days"`
	IsPaid      bool      `json:"is_paid"`
	IsUnlimited bool      `json:"is_unlimited"`
	IsHalfDay   bool      `json:"is_half_day"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
