package delivery

import (
	"time"

	"hrms/internal/payroll/entity"
)

type BaseSalaryResponse struct {
	ID            string     `json:"id"`
	EmployeeID    string     `json:"employee_id"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	EffectiveDate time.Time  `json:"effective_date"`
	EndDate       *time.Time `json:"end_date"`
	Notes         string     `json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CompensationItemResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ItemType    string    `json:"item_type"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	IsTaxable   bool      `json:"is_taxable"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type EmployeeCompensationResponse struct {
	ID                 string     `json:"id"`
	EmployeeID         string     `json:"employee_id"`
	CompensationItemID string     `json:"compensation_item_id"`
	Amount             float64    `json:"amount"`
	Frequency          string     `json:"frequency"`
	EffectiveDate      time.Time  `json:"effective_date"`
	EndDate            *time.Time `json:"end_date"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type BenefitTypeResponse struct {
	ID                        string    `json:"id"`
	Name                      string    `json:"name"`
	Description               string    `json:"description"`
	EmployerContributionType  string    `json:"employer_contribution_type"`
	EmployerContributionValue float64   `json:"employer_contribution_value"`
	EmployeeContributionType  string    `json:"employee_contribution_type"`
	EmployeeContributionValue float64   `json:"employee_contribution_value"`
	IsActive                  bool      `json:"is_active"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type EmployeeBenefitResponse struct {
	ID                string     `json:"id"`
	EmployeeID        string     `json:"employee_id"`
	BenefitTypeID     string     `json:"benefit_type_id"`
	ParticipantNumber string     `json:"participant_number"`
	EffectiveDate     time.Time  `json:"effective_date"`
	EndDate           *time.Time `json:"end_date"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type DeductionTypeResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	DeductionType string    `json:"deduction_type"`
	DefaultValue  float64   `json:"default_value"`
	IsActive      bool      `json:"is_active"`
	IsMandatory   bool      `json:"is_mandatory"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// --- Period ---

type PayrollPeriodResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- Period Overview ---

type PeriodOverviewResponse struct {
	Period    *PayrollPeriodResponse   `json:"period"`
	Employees []*PeriodOverviewEmployee `json:"employees"`
}

type PeriodOverviewEmployee struct {
	EmployeeID         string              `json:"employee_id"`
	EmployeeName       string              `json:"employee_name"`
	EmployeeNumber     string              `json:"employee_number"`
	DesignationName    *string             `json:"designation_name"`
	ProfilePhotoURL    string              `json:"profile_photo_url"`
	BaseSalary         *BaseSalaryResponse `json:"base_salary"`
	TotalCompensations float64             `json:"total_compensations"`
	TotalDeductions    float64             `json:"total_deductions"`
	AbsentDays         int                 `json:"absent_days"`
	NetSalary          float64             `json:"net_salary"`
}

// --- Employee Components ---

type EmployeeComponentsResponse struct {
	BaseSalary    *BaseSalaryResponse             `json:"base_salary"`
	Compensations []*EmployeeCompensationResponse `json:"compensations"`
	Benefits      []*EmployeeBenefitResponse      `json:"benefits"`
	Deductions    []*EmployeeDeductionResponse    `json:"deductions"`
}

// --- Pay Slip ---

type BreakdownItem struct {
	DeductionTypeID     string  `json:"deduction_type_id,omitempty"`
	CompensationItemID  string  `json:"compensation_item_id,omitempty"`
	Name                string  `json:"name"`
	Amount              float64 `json:"amount"`
}

type PaySlipResponse struct {
	ID                     string          `json:"id"`
	PeriodID               string          `json:"period_id"`
	PeriodName             string          `json:"period_name"`
	EmployeeID             string          `json:"employee_id"`
	EmployeeName           string          `json:"employee_name"`
	EmployeeNumber         string          `json:"employee_number"`
	DesignationName        string          `json:"designation_name"`
	ProfilePhotoURL        string          `json:"profile_photo_url"`
	BaseSalary             float64         `json:"base_salary"`
	TotalCompensations     float64         `json:"total_compensations"`
	TotalDeductions        float64         `json:"total_deductions"`
	AbsentDays             int             `json:"absent_days"`
	NetSalary              float64         `json:"net_salary"`
	Currency               string          `json:"currency"`
	Source                 string          `json:"source"`
	CompensationsBreakdown []BreakdownItem `json:"compensations_breakdown"`
	DeductionsBreakdown    []BreakdownItem `json:"deductions_breakdown"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

type EmployeeDeductionResponse struct {
	ID              string     `json:"id"`
	EmployeeID      string     `json:"employee_id"`
	DeductionTypeID string     `json:"deduction_type_id"`
	Value           *float64   `json:"value"`
	EffectiveDate   time.Time  `json:"effective_date"`
	EndDate         *time.Time `json:"end_date"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// --- converters ---

func baseSalaryToResponse(s *entity.EmployeeBaseSalary) *BaseSalaryResponse {
	return &BaseSalaryResponse{
		ID:            s.ID,
		EmployeeID:    s.EmployeeID,
		Amount:        s.Amount.Float(),
		Currency:      s.Currency.String(),
		EffectiveDate: s.EffectiveDate,
		EndDate:       s.EndDate,
		Notes:         s.Notes,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func baseSalariesToResponse(ss []*entity.EmployeeBaseSalary) []*BaseSalaryResponse {
	out := make([]*BaseSalaryResponse, len(ss))
	for i, s := range ss {
		out[i] = baseSalaryToResponse(s)
	}
	return out
}

func compItemToResponse(ci *entity.CompensationItem) *CompensationItemResponse {
	return &CompensationItemResponse{
		ID:          ci.ID,
		Name:        ci.Name,
		ItemType:    string(ci.ItemType),
		Description: ci.Description,
		IsActive:    ci.IsActive,
		IsTaxable:   ci.IsTaxable,
		CreatedAt:   ci.CreatedAt,
		UpdatedAt:   ci.UpdatedAt,
	}
}

func compItemsToResponse(cis []*entity.CompensationItem) []*CompensationItemResponse {
	out := make([]*CompensationItemResponse, len(cis))
	for i, ci := range cis {
		out[i] = compItemToResponse(ci)
	}
	return out
}

func empCompToResponse(ec *entity.EmployeeCompensation) *EmployeeCompensationResponse {
	return &EmployeeCompensationResponse{
		ID:                 ec.ID,
		EmployeeID:         ec.EmployeeID,
		CompensationItemID: ec.CompensationItemID,
		Amount:             ec.Amount.Float(),
		Frequency:          string(ec.Frequency),
		EffectiveDate:      ec.EffectiveDate,
		EndDate:            ec.EndDate,
		CreatedAt:          ec.CreatedAt,
		UpdatedAt:          ec.UpdatedAt,
	}
}

func empCompsToResponse(ecs []*entity.EmployeeCompensation) []*EmployeeCompensationResponse {
	out := make([]*EmployeeCompensationResponse, len(ecs))
	for i, ec := range ecs {
		out[i] = empCompToResponse(ec)
	}
	return out
}

func benefitTypeToResponse(bt *entity.BenefitType) *BenefitTypeResponse {
	return &BenefitTypeResponse{
		ID:                        bt.ID,
		Name:                      bt.Name,
		Description:               bt.Description,
		EmployerContributionType:  string(bt.EmployerContributionType),
		EmployerContributionValue: bt.EmployerContributionValue,
		EmployeeContributionType:  string(bt.EmployeeContributionType),
		EmployeeContributionValue: bt.EmployeeContributionValue,
		IsActive:                  bt.IsActive,
		CreatedAt:                 bt.CreatedAt,
		UpdatedAt:                 bt.UpdatedAt,
	}
}

func benefitTypesToResponse(bts []*entity.BenefitType) []*BenefitTypeResponse {
	out := make([]*BenefitTypeResponse, len(bts))
	for i, bt := range bts {
		out[i] = benefitTypeToResponse(bt)
	}
	return out
}

func empBenefitToResponse(eb *entity.EmployeeBenefit) *EmployeeBenefitResponse {
	return &EmployeeBenefitResponse{
		ID:                eb.ID,
		EmployeeID:        eb.EmployeeID,
		BenefitTypeID:     eb.BenefitTypeID,
		ParticipantNumber: eb.ParticipantNumber,
		EffectiveDate:     eb.EffectiveDate,
		EndDate:           eb.EndDate,
		CreatedAt:         eb.CreatedAt,
		UpdatedAt:         eb.UpdatedAt,
	}
}

func empBenefitsToResponse(ebs []*entity.EmployeeBenefit) []*EmployeeBenefitResponse {
	out := make([]*EmployeeBenefitResponse, len(ebs))
	for i, eb := range ebs {
		out[i] = empBenefitToResponse(eb)
	}
	return out
}

func deductionTypeToResponse(dt *entity.DeductionType) *DeductionTypeResponse {
	return &DeductionTypeResponse{
		ID:            dt.ID,
		Name:          dt.Name,
		Slug:          dt.Slug,
		Description:   dt.Description,
		DeductionType: string(dt.DeductionType),
		DefaultValue:  dt.DefaultValue,
		IsActive:      dt.IsActive,
		IsMandatory:   dt.IsMandatory,
		CreatedAt:     dt.CreatedAt,
		UpdatedAt:     dt.UpdatedAt,
	}
}

func deductionTypesToResponse(dts []*entity.DeductionType) []*DeductionTypeResponse {
	out := make([]*DeductionTypeResponse, len(dts))
	for i, dt := range dts {
		out[i] = deductionTypeToResponse(dt)
	}
	return out
}

func empDeductionToResponse(ed *entity.EmployeeDeduction) *EmployeeDeductionResponse {
	return &EmployeeDeductionResponse{
		ID:              ed.ID,
		EmployeeID:      ed.EmployeeID,
		DeductionTypeID: ed.DeductionTypeID,
		Value:           ed.Value,
		EffectiveDate:   ed.EffectiveDate,
		EndDate:         ed.EndDate,
		CreatedAt:       ed.CreatedAt,
		UpdatedAt:       ed.UpdatedAt,
	}
}

func empDeductionsToResponse(eds []*entity.EmployeeDeduction) []*EmployeeDeductionResponse {
	out := make([]*EmployeeDeductionResponse, len(eds))
	for i, ed := range eds {
		out[i] = empDeductionToResponse(ed)
	}
	return out
}

// --- Period converters ---

func periodToResponse(p *entity.PayrollPeriod) *PayrollPeriodResponse {
	return &PayrollPeriodResponse{
		ID:        p.ID,
		Name:      p.Name,
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func periodsToResponse(pp []*entity.PayrollPeriod) []*PayrollPeriodResponse {
	out := make([]*PayrollPeriodResponse, len(pp))
	for i, p := range pp {
		out[i] = periodToResponse(p)
	}
	return out
}

// --- Pay Slip converters ---

func paySlipToResponse(ps *entity.PaySlip) *PaySlipResponse {
	var compB []BreakdownItem
	for _, c := range ps.CompensationsBreakdown {
		compB = append(compB, BreakdownItem{
			CompensationItemID: c.CompensationItemID,
			Name:   c.Name,
			Amount: c.Amount,
		})
	}
	var dedB []BreakdownItem
	for _, d := range ps.DeductionsBreakdown {
		dedB = append(dedB, BreakdownItem{
			DeductionTypeID: d.DeductionTypeID,
			Name:   d.Name,
			Amount: d.Amount,
		})
	}
	return &PaySlipResponse{
		ID:                   ps.ID,
		PeriodID:             ps.PeriodID,
		EmployeeID:           ps.EmployeeID,
		BaseSalary:           ps.BaseSalary.Float(),
		TotalCompensations:   ps.TotalCompensations.Float(),
		TotalDeductions:      ps.TotalDeductions.Float(),
		AbsentDays:           ps.AbsentDays,
		NetSalary:            ps.NetSalary.Float(),
		Currency:             ps.Currency.String(),
		Source:               string(ps.Source),
		CompensationsBreakdown: compB,
		DeductionsBreakdown:    dedB,
		CreatedAt:            ps.CreatedAt,
		UpdatedAt:            ps.UpdatedAt,
	}
}

func paySlipsToResponse(pss []*entity.PaySlip) []*PaySlipResponse {
	out := make([]*PaySlipResponse, len(pss))
	for i, ps := range pss {
		out[i] = paySlipToResponse(ps)
	}
	return out
}

func enrichPaySlipWithPeriodName(resp *PaySlipResponse, periodMap map[string]string) {
	if name, ok := periodMap[resp.PeriodID]; ok {
		resp.PeriodName = name
	}
}

func enrichPaySlipsWithPeriodName(resps []*PaySlipResponse, periodMap map[string]string) {
	for _, r := range resps {
		enrichPaySlipWithPeriodName(r, periodMap)
	}
}

func collectPeriodIDsFromPaySlips(resps []*PaySlipResponse) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, r := range resps {
		if _, ok := seen[r.PeriodID]; !ok {
			seen[r.PeriodID] = struct{}{}
			ids = append(ids, r.PeriodID)
		}
	}
	return ids
}
