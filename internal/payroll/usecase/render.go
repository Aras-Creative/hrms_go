package usecase

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"

	"hrms/internal/payroll/models"
	"hrms/internal/payroll/repository"
	"hrms/internal/pkg/fmtutil"
	errors "hrms/internal/pkg/apperror"

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

type RenderUsecase struct {
	periodRepo  repository.PayrollPeriodRepository
	paySlipRepo repository.PaySlipRepository
	empFetcher  PayslipEmployeeFetcher
	pdf         PDFRenderer
}

func NewRenderUsecase(
	periodRepo repository.PayrollPeriodRepository,
	paySlipRepo repository.PaySlipRepository,
	empFetcher PayslipEmployeeFetcher,
	pdf PDFRenderer,
) *RenderUsecase {
	return &RenderUsecase{
		periodRepo:  periodRepo,
		paySlipRepo: paySlipRepo,
		empFetcher:  empFetcher,
		pdf:         pdf,
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

	data := uc.buildRenderData(ps, p, empData)

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

func (uc *RenderUsecase) buildRenderData(ps *entity.PaySlip, p *entity.PayrollPeriod, emp models.PayslipEmployeeData) *models.PayslipRenderData {
	currency := ps.Currency.String()

	data := &models.PayslipRenderData{
		LogoURL:         "",
		CompanyName:    "PT Nama Perusahaan Anda",
		CompanyAddress: "Jl. Contoh Alamat No. 123, Jakarta Selatan, Indonesia 12345",
		DocNumber:      fmt.Sprintf("SG/%s/%s", p.Name, ps.ID[:8]),
		PeriodName:     p.Name,
		PeriodRange:    fmt.Sprintf("%s – %s", p.StartDate.Format("02 Jan 2006"), p.EndDate.Format("02 Jan 2006")),

		EmployeeName:    emp.FullName,
		EmployeeNumber:  emp.EmployeeNumber,
		DesignationName: emp.DesignationName,
		Status:          normalizeStatus(emp.Status),
		AbsentDays:      ps.AbsentDays,

		BaseSalary:      fmtutil.FormatMoneyFloat(ps.BaseSalary.Float(), currency),
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

	return data
}
