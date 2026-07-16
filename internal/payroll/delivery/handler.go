package delivery

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"

	payrollAdapter "hrms/internal/payroll/adapter"
	"hrms/internal/payroll/models"
	"hrms/internal/payroll/usecase"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
)

const dateFormat = "2006-01-02"

type Notifier interface {
	Notify(ctx context.Context, userID, ntype, title, body, resource, resourceID string) error
}

const notifTypePayroll = "payroll"

type PayrollHandler struct {
	salaryUc         *usecase.SalaryUsecase
	compUc           *usecase.CompensationUsecase
	benefitUc        *usecase.BenefitUsecase
	deductionUc      *usecase.DeductionUsecase
	periodUc         *usecase.PeriodUsecase
	procUc           *usecase.ProcessorUsecase
	overviewUc       *usecase.OverviewUsecase
	setupUc          *usecase.SetupUsecase
	manualPayslipUc  *usecase.ManualPaySlipUsecase
	renderUc         *usecase.RenderUsecase
	auditLogger      *payrollAdapter.AuditLogger
	notifUC          Notifier
	empFetcher       usecase.EmployeeFetcher
	photoResolver    overviewPhotoResolver
}

type overviewPhotoResolver interface {
	ResolveURLs(ctx context.Context, documentIDs []string) (map[string]string, error)
}

func NewPayrollHandler(
	salaryUc *usecase.SalaryUsecase,
	compUc *usecase.CompensationUsecase,
	benefitUc *usecase.BenefitUsecase,
	deductionUc *usecase.DeductionUsecase,
	periodUc *usecase.PeriodUsecase,
	procUc *usecase.ProcessorUsecase,
	overviewUc *usecase.OverviewUsecase,
	setupUc *usecase.SetupUsecase,
	manualPayslipUc *usecase.ManualPaySlipUsecase,
	renderUc *usecase.RenderUsecase,
	auditLogger *payrollAdapter.AuditLogger,
	notifUC Notifier,
	empFetcher usecase.EmployeeFetcher,
	photoResolver overviewPhotoResolver,
) *PayrollHandler {
	return &PayrollHandler{
		salaryUc:        salaryUc,
		compUc:          compUc,
		benefitUc:       benefitUc,
		deductionUc:     deductionUc,
		periodUc:        periodUc,
		procUc:          procUc,
		overviewUc:      overviewUc,
		setupUc:         setupUc,
		manualPayslipUc: manualPayslipUc,
		renderUc:        renderUc,
		auditLogger:     auditLogger,
		notifUC:         notifUC,
		empFetcher:      empFetcher,
		photoResolver:   photoResolver,
	}
}

func userIDFromCtx(c fiber.Ctx) *string {
	uid, ok := c.Locals("user_id").(string)
	if !ok || uid == "" {
		return nil
	}
	return &uid
}

// --- Compensation Item ---

func (h *PayrollHandler) CreateCompensationItem(c fiber.Ctx) error {
	var req CreateCompensationItemRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}

	ci, err := h.compUc.CreateItem(c.RequestCtx(), models.CreateCompensationItemInput{
		Name:        req.Name,
		ItemType:    req.ItemType,
		Description: req.Description,
		IsTaxable:   req.IsTaxable,
	})
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, "compensation_item", ci.ID, "",
				payrollAdapter.ActionCompCreate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name, "item_type": req.ItemType},
			)
		}
	}

	return response.Created(c, compItemToResponse(ci))
}

// --- Benefit Type ---

func (h *PayrollHandler) ListCompensationItems(c fiber.Ctx) error {
	page, perPage := parsePagination(c)
	var isActive *bool
	if active := c.Query("is_active"); active != "" {
		b := active == "true"
		isActive = &b
	}

	items, total, err := h.compUc.ListItems(c.RequestCtx(), page, perPage, isActive)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Paginate(c, compItemsToResponse(items), page, perPage, total)
}

// --- Benefit Type ---

