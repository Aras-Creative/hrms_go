package models

type PayslipEmployeeData struct {
	FullName        string
	EmployeeNumber  string
	DesignationName string
	Status          string
	BankName        string
	BankNumber      string
}

type PayslipRenderData struct {
	LogoURL        string
	CompanyName    string
	CompanyAddress string
	DocNumber      string
	PeriodName     string
	PeriodRange    string

	EmployeeName    string
	EmployeeNumber  string
	DesignationName string
	Status          string
	BankInfo        string

	BaseSalary      string
	Compensations   []BreakdownRow
	TotalIncome     string
	Deductions      []BreakdownRow
	TotalDeductions string
	AbsentDays      int
	NetSalary       string

	// IncomeNotes are the descriptive figures an admin recorded on this slip. They sit
	// outside every total on purpose, which is why they render below the net-pay block
	// instead of inside the income or deduction table.
	IncomeNotes []IncomeNoteRow
}

// IncomeNoteRow is one rendered line of the descriptive-figures section. Value is already
// formatted for display, unit and all, because the unit is a display hint that no
// calculation reads.
type IncomeNoteRow struct {
	Label string
	Value string
	Notes string
}

type BreakdownRow struct {
	Name   string
	Amount string
}
