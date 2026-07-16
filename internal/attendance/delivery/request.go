package delivery

import "time"

type CreateCorrectionRequest struct {
	EmployeeID string     `json:"employee_id"`
	Date       string     `json:"date"`
	ClockIn    *time.Time `json:"clock_in"`
	ClockOut   *time.Time `json:"clock_out"`
	Status     *string    `json:"status"`
	Reason     string     `json:"reason"`
}

type PunchTodayQuery struct {
	EmployeeID string `query:"employee_id"`
}

type PunchHistoryQuery struct {
	EmployeeID string `query:"employee_id"`
	From       string `query:"from"`
	To         string `query:"to"`
}

type DailyQueryRequest struct {
	EmployeeID string `query:"employee_id"`
	From       string `query:"from"`
	To         string `query:"to"`
}

type AdminListRequest struct {
	SearchName    string `query:"search"`
	Status        string `query:"status"`
	DesignationID string `query:"designation_id"`
	IsLate        string `query:"is_late"`
	IsEarlyLeave  string `query:"is_early_leave"`
	From          string `query:"from"`
	To            string `query:"to"`
	Page          int    `query:"page"`
	PerPage       int    `query:"per_page"`
}

type RecapQueryRequest struct {
	From          string `query:"from"`
	To            string `query:"to"`
	DesignationID string `query:"designation_id"`
}

type MyHistoryQuery struct {
	From string `query:"from"`
	To   string `query:"to"`
}

type CorrectionListQuery struct {
	SearchName string `query:"search"`
	StartDate  string `query:"start_date"`
	EndDate    string `query:"end_date"`
	Page       int    `query:"page"`
	PerPage    int    `query:"per_page"`
}
