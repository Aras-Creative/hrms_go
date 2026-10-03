package usecase

import (
	"context"
	"fmt"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/repository"
	errors "hrms/internal/pkg/apperror"
)

type EmployeePeriodOverview struct {
	EmployeeID         string
	EmployeeName       string
	EmployeeNumber     string
	DesignationName    *string
	ProfilePhotoURL    string
	BaseSalary         *entity.EmployeeBaseSalary
	TotalCompensations float64
	TotalDeductions    float64
	AbsentDays         int
	NetSalary          float64
}

type PhotoResolver interface {
	ResolveURLs(ctx context.Context, documentIDs []string) (map[string]string, error)
}

type OverviewUsecase struct {
	periodRepo    repository.PayrollPeriodRepository
	salaryRepo    repository.EmployeeBaseSalaryRepository
	ovRepo        repository.OverviewRepository
	calcRepo      repository.PayrollCalculationRepository
	photoResolver PhotoResolver
}

func NewOverviewUsecase(
	periodRepo repository.PayrollPeriodRepository,
	salaryRepo repository.EmployeeBaseSalaryRepository,
	ovRepo repository.OverviewRepository,
	calcRepo repository.PayrollCalculationRepository,
	photoResolver PhotoResolver,
) *OverviewUsecase {
	return &OverviewUsecase{
		periodRepo:    periodRepo,
		salaryRepo:    salaryRepo,
		ovRepo:        ovRepo,
		calcRepo:      calcRepo,
		photoResolver: photoResolver,
	}
}

func (uc *OverviewUsecase) GetPeriodOverview(ctx context.Context, periodID string) ([]*EmployeePeriodOverview, error) {
	period, err := uc.periodRepo.FindByID(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("find period: %w", err)
	}
	if period == nil {
		return nil, errors.NewNotFound("period not found")
	}

	employees, err := uc.ovRepo.QueryEmployees(ctx, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query employees: %w", err)
	}

	employeeIDs := make([]string, len(employees))
	for i, e := range employees {
		employeeIDs[i] = e.EmployeeID
	}

	salaryMap, err := uc.salaryRepo.FindCurrentByEmployeeIDs(ctx, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("find salaries: %w", err)
	}

	salaryCentsMap := make(map[string]int64, len(salaryMap))
	for empID, s := range salaryMap {
		if s != nil {
			salaryCentsMap[empID] = s.Amount.Cents()
		}
	}

	compRows, err := uc.ovRepo.QueryCompensationRowsBatch(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query compensations: %w", err)
	}

	dedRows, err := uc.ovRepo.QueryDeductionRowsBatch(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query deductions: %w", err)
	}

	attendanceMap, err := uc.calcRepo.QueryAttendanceDayCounts(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query attendance day counts: %w", err)
	}

	workingDaysMap, err := uc.calcRepo.QueryEmployeeWorkingDaysBatch(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query working days: %w", err)
	}

	// Totals are derived with the same row-level calculation the processor uses, so
	// the overview always agrees with the payslips that were generated from it.
	compMap := make(map[string]float64, len(employeeIDs))
	dedMap := make(map[string]float64, len(employeeIDs))
	absentDaysMap := make(map[string]int, len(employeeIDs))
	for _, empID := range employeeIDs {
		attendance := attendanceMap[empID]
		calcCtx := entity.CalcContext{
			BaseSalaryCents:  salaryCentsMap[empID],
			WorkingDays:      workingDaysMap[empID],
			UnpaidAbsentDays: attendance.UnpaidAbsent,
			AttendedDays:     attendance.Attended,
		}
		absentDaysMap[empID] = attendance.UnpaidAbsent

		var compCents int64
		for _, c := range compRows[empID] {
			compCents += c.CalculateCents(calcCtx)
		}
		compMap[empID] = float64(compCents) / 100

		var dedCents int64
		for _, d := range dedRows[empID] {
			dedCents += d.CalculateCents(calcCtx)
		}
		dedMap[empID] = float64(dedCents) / 100
	}

	result := make([]*EmployeePeriodOverview, 0, len(employees))

	var photoIDs []string
	for _, e := range employees {
		if e.ProfilePhotoID != nil {
			photoIDs = append(photoIDs, *e.ProfilePhotoID)
		}
	}
	photoURLs := make(map[string]string)
	if len(photoIDs) > 0 && uc.photoResolver != nil {
		resolved, err := uc.photoResolver.ResolveURLs(ctx, photoIDs)
		if err == nil {
			photoURLs = resolved
		}
	}

	for _, e := range employees {
		overview := &EmployeePeriodOverview{
			EmployeeID:      e.EmployeeID,
			EmployeeName:    e.EmployeeName,
			EmployeeNumber:  e.EmployeeNumber,
			DesignationName: e.DesignationName,
			BaseSalary:      salaryMap[e.EmployeeID],
			AbsentDays:      absentDaysMap[e.EmployeeID],
		}
		if e.ProfilePhotoID != nil {
			overview.ProfilePhotoURL = photoURLs[*e.ProfilePhotoID]
		}
		overview.TotalCompensations = compMap[e.EmployeeID]
		overview.TotalDeductions = dedMap[e.EmployeeID]

		var salaryFloat float64
		if salaryMap[e.EmployeeID] != nil {
			salaryFloat = salaryMap[e.EmployeeID].Amount.Float()
		}
		overview.NetSalary = salaryFloat + overview.TotalCompensations - overview.TotalDeductions

		result = append(result, overview)
	}
	return result, nil
}