func (h *PayrollHandler) CreateBenefitType(c fiber.Ctx) error {
	var req CreateBenefitTypeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}
	if float64(req.EmployerContributionValue) < 0 {
		return response.Error(c, errors.NewInvalidInput("employer_contribution_value must be >= 0"))
	}
	if float64(req.EmployeeContributionValue) < 0 {
		return response.Error(c, errors.NewInvalidInput("employee_contribution_value must be >= 0"))
	}

	bt, err := h.benefitUc.CreateType(c.RequestCtx(), models.CreateBenefitTypeInput{
		Name:                      req.Name,
		Description:               req.Description,
		EmployerContributionType:  req.EmployerContributionType,
		EmployerContributionValue: float64(req.EmployerContributionValue),
		EmployeeContributionType:  req.EmployeeContributionType,
		EmployeeContributionValue: float64(req.EmployeeContributionValue),
	})
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, "benefit_type", bt.ID, "",
				payrollAdapter.ActionBenefitCreate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name},
			)
		}
	}

	return response.Created(c, benefitTypeToResponse(bt))
}

func (h *PayrollHandler) ListBenefitTypes(c fiber.Ctx) error {
	page, perPage := parsePagination(c)
	var isActive *bool
	if active := c.Query("is_active"); active != "" {
		b := active == "true"
		isActive = &b
	}

	items, total, err := h.benefitUc.ListTypes(c.RequestCtx(), page, perPage, isActive)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Paginate(c, benefitTypesToResponse(items), page, perPage, total)
}

// --- Deduction Type ---

func (h *PayrollHandler) CreateDeductionType(c fiber.Ctx) error {
	var req CreateDeductionTypeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}
	if float64(req.DefaultValue) < 0 {
		return response.Error(c, errors.NewInvalidInput("default_value must be >= 0"))
	}

	dt, err := h.deductionUc.CreateType(c.RequestCtx(), models.CreateDeductionTypeInput{
		Name:          req.Name,
		Description:   req.Description,
		DeductionType: req.DeductionType,
		DefaultValue:  float64(req.DefaultValue),
		IsMandatory:   req.IsMandatory,
	})
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, "deduction_type", dt.ID, "",
				payrollAdapter.ActionDeductionCreate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name, "deduction_type": req.DeductionType},
			)
		}
	}

	return response.Created(c, deductionTypeToResponse(dt))
}

func (h *PayrollHandler) ListDeductionTypes(c fiber.Ctx) error {
	page, perPage := parsePagination(c)
	var isActive, isMandatory *bool
	if active := c.Query("is_active"); active != "" {
		b := active == "true"
		isActive = &b
	}
	if mandatory := c.Query("is_mandatory"); mandatory != "" {
		b := mandatory == "true"
		isMandatory = &b
	}

	items, total, err := h.deductionUc.ListTypes(c.RequestCtx(), page, perPage, isActive, isMandatory)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Paginate(c, deductionTypesToResponse(items), page, perPage, total)
}

// --- Setup Employee Payroll ---

