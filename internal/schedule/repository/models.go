package repository

import "time"

type WorkPatternModel struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	IsActive    bool       `db:"is_active"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type WorkPatternDetailModel struct {
	ID               string  `db:"id"`
	WorkingPatternID string  `db:"work_pattern_id"`
	DayOfWeek        int     `db:"day_of_week"`
	Type             string  `db:"working_type"`
	StartTime        *string `db:"start_time"`
	EndTime          *string `db:"end_time"`
}

type EmployeeScheduleOverrideModel struct {
	ID           string    `db:"id"`
	EmployeeID   string    `db:"employee_id"`
	Date         time.Time `db:"date"`
	IsWorkingDay bool      `db:"is_working_day"`
	StartTime    *string   `db:"start_time"`
	EndTime      *string   `db:"end_time"`
	Reason       *string   `db:"reason"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type EmployeeWorkPatternModel struct {
	ID            string     `db:"id"`
	EmployeeID    string     `db:"employee_id"`
	WorkPatternID string     `db:"work_pattern_id"`
	ValidFrom     time.Time  `db:"valid_from"`
	ValidTo       *time.Time `db:"valid_to"`
	IsActive      bool       `db:"is_active"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

type ScheduleOverviewParams struct {
	EmployeeID    string
	DesignationID string
	Search        string
	From          time.Time
	To            time.Time
}

type ScheduleOverviewRow struct {
	EmployeeID         string    `db:"employee_id"`
	FullName           string    `db:"full_name"`
	EmployeeNumber     string    `db:"employee_number"`
	DesignationName    *string   `db:"designation_name"`
	WorkPatternID      *string   `db:"work_pattern_id"`
	WorkingPatternName *string   `db:"working_pattern_name"`
	Date               time.Time `db:"date"`
	DayOfWeek          int       `db:"day_of_week"`
	PatternDetailID    *string   `db:"pattern_detail_id"`
	OverrideID         *string   `db:"override_id"`
	OverrideIsWorking  *bool     `db:"override_is_working"`
	OverrideStartTime  *string   `db:"override_start_time"`
	OverrideEndTime    *string   `db:"override_end_time"`
	OverrideNotes      *string   `db:"override_notes"`
	PatternType        *string   `db:"pattern_working_type"`
	PatternStartTime   *string   `db:"pattern_start_time"`
	PatternEndTime     *string   `db:"pattern_end_time"`
}
