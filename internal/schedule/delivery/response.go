package delivery

import "time"

type WorkingPatternDetailResponse struct {
	DayOfWeek int     `json:"day_of_week"`
	Type      string  `json:"type"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
}

type WorkPatternResponse struct {
	ID          string                        `json:"id"`
	Name        string                        `json:"name"`
	Description *string                       `json:"description,omitempty"`
	IsActive    bool                          `json:"is_active"`
	Details     []WorkingPatternDetailResponse `json:"details"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
}

type EmployeeWorkPatternResponse struct {
	ID            string    `json:"id"`
	EmployeeID    string    `json:"employee_id"`
	WorkPatternID string    `json:"work_pattern_id"`
	ValidFrom     string    `json:"valid_from"`
	ValidTo       *string   `json:"valid_to,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OverrideResponse struct {
	ID           string    `json:"id"`
	EmployeeID   string    `json:"employee_id"`
	Date         string    `json:"date"`
	IsWorkingDay bool      `json:"is_working_day"`
	StartTime    *string   `json:"start_time,omitempty"`
	EndTime      *string   `json:"end_time,omitempty"`
	Reason       *string   `json:"reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
