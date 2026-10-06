package usecase

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"hrms/internal/payroll/repository"
)

// Migration 000050 dropped effective_date/end_date from the payroll setup tables, and the
// rewrite that came with it mangled the insert statements it touched: a dropped column
// glued onto its neighbour ("participant_numbercreated_at") and placeholder counts left
// matching the old column count. None of it fails at build time. Every failure waits for
// the statement to reach postgres, which for the setup inserts means a 500 on saving an
// employee's payroll package.
//
// The checks below are pure string and reflection checks for the same reason
// query_placeholder_test.go exists in the repository package: these statements cannot run
// without a database, and a typo that postgres would reject is invisible until then.
var (
	reInsertColumns = regexp.MustCompile(`(?s)INSERT INTO \w+ \((.*?)\)`)
	rePlaceholder   = regexp.MustCompile(`\$\d+`)
)

// dbColumns lists the columns a repository model knows about, which is the closest thing
// to the table definition available without a database: every row scan of that table binds
// these names, so a column absent here does not exist for this code.
func dbColumns(model interface{}) map[string]bool {
	t := reflect.TypeOf(model)
	cols := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag != "" {
			cols[tag] = true
		}
	}
	return cols
}

func TestSetupInsertQueriesMatchTheirColumns(t *testing.T) {
	// args mirrors what each helper passes to ExecContext, in order.
	tests := []struct {
		name  string
		query string
		args  int
		model interface{}
	}{
		{"base salary", qryInsertSalary, 9, repository.EmployeeBaseSalaryModel{}},
		{"compensation", qryInsertEmpComp, 8, repository.EmployeeCompensationModel{}},
		{"benefit", qryInsertEmpBenefit, 6, repository.EmployeeBenefitModel{}},
		{"deduction", qryInsertEmpDeduction, 7, repository.EmployeeDeductionModel{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols := reInsertColumns.FindStringSubmatch(tt.query)
			if cols == nil {
				t.Fatal("query has no INSERT column list to check")
			}
			columns := strings.Split(cols[1], ",")

			known := dbColumns(tt.model)
			for _, c := range columns {
				name := strings.TrimSpace(c)
				if name == "" {
					t.Fatalf("query has an empty column name, which is what a dropped column leaves behind: %s", tt.query)
				}
				if !known[name] {
					t.Errorf("column %q is not a column of the model; a removed column glued to a neighbour is the usual cause", name)
				}
			}

			placeholders := len(rePlaceholder.FindAllString(tt.query, -1))
			if placeholders != len(columns) {
				t.Errorf("placeholder count = %d, column count = %d", placeholders, len(columns))
			}
			if placeholders != tt.args {
				t.Errorf("placeholder count = %d, args passed by the caller = %d", placeholders, tt.args)
			}
		})
	}
}
