package repository

import (
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

// sqlx.In decides how many bind variables to expect by counting ? placeholders, so a
// query written with postgres-style $1 placeholders expands to zero bind vars and every
// call fails with "number of bindVars less than number arguments". That regression took
// down the whole payroll processor, and it is invisible to unit tests that stub the
// database, so the placeholder style is asserted here instead.
func TestBatchQueriesExpandForSqlxIn(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		query string
		args  []interface{}
	}{
		{
			name:  "attendance day counts, single employee",
			query: qryCalcAttendanceDayCounts,
			args:  []interface{}{[]string{"emp-1"}, start, end},
		},
		{
			name:  "attendance day counts, many employees",
			query: qryCalcAttendanceDayCounts,
			args:  []interface{}{[]string{"emp-1", "emp-2", "emp-3"}, start, end},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.query, "$1") {
				t.Fatalf("query uses postgres-style $1 placeholders but is expanded by sqlx.In")
			}
			if !strings.Contains(tt.query, "?") {
				t.Fatal("query has no ? placeholder for sqlx.In to expand")
			}

			_, args, err := sqlx.In(tt.query, tt.args...)
			if err != nil {
				t.Fatalf("sqlx.In failed: %v", err)
			}
			// After expansion the slice contributes one arg per element.
			want := len(tt.args)
			if ids, ok := tt.args[0].([]string); ok {
				want = len(ids) + len(tt.args) - 1
			}
			if len(args) != want {
				t.Errorf("expanded arg count = %d, want %d", len(args), want)
			}
		})
	}
}

// The same guard for the working-days query, which carries ten placeholders: two IN (?)
// id lists plus start/end repeated across the pattern window, the override window, the
// date filter and the generate_series bounds. This assertion is what caught the mismatch.
func TestWorkingDaysQueryExpandsForSqlxIn(t *testing.T) {
	ids := []string{"emp-1", "emp-2"}
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	if strings.Contains(qryEmployeeWorkingDaysBatch, "$1") {
		t.Fatal("query uses postgres-style $1 placeholders but is expanded by sqlx.In")
	}

	args := []interface{}{ids, end, start, ids, start, end, start, end, start, end}
	_, expanded, err := sqlx.In(qryEmployeeWorkingDaysBatch, args...)
	if err != nil {
		t.Fatalf("sqlx.In failed: %v", err)
	}
	// ids is bound twice (work-pattern CTE and override CTE), so each slice expands
	// to len(ids) arguments; the remaining eight placeholders are scalars.
	expected := 2*len(ids) + 8
	if len(expanded) != expected {
		t.Errorf("expanded arg count = %d, want %d", len(expanded), expected)
	}
}
