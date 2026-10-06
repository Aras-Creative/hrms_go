package repository

import (
	"strings"
	"testing"
)

// Migration 000050 dropped effective_date and end_date from employee_compensations,
// employee_deductions and employee_benefits. Queries that still name those columns fail
// at run time with pq: column ... does not exist, which is how the period overview broke:
// nothing in the build or the unit tests touches those tables, so a stale reference only
// surfaces as a 500 on a live request.
//
// employee_base_salaries is deliberately excluded. It kept its validity window so a raise
// cannot reprice an earlier period, and these queries still rely on that.
func TestSetupQueriesAvoidDroppedColumns(t *testing.T) {
	queries := map[string]string{
		"insert employee compensation": qryInsertEmpComp,
		"select employee compensation": qrySelectEmpComp,
		"update employee compensation": qryUpdateEmpComp,
		"insert employee deduction":    qryInsertEmpDeduction,
		"select employee deduction":    qrySelectEmpDeduction,
		"update employee deduction":    qryUpdateEmpDeduction,
		"insert employee benefit":      qryInsertEmpBenefit,
		"select employee benefit":      qrySelectEmpBenefit,
		"update employee benefit":      qryUpdateEmpBenefit,
	}

	for name, query := range queries {
		t.Run(name, func(t *testing.T) {
			for _, dropped := range []string{"effective_date", "end_date"} {
				if strings.Contains(query, dropped) {
					t.Errorf("query references %s, dropped by migration 000050", dropped)
				}
			}
		})
	}
}

// The overview resolves each employee's period totals from the same setup rows the payslip
// processor reads, so it carries the same constraint: a validity-window filter here would
// reintroduce exactly the behaviour 000050 removed, silently excluding an item from the
// total while the slip that was generated from it included it.
func TestOverviewBatchQueriesAvoidDroppedColumns(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"compensations", overviewQueryCompensationRowsBatch},
		{"deductions", overviewQueryDeductionRowsBatch},
		{"employees", overviewQueryEmployees},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.query, "ec.effective_date") || strings.Contains(tt.query, "ec.end_date") {
				t.Error("compensations are filtered by a validity window, dropped by migration 000050")
			}
			if strings.Contains(tt.query, "ed.effective_date") || strings.Contains(tt.query, "ed.end_date") {
				t.Error("deductions are filtered by a validity window, dropped by migration 000050")
			}
		})
	}
}
