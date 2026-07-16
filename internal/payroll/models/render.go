package models

type PayslipEmployeeData struct {
	FullName       string
	EmployeeNumber string
	DesignationName string
	Status         string
	BankName       string
	BankNumber     string
}

type PayslipRenderData struct {
	LogoURL         string
	CompanyName    string
	CompanyAddress string
	DocNumber      string
	PeriodName     string
	PeriodRange    string

	EmployeeName       string
	EmployeeNumber     string
	DesignationName    string
	Status             string
	BankInfo           string

	BaseSalary      string
	Compensations   []BreakdownRow
	TotalIncome     string
	Deductions      []BreakdownRow
	TotalDeductions string
	AbsentDays      int
	NetSalary       string
}

type BreakdownRow struct {
	Name   string
	Amount string
}
