package delivery

type WorkingPatternDetailRequest struct {
	DayOfWeek int     `json:"day_of_week" validate:"required,min=0,max=6"`
	Type      string  `json:"type" validate:"required,oneof=fixed dynamic off"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
}

type CreateWorkingPatternRequest struct {
	Name        string                       `json:"name" validate:"required,min=1,max=255"`
	Description *string                      `json:"description"`
	Details     []WorkingPatternDetailRequest `json:"details"`
}

type UpdateWorkingPatternRequest struct {
	Name        string                       `json:"name" validate:"omitempty,min=1,max=255"`
	Description *string                      `json:"description"`
	Details     []WorkingPatternDetailRequest `json:"details"`
}

type AssignPatternRequest struct {
	EmployeeIDs   []string `json:"employee_ids" validate:"required,min=1"`
	WorkPatternID string   `json:"work_pattern_id" validate:"required"`
	ValidFrom     string   `json:"valid_from,omitempty"`
	ValidTo       *string  `json:"valid_to,omitempty"`
}

type SetOverrideRequest struct {
	EmployeeID   string  `json:"employee_id" validate:"required"`
	DateFrom     string  `json:"date_from" validate:"required"`
	DateTo       string  `json:"date_to" validate:"required"`
	IsWorkingDay bool    `json:"is_working_day"`
	StartTime    *string `json:"start_time,omitempty"`
	EndTime      *string `json:"end_time,omitempty"`
	Reason       *string `json:"reason,omitempty"`
}