func (h *PayrollHandler) SetupEmployee(c fiber.Ctx) error {
	var req SetupEmployeePayrollRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}
	if req.BaseSalary != nil && float64(req.BaseSalary.Amount) < 0 {
		return response.Error(c, errors.NewInvalidInput("base_salary.amount must be >= 0"))
	}

	input := models.SetupEmployeePayrollInput{
		EmployeeID: req.EmployeeID,
	}

	if req.BaseSalary != nil {
		effDate, err := time.Parse(dateFormat, req.BaseSalary.EffectiveDate)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid base_salary.effective_date, expected YYYY-MM-DD"))
		}
		var endDate *time.Time
		if req.BaseSalary.EndDate != nil {
			t, err := time.Parse(dateFormat, *req.BaseSalary.EndDate)
			if err != nil {
				return response.Error(c, errors.NewInvalidInput("invalid base_salary.end_date, expected YYYY-MM-DD"))
			}
			endDate = &t
		}
		currency := req.BaseSalary.Currency
		if currency == "" {
			currency = "IDR"
		}
		input.BaseSalary = &models.SetupBaseSalary{
			Amount:        float64(req.BaseSalary.Amount),
			Currency:      currency,
			EffectiveDate: effDate,
			EndDate:       endDate,
			Notes:         req.BaseSalary.Notes,
		}
	}

	for _, ci := range req.Compensations {
		effDate, err := time.Parse(dateFormat, ci.EffectiveDate)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid compensation effective_date, expected YYYY-MM-DD"))
		}
		var endDate *time.Time
		if ci.EndDate != nil {
			t, err := time.Parse(dateFormat, *ci.EndDate)
			if err != nil {
				return response.Error(c, errors.NewInvalidInput("invalid compensation end_date, expected YYYY-MM-DD"))
			}
			endDate = &t
		}
		var amount float64
		if ci.Amount != nil {
			if float64(*ci.Amount) < 0 {
				return response.Error(c, errors.NewInvalidInput("compensation amount must be >= 0"))
			}
			amount = float64(*ci.Amount)
		}
		input.Compensations = append(input.Compensations, models.SetupCompensationItem{
			CompensationItemID: ci.CompensationItemID,
			Amount:             amount,
			Frequency:          ci.Frequency,
			EffectiveDate:      effDate,
			EndDate:            endDate,
		})
	}

	for _, bi := range req.Benefits {
		effDate, err := time.Parse(dateFormat, bi.EffectiveDate)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid benefit effective_date, expected YYYY-MM-DD"))
		}
		var endDate *time.Time
		if bi.EndDate != nil {
			t, err := time.Parse(dateFormat, *bi.EndDate)
			if err != nil {
				return response.Error(c, errors.NewInvalidInput("invalid benefit end_date, expected YYYY-MM-DD"))
			}
			endDate = &t
		}
		input.Benefits = append(input.Benefits, models.SetupBenefitItem{
			BenefitTypeID:     bi.BenefitTypeID,
			ParticipantNumber: bi.ParticipantNumber,
			EffectiveDate:     effDate,
			EndDate:           endDate,
		})
	}

	for _, di := range req.Deductions {
		effDate, err := time.Parse(dateFormat, di.EffectiveDate)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid deduction effective_date, expected YYYY-MM-DD"))
		}
		var endDate *time.Time
		if di.EndDate != nil {
			t, err := time.Parse(dateFormat, *di.EndDate)
			if err != nil {
				return response.Error(c, errors.NewInvalidInput("invalid deduction end_date, expected YYYY-MM-DD"))
			}
			endDate = &t
		}
		var val *float64
		if di.Value != nil {
			if float64(*di.Value) < 0 {
				return response.Error(c, errors.NewInvalidInput("deduction value must be >= 0"))
			}
			v := float64(*di.Value)
			val = &v
		}
		input.Deductions = append(input.Deductions, models.SetupDeductionItem{
			DeductionTypeID: di.DeductionTypeID,
			Value:           val,
			EffectiveDate:   effDate,
			EndDate:         endDate,
		})
	}

	// Capture before state for audit trail
	beforeSalary, _ := h.salaryUc.ListByEmployee(c.RequestCtx(), req.EmployeeID)
	beforeComps, _ := h.compUc.ListAssignmentsByEmployee(c.RequestCtx(), req.EmployeeID)
	beforeBenefits, _ := h.benefitUc.ListAssignmentsByEmployee(c.RequestCtx(), req.EmployeeID)
	beforeDeductions, _ := h.deductionUc.ListAssignmentsByEmployee(c.RequestCtx(), req.EmployeeID)

	if err := h.setupUc.SetupEmployeePayroll(c.RequestCtx(), input); err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			payload := map[string]any{
				"employee_id":             req.EmployeeID,
				"old_salary_count":        len(beforeSalary),
				"old_compensation_count":  len(beforeComps),
				"old_benefit_count":       len(beforeBenefits),
				"old_deduction_count":     len(beforeDeductions),
				"new_compensation_count":  len(req.Compensations),
				"new_benefit_count":       len(req.Benefits),
				"new_deduction_count":     len(req.Deductions),
			}
			if req.BaseSalary != nil {
				payload["new_salary_amount"] = float64(req.BaseSalary.Amount)
			}
			if len(beforeSalary) > 0 {
				payload["old_salary_amount"] = beforeSalary[0].Amount.Float()
			}
			h.auditLogger.Log(c.RequestCtx(), *actorID, "employee_payroll_setup", req.EmployeeID, "",
				payrollAdapter.ActionSetup, c.IP(), string(c.RequestCtx().UserAgent()),
				payload,
			)
		}
	}

	return response.NoContent(c)
}

