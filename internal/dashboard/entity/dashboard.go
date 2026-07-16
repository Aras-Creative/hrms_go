package entity

import "time"

type DashboardMetrics struct {
	TotalEmployees  int        `json:"total_employees"`
	ActiveContracts int        `json:"active_contracts"`
	Present         int        `json:"present"`
	PresentPct      float64    `json:"present_pct"`
	Late            int        `json:"late"`
	LatePct         float64    `json:"late_pct"`
	OnLeave         int        `json:"on_leave"`
	PendingLeave    int        `json:"pending_leave"`
	From            time.Time  `json:"from"`
	To              time.Time  `json:"to"`
}

type AttendanceTrend struct {
	Date    string `json:"date"`
	Present int    `json:"present"`
	OnLeave int    `json:"on_leave"`
	Absent  int    `json:"absent"`
}
