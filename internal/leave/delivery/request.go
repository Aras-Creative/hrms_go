package delivery

type CreateLeaveTypeRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	DefaultDays int    `json:"default_days"`
	IsPaid      bool   `json:"is_paid"`
	IsUnlimited bool   `json:"is_unlimited"`
	IsHalfDay   bool   `json:"is_half_day"`
}

type UpdateLeaveTypeRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1,max=255"`
	DefaultDays int    `json:"default_days"`
	IsPaid      bool   `json:"is_paid"`
	IsUnlimited bool   `json:"is_unlimited"`
	IsHalfDay   bool   `json:"is_half_day"`
}

type UpdateLeaveBalanceRequest struct {
	TotalDays int `json:"total_days"`
}

type CreateLeaveSubmissionRequest struct {
	LeaveTypeID  string  `json:"leave_type_id" validate:"required,uuid"`
	StartDate    string  `json:"start_date" validate:"required"`
	EndDate      string  `json:"end_date" validate:"required"`
	Reason       string  `json:"reason"`
	AttachmentID *string `json:"attachment_id"`
}
