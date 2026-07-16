package repository

import "time"

type EmployeeBaseSalaryModel struct {
	ID            string    `db:"id"`
	EmployeeID    string    `db:"employee_id"`
	Amount        int64     `db:"amount"`
	Currency      string    `db:"currency"`
	EffectiveDate time.Time `db:"effective_date"`
	EndDate       *time.Time `db:"end_date"`
	Notes         string    `db:"notes"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type CompensationItemModel struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	ItemType    string    `db:"item_type"`
	Description string    `db:"description"`
	IsActive    bool      `db:"is_active"`
	IsTaxable   bool      `db:"is_taxable"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type EmployeeCompensationModel struct {
	ID                 string     `db:"id"`
	EmployeeID         string     `db:"employee_id"`
	CompensationItemID string     `db:"compensation_item_id"`
	Amount             int64      `db:"amount"`
	Frequency          string     `db:"frequency"`
	EffectiveDate      time.Time  `db:"effective_date"`
	EndDate            *time.Time `db:"end_date"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

type BenefitTypeModel struct {
	ID                        string    `db:"id"`
	Name                      string    `db:"name"`
	Description               string    `db:"description"`
	EmployerContributionType  string    `db:"employer_contribution_type"`
	EmployerContributionValue float64   `db:"employer_contribution_value"`
	EmployeeContributionType  string    `db:"employee_contribution_type"`
	EmployeeContributionValue float64   `db:"employee_contribution_value"`
	IsActive                  bool      `db:"is_active"`
	CreatedAt                 time.Time `db:"created_at"`
	UpdatedAt                 time.Time `db:"updated_at"`
}

type EmployeeBenefitModel struct {
	ID                string     `db:"id"`
	EmployeeID        string     `db:"employee_id"`
	BenefitTypeID     string     `db:"benefit_type_id"`
	ParticipantNumber string     `db:"participant_number"`
	EffectiveDate     time.Time  `db:"effective_date"`
	EndDate           *time.Time `db:"end_date"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

type DeductionTypeModel struct {
	ID            string    `db:"id"`
	Name          string    `db:"name"`
	Slug          string    `db:"slug"`
	Description   string    `db:"description"`
	DeductionType string    `db:"deduction_type"`
	DefaultValue  float64   `db:"default_value"`
	IsActive      bool      `db:"is_active"`
	IsMandatory   bool      `db:"is_mandatory"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type EmployeeDeductionModel struct {
	ID              string     `db:"id"`
	EmployeeID      string     `db:"employee_id"`
	DeductionTypeID string     `db:"deduction_type_id"`
	Value           *float64   `db:"value"`
	EffectiveDate   time.Time  `db:"effective_date"`
	EndDate         *time.Time `db:"end_date"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}
