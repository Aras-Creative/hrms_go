package usecase

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"strconv"

	"hrms/internal/payroll/models"
	"hrms/internal/payroll/repository"
	errors "hrms/internal/pkg/apperror"
	"hrms/internal/pkg/fmtutil"

	"hrms/internal/payroll/entity"
)

//go:embed templates/payslip.html
var payslipTemplate string

type PDFRenderer interface {
	Render(ctx context.Context, htmlContent []byte) ([]byte, error)
}

type PayslipEmployeeFetcher interface {
	FindByID(ctx context.Context, id string) (models.PayslipEmployeeData, error)
}

type CompanySettingsProvider interface {
	GetCompanySettings(ctx context.Context) (companyName, companyAddress, logoURL string, err error)
}

type RenderUsecase struct {
	periodRepo  repository.PayrollPeriodRepository
	paySlipRepo repository.PaySlipRepository
	empFetcher  PayslipEmployeeFetcher
	pdf         PDFRenderer
	settings    CompanySettingsProvider
}

func NewRenderUsecase(
	periodRepo repository.PayrollPeriodRepository,
	paySlipRepo repository.PaySlipRepository,
	empFetcher PayslipEmployeeFetcher,
	pdf PDFRenderer,
	settings CompanySettingsProvider,
) *RenderUsecase {
	return &RenderUsecase{
		periodRepo:  periodRepo,
		paySlipRepo: paySlipRepo,
		empFetcher:  empFetcher,
		pdf:         pdf,
		settings:    settings,
	}
}

func (uc *RenderUsecase) PrintPayslip(ctx context.Context, payslipID string) ([]byte, error) {
	ps, err := uc.paySlipRepo.FindByID(ctx, payslipID)
	if err != nil {
		return nil, fmt.Errorf("find payslip: %w", err)
	}
	if ps == nil {
		return nil, errors.NewNotFound("payslip not found")
	}

	p, err := uc.periodRepo.FindByID(ctx, ps.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("find period: %w", err)
	}

	empData, err := uc.empFetcher.FindByID(ctx, ps.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}

	companyName, companyAddress, logoURL, err := uc.settings.GetCompanySettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get company settings: %w", err)
	}

	data := uc.buildRenderData(ps, p, empData, companyName, companyAddress, logoURL)

	tmpl, err := template.New("payslip").Parse(payslipTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return uc.pdf.Render(ctx, buf.Bytes())
}

func normalizeStatus(status string) string {
	switch status {
	case "active":
		return "Karyawan Aktif"
	case "inactive":
		return "Tidak Aktif"
	case "expired_contract":
		return "Kontrak Berakhir"
	case "pending_contract":
		return "Menunggu Kontrak"
	default:
		return status
	}
}

func (uc *RenderUsecase) buildRenderData(ps *entity.PaySlip, p *entity.PayrollPeriod, emp models.PayslipEmployeeData, companyName, companyAddress, logoURL string) *models.PayslipRenderData {
	currency := ps.Currency.String()

	data := &models.PayslipRenderData{
		LogoURL:        logoURL,
		CompanyName:    companyName,
		CompanyAddress: companyAddress,
		DocNumber:      fmt.Sprintf("SG/%s/%s", p.Name, docNumberSuffix(ps.ID)),
		PeriodName:     p.Name,
		PeriodRange:    fmt.Sprintf("%s – %s", p.StartDate.Format("02 Jan 2006"), p.EndDate.Format("02 Jan 2006")),

		EmployeeName:    emp.FullName,
		EmployeeNumber:  emp.EmployeeNumber,
		DesignationName: emp.DesignationName,
		Status:          normalizeStatus(emp.Status),
		AbsentDays:      ps.AbsentDays,

		BaseSalary:      fmtutil.FormatMoneyFloat(ps.BaseSalary.Float(), currency),
		BaseSalaryLabel: baseSalaryLabel(ps.BaseSalaryLabel),
		TotalIncome:     fmtutil.FormatMoneyFloat(ps.BaseSalary.Float()+ps.TotalCompensations.Float(), currency),
		TotalDeductions: fmtutil.FormatMoneyFloat(ps.TotalDeductions.Float(), currency),
		NetSalary:       fmtutil.FormatMoneyFloat(ps.NetSalary.Float(), currency),
	}

	if emp.BankName != "" {
		data.BankInfo = emp.BankNumber + " (" + emp.BankName + ")"
	}

	for _, c := range ps.CompensationsBreakdown {
		data.Compensations = append(data.Compensations, models.BreakdownRow{
			Name:   c.Name,
			Amount: fmtutil.FormatMoneyFloat(c.Amount, currency),
		})
	}

	for _, d := range ps.DeductionsBreakdown {
		data.Deductions = append(data.Deductions, models.BreakdownRow{
			Name:   d.Name,
			Amount: fmtutil.FormatMoneyFloat(d.Amount, currency),
		})
	}

	// Formatted with strconv rather than fmtutil: these are counts and percentages, not
	// money. Routing them through FormatMoneyFloat would print "10 closings" as "Rp 10",
	// which is exactly the misreading the separate section exists to prevent.
	for _, in := range ps.IncomeInputs {
		data.IncomeNotes = append(data.IncomeNotes, models.IncomeNoteRow{
			Label: entity.IncomeInputLabel(in.Key),
			Value: formatIncomeInputValue(in),
			Notes: in.Notes,
		})
	}

	return data
}

// baseSalaryLabel falls back to the historical wording so a slip without an explicit label
// (every auto slip, and every slip written before the column existed) prints as before.
func baseSalaryLabel(label string) string {
	if label == "" {
		return "Gaji Pokok"
	}
	return label
}

// docNumberSuffix takes the short form of the slip ID for the printed document number.
// Sliced defensively: the column is a UUID today, but an eight-byte slice on a shorter
// value panics, and this runs while rendering a document an employee is waiting for.
func docNumberSuffix(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// formatIncomeInputValue renders a figure plus its unit. The value itself is printed
// unformatted so 12345.6 does not become "12,345.60" and read as money; a percent carries
// its sign so the reader does not have to know the unit.
func formatIncomeInputValue(in entity.IncomeInput) string {
	value := strconv.FormatFloat(in.Value, 'f', -1, 64)
	switch in.Unit {
	case entity.IncomeInputUnitPercent:
		return value + "%"
	case entity.IncomeInputUnitCurrency:
		return fmtutil.FormatMoneyFloat(in.Value, "IDR")
	case entity.IncomeInputUnitDays:
		return value + " hari"
	default:
		return value
	}
}