// --- helpers ---

func parsePagination(c fiber.Ctx) (int, int) {
	page := 1
	perPage := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			perPage = v
		}
	}
	return page, perPage
}

func (h *PayrollHandler) buildPeriodNameMap(ctx context.Context, periodIDs []string) map[string]string {
	m := make(map[string]string, len(periodIDs))
	for _, pid := range periodIDs {
		p, err := h.periodUc.GetPeriod(ctx, pid)
		if err == nil && p != nil {
			m[pid] = p.Name
		}
	}
	return m
}

func collectEmployeeIDsFromPaySlips(resps []*PaySlipResponse) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, r := range resps {
		if _, ok := seen[r.EmployeeID]; !ok {
			seen[r.EmployeeID] = struct{}{}
			ids = append(ids, r.EmployeeID)
		}
	}
	return ids
}

func (h *PayrollHandler) enrichPaySlipsWithEmployee(ctx context.Context, resps []*PaySlipResponse) {
	if h.empFetcher == nil || len(resps) == 0 {
		return
	}
	empIDs := collectEmployeeIDsFromPaySlips(resps)
	briefs, err := h.empFetcher.FindBriefByIDs(ctx, empIDs)
	if err != nil {
		return
	}

	var photoIDs []string
	for _, r := range resps {
		if b, ok := briefs[r.EmployeeID]; ok {
			r.EmployeeName = b.FullName
			r.EmployeeNumber = b.EmployeeNumber
			r.DesignationName = b.DesignationName
			if b.ProfilePhotoID != nil {
				photoIDs = append(photoIDs, *b.ProfilePhotoID)
			}
		}
	}

	photoURLs := make(map[string]string)
	if len(photoIDs) > 0 && h.photoResolver != nil {
		resolved, err := h.photoResolver.ResolveURLs(ctx, photoIDs)
		if err == nil {
			photoURLs = resolved
		}
	}

	for _, r := range resps {
		if b, ok := briefs[r.EmployeeID]; ok {
			if b.ProfilePhotoID != nil {
				r.ProfilePhotoURL = photoURLs[*b.ProfilePhotoID]
			}
		}
	}
}

func (h *PayrollHandler) enrichPaySlipWithEmployee(ctx context.Context, resp *PaySlipResponse) {
	h.enrichPaySlipsWithEmployee(ctx, []*PaySlipResponse{resp})
}

// --- Options ---

func (h *PayrollHandler) ListCompensationItemOptions(c fiber.Ctx) error {
	items, err := h.compUc.GetItemOptions(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}
	opts := make([]response.Option, 0, len(items))
	for _, ci := range items {
		opts = append(opts, response.Option{
			Value: ci.ID,
			Label: ci.Name,
			Extra: ci.ItemType,
		})
	}
	return response.Options(c, opts)
}

func (h *PayrollHandler) ListBenefitTypeOptions(c fiber.Ctx) error {
	items, err := h.benefitUc.GetTypeOptions(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}
	opts := make([]response.Option, 0, len(items))
	for _, bt := range items {
		opts = append(opts, response.Option{
			Value: bt.ID,
			Label: bt.Name,
		})
	}
	return response.Options(c, opts)
}

