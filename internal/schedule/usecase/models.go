package usecase

import (
	"fmt"
	"time"
)

var dateLayouts = []string{"2006-01-02", time.RFC3339}

func tryParseDate(s string) (time.Time, error) {
	for _, layout := range dateLayouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

type WorkingPatternDetailInput struct {
	DayOfWeek int
	Type      string
	StartTime *string
	EndTime   *string
}

type CreateWorkingPatternInput struct {
	Name        string
	Description *string
	Details     []WorkingPatternDetailInput
}

type UpdateWorkingPatternInput struct {
	Name        string
	Description *string
	Details     []WorkingPatternDetailInput
}

type AssignPatternInput struct {
	EmployeeIDs   []string
	WorkPatternID string
	ValidFrom     string
	ValidTo       *time.Time
}

type AssignResult struct {
	Succeeded []string         `json:"succeeded"`
	Failed    []AssignError    `json:"failed"`
}

type AssignError struct {
	EmployeeID string `json:"employee_id"`
	Error      string `json:"error"`
}

type ScheduleOverviewParams struct {
	EmployeeID    string
	DesignationID string
	Search        string
	From          string
	To            string
}

type ScheduleSource string

const (
	SourceOverride       ScheduleSource = "override"
	SourceWorkingPattern ScheduleSource = "working_pattern"
	SourceNoPattern      ScheduleSource = "no_pattern"
	SourceNone           ScheduleSource = "none"
)

type ScheduleOverviewDate struct {
	Date         string         `json:"date"`
	Source       ScheduleSource `json:"source"`
	IsWorkingDay bool           `json:"is_working_day"`
	WorkingType  *string        `json:"working_type,omitempty"`
	StartTime    *string        `json:"start_time"`
	EndTime      *string        `json:"end_time"`
	Notes        *string        `json:"notes"`
	OverrideID   *string        `json:"override_id"`
}

type ScheduleOverviewEmployee struct {
	EmployeeID         string                 `json:"employee_id"`
	EmployeeName       string                 `json:"employee_name"`
	EmployeeNumber     string                 `json:"employee_number"`
	DesignationName    *string                `json:"designation_name"`
	WorkingPatternID   *string                `json:"working_pattern_id"`
	WorkingPatternName *string                `json:"working_pattern_name"`
	Dates              []ScheduleOverviewDate `json:"dates"`
}
