package repository

import (
	"context"
	"time"
)

type MetricsCounts struct {
	TotalEmployees  int `db:"total_employees"`
	ActiveContracts int `db:"active_contracts"`
	Present         int `db:"present"`
	Absent          int `db:"absent"`
	Late            int `db:"late"`
	OnLeave         int `db:"on_leave"`
}

type TrendRow struct {
	Date    string `db:"date"`
	Present int    `db:"present"`
	OnLeave int    `db:"on_leave"`
	Absent  int    `db:"absent"`
}

type DashboardRepository interface {
	CountMetrics(ctx context.Context, from, to time.Time) (*MetricsCounts, error)
	GetTrends(ctx context.Context, from, to time.Time) ([]TrendRow, error)
}