func (h *PayrollHandler) ListDeductionTypeOptions(c fiber.Ctx) error {
	items, err := h.deductionUc.GetTypeOptions(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}
	opts := make([]response.Option, 0, len(items))
	for _, dt := range items {
		opts = append(opts, response.Option{
			Value: dt.ID,
			Label: dt.Name,
		})
	}
	return response.Options(c, opts)
}

// --- Payroll Period ---

func (h *PayrollHandler) CreatePeriod(c fiber.Ctx) error {
	var req CreatePeriodRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}

	startDate, err := time.Parse(dateFormat, req.StartDate)
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid start_date, expected YYYY-MM-DD"))
	}
	endDate, err := time.Parse(dateFormat, req.EndDate)
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid end_date, expected YYYY-MM-DD"))
	}

	p, err := h.periodUc.CreatePeriod(c.RequestCtx(), models.CreatePeriodInput{
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, "payroll_period", p.ID, "",
				payrollAdapter.ActionPeriodCreate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name, "start_date": startDate.Format("2006-01-02"), "end_date": endDate.Format("2006-01-02")},
			)
		}
	}

	return response.Created(c, periodToResponse(p))
}

func (h *PayrollHandler) ListPeriods(c fiber.Ctx) error {
	page, perPage := parsePagination(c)
	periods, total, err := h.periodUc.ListPeriods(c.RequestCtx(), page, perPage)
	if err != nil {
		return response.Error(c, err)
	}
	return response.Paginate(c, periodsToResponse(periods), page, perPage, total)
}

func (h *PayrollHandler) ProcessPeriod(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.procUc.ProcessPeriod(c.RequestCtx(), id); err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, "payroll_period", id, "",
				payrollAdapter.ActionPeriodProcess, c.IP(), string(c.RequestCtx().UserAgent()),
				nil,
			)
		}
	}

	if h.notifUC != nil {
		slips, _ := h.periodUc.ListPaySlips(c.RequestCtx(), id)
		seen := make(map[string]struct{})
		for _, s := range slips {
			if _, ok := seen[s.EmployeeID]; ok {
				continue
			}
			seen[s.EmployeeID] = struct{}{}
			empUserID, _ := h.empFetcher.FindUserIDByEmployeeID(c.RequestCtx(), s.EmployeeID)
			if empUserID != "" {
				h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypePayroll,
					"Payroll Diproses",
					"Payroll periode ini telah diproses dan slip gaji Anda tersedia",
					"payroll_period", id,
				)
			}
		}
	}

	return response.NoContent(c)
}

func (h *PayrollHandler) ClosePeriod(c fiber.Ctx) error {
	id := c.Params("id")
	p, err := h.periodUc.ClosePeriod(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, "payroll_period", id, "",
				payrollAdapter.ActionPeriodClose, c.IP(), string(c.RequestCtx().UserAgent()),
				nil,
			)
		}
	}

	if h.notifUC != nil {
		slips, _ := h.periodUc.ListPaySlips(c.RequestCtx(), id)
		seen := make(map[string]struct{})
		for _, s := range slips {
			if _, ok := seen[s.EmployeeID]; ok {
				continue
			}
			seen[s.EmployeeID] = struct{}{}
			empUserID, _ := h.empFetcher.FindUserIDByEmployeeID(c.RequestCtx(), s.EmployeeID)
			if empUserID != "" {
				h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypePayroll,
					"Payroll Ditutup",
					"Payroll periode ini telah ditutup dan tidak dapat diubah",
					"payroll_period", id,
				)
			}
		}
	}

	return response.OK(c, periodToResponse(p))
}

// --- Pay Slip ---

func (h *PayrollHandler) ListPaySlips(c fiber.Ctx) error {
	slips, err := h.periodUc.ListPaySlips(c.RequestCtx(), c.Params("id"))
	if err != nil {
		return response.Error(c, err)
	}
	resps := paySlipsToResponse(slips)
	periodMap := h.buildPeriodNameMap(c.RequestCtx(), []string{c.Params("id")})
	enrichPaySlipsWithPeriodName(resps, periodMap)
	h.enrichPaySlipsWithEmployee(c.RequestCtx(), resps)
	return response.OK(c, resps)
}

