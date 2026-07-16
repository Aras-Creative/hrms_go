package repository

import (
	"context"
	"time"
)

type MetricsCounts struct {
	TotalEmployees  int
	ActiveContracts int
	Present         int
	Late            int
	OnLeave         int
	PendingLeave    int
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
