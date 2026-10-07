package delivery

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestSetupBaseSalaryAmountValidation(t *testing.T) {
	validate := validator.New()
	for _, tc := range []struct {
		name    string
		amount  string
		want    float64
		wantErr bool
	}{
		{"zero number", `0`, 0, false},
		{"zero string", `"0"`, 0, false},
		{"positive amount", `100`, 100, false},
		{"null amount", `null`, 0, true},
		{"missing amount", ``, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var req SetupEmployeePayrollRequest
			body := `{"employee_id":"123e4567-e89b-12d3-a456-426614174000","base_salary":{`
			if tc.amount != "" {
				body += `"amount":` + tc.amount
			}
			body += `}}`
			if err := json.Unmarshal([]byte(body), &req); err != nil {
				t.Fatal(err)
			}
			err := validate.Struct(req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validation error = %v, want error = %v", err, tc.wantErr)
			}
			if !tc.wantErr && (req.BaseSalary.Amount == nil || float64(*req.BaseSalary.Amount) != tc.want) {
				t.Fatalf("valid amount was not preserved: %+v", req.BaseSalary.Amount)
			}
		})
	}
}