func (h *PayrollHandler) GetPaySlip(c fiber.Ctx) error {
	ps, err := h.periodUc.GetPaySlip(c.RequestCtx(), c.Params("id"))
	if err != nil {
		return response.Error(c, err)
	}
	resp := paySlipToResponse(ps)
	periodMap := h.buildPeriodNameMap(c.RequestCtx(), []string{ps.PeriodID})
	enrichPaySlipWithPeriodName(resp, periodMap)
	h.enrichPaySlipWithEmployee(c.RequestCtx(), resp)
	return response.OK(c, resp)
}

func (h *PayrollHandler) PrintPaySlip(c fiber.Ctx) error {
	pdfBytes, err := h.renderUc.PrintPayslip(c.RequestCtx(), c.Params("id"))
	if err != nil {
		return response.Error(c, err)
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "inline; filename=payslip.pdf")
	return c.SendStream(bytes.NewReader(pdfBytes))
}

func (h *PayrollHandler) CreateManualPaySlip(c fiber.Ctx) error {
	periodID := c.Params("id")

	var req CreateManualPaySlipRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}
	if float64(req.BaseSalary) < 0 {
		return response.Error(c, errors.NewInvalidInput("base_salary must be >= 0"))
	}
	if float64(req.AbsentDeduction) < 0 {
		return response.Error(c, errors.NewInvalidInput("absent_deduction must be >= 0"))
	}

	input := models.ManualPaySlipInput{
		PeriodID:        periodID,
		EmployeeID:      req.EmployeeID,
		BaseSalary:      float64(req.BaseSalary),
		Currency:        req.Currency,
		AbsentDeduction: float64(req.AbsentDeduction),
		AbsentDays:      req.AbsentDays,
	}
	for _, c := range req.Compensations {
		input.Compensations = append(input.Compensations, models.ManualCompensationInput{
			CompensationItemID: c.CompensationItemID,
			Amount:             float64(c.Amount),
		})
	}
	for _, d := range req.Deductions {
		input.Deductions = append(input.Deductions, models.ManualDeductionInput{
			DeductionTypeID: d.DeductionTypeID,
			Amount:          float64(d.Amount),
		})
	}

	ps, err := h.manualPayslipUc.CreateManualPaySlip(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, "pay_slip", ps.ID, "",
				payrollAdapter.ActionPayslipCreate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"period_id": periodID, "employee_id": req.EmployeeID, "source": "manual"},
			)
		}
	}

	if h.notifUC != nil && h.empFetcher != nil {
		empUserID, _ := h.empFetcher.FindUserIDByEmployeeID(c.RequestCtx(), req.EmployeeID)
		if empUserID != "" {
			h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypePayroll,
				"Slip Gaji Baru",
				"Slip gaji manual telah dibuat untuk Anda",
				"pay_slip", ps.ID,
			)
		}
	}

	resp := paySlipToResponse(ps)
	periodMap := h.buildPeriodNameMap(c.RequestCtx(), []string{periodID})
	enrichPaySlipWithPeriodName(resp, periodMap)
	h.enrichPaySlipWithEmployee(c.RequestCtx(), resp)
	return response.Created(c, resp)
}

// --- Period Overview ---

