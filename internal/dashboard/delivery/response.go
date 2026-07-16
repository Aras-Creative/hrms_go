package delivery

import (
	"time"

	"hrms/internal/dashboard/entity"
)

type DashboardMetricsResponse struct {
	TotalEmployees  int       `json:"total_employees"`
	ActiveContracts int       `json:"active_contracts"`
	Present         int       `json:"present"`
	PresentPct      float64   `json:"present_pct"`
	Late            int       `json:"late"`
	LatePct         float64   `json:"late_pct"`
	OnLeave         int       `json:"on_leave"`
	PendingLeave    int       `json:"pending_leave"`
	From            time.Time `json:"from"`
	To              time.Time `json:"to"`
}

type AttendanceTrendResponse struct {
	Date    string `json:"date"`
	Present int    `json:"present"`
	OnLeave int    `json:"on_leave"`
	Absent  int    `json:"absent"`
}

func toMetricsResponse(m *entity.DashboardMetrics) DashboardMetricsResponse {
	return DashboardMetricsResponse{
		TotalEmployees:  m.TotalEmployees,
		ActiveContracts: m.ActiveContracts,
		Present:         m.Present,
		PresentPct:      m.PresentPct,
		Late:            m.Late,
		LatePct:         m.LatePct,
		OnLeave:         m.OnLeave,
		PendingLeave:    m.PendingLeave,
		From:            m.From,
		To:              m.To,
	}
}

func toTrendsResponse(trends []entity.AttendanceTrend) []AttendanceTrendResponse {
	resp := make([]AttendanceTrendResponse, 0, len(trends))
	for _, t := range trends {
		resp = append(resp, AttendanceTrendResponse{
			Date: t.Date, Present: t.Present,
			OnLeave: t.OnLeave, Absent: t.Absent,
		})
	}
	return resp
}