func (h *PayrollHandler) GetPeriodOverview(c fiber.Ctx) error {
	periodID := c.Params("id")

	period, err := h.periodUc.GetPeriod(c.RequestCtx(), periodID)
	if err != nil {
		return response.Error(c, err)
	}

	overviews, err := h.overviewUc.GetPeriodOverview(c.RequestCtx(), periodID)
	if err != nil {
		return response.Error(c, err)
	}

	out := make([]*PeriodOverviewEmployee, len(overviews))
	for i, o := range overviews {
		r := &PeriodOverviewEmployee{
			EmployeeID:         o.EmployeeID,
			EmployeeName:       o.EmployeeName,
			EmployeeNumber:     o.EmployeeNumber,
			DesignationName:    o.DesignationName,
			ProfilePhotoURL:    o.ProfilePhotoURL,
			TotalCompensations: o.TotalCompensations,
			TotalDeductions:    o.TotalDeductions,
			NetSalary:          o.NetSalary,
		}
		if o.BaseSalary != nil {
			r.BaseSalary = baseSalaryToResponse(o.BaseSalary)
		}
		out[i] = r
	}

	resp := &PeriodOverviewResponse{
		Period:    periodToResponse(period),
		Employees: out,
	}
	return response.OK(c, resp)
}

// --- Employee Components ---

func (h *PayrollHandler) GetEmployeeComponents(c fiber.Ctx) error {
	employeeID := c.Params("id")

	salary, err := h.salaryUc.ListByEmployee(c.RequestCtx(), employeeID)
	if err != nil {
		return response.Error(c, err)
	}
	comps, err := h.compUc.ListAssignmentsByEmployee(c.RequestCtx(), employeeID)
	if err != nil {
		return response.Error(c, err)
	}
	benefits, err := h.benefitUc.ListAssignmentsByEmployee(c.RequestCtx(), employeeID)
	if err != nil {
		return response.Error(c, err)
	}
	deductions, err := h.deductionUc.ListAssignmentsByEmployee(c.RequestCtx(), employeeID)
	if err != nil {
		return response.Error(c, err)
	}

	resp := &EmployeeComponentsResponse{}
	if len(salary) > 0 {
		resp.BaseSalary = baseSalaryToResponse(salary[0])
	}
	resp.Compensations = empCompsToResponse(comps)
	resp.Benefits = empBenefitsToResponse(benefits)
	resp.Deductions = empDeductionsToResponse(deductions)
	return response.OK(c, resp)
}

func (h *PayrollHandler) GetMyPayslips(c fiber.Ctx) error {
	uid := userIDFromCtx(c)
	if uid == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}
	slips, err := h.periodUc.GetMyPayslips(c.RequestCtx(), *uid)
	if err != nil {
		return response.Error(c, err)
	}
	resps := paySlipsToResponse(slips)
	periodMap := h.buildPeriodNameMap(c.RequestCtx(), collectPeriodIDsFromPaySlips(resps))
	enrichPaySlipsWithPeriodName(resps, periodMap)
	h.enrichPaySlipsWithEmployee(c.RequestCtx(), resps)
	return response.OK(c, resps)
}

func (h *PayrollHandler) GetMyPaySlip(c fiber.Ctx) error {
	uid := userIDFromCtx(c)
	if uid == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}
	periodID := c.Params("period_id")
	ps, err := h.periodUc.GetMyPaySlip(c.RequestCtx(), *uid, periodID)
	if err != nil {
		return response.Error(c, err)
	}
	resp := paySlipToResponse(ps)
	periodMap := h.buildPeriodNameMap(c.RequestCtx(), []string{ps.PeriodID})
	enrichPaySlipWithPeriodName(resp, periodMap)
	h.enrichPaySlipWithEmployee(c.RequestCtx(), resp)
	return response.OK(c, resp)
}

func (h *PayrollHandler) PrintMyPaySlip(c fiber.Ctx) error {
	uid := userIDFromCtx(c)
	if uid == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}
	periodID := c.Params("period_id")
	ps, err := h.periodUc.GetMyPaySlip(c.RequestCtx(), *uid, periodID)
	if err != nil {
		return response.Error(c, err)
	}
	pdfBytes, err := h.renderUc.PrintPayslip(c.RequestCtx(), ps.ID)
	if err != nil {
		return response.Error(c, err)
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "inline; filename=payslip.pdf")
	return c.SendStream(bytes.NewReader(pdfBytes))
}
