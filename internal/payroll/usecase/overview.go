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
	dedTypeRepo   repository.DeductionTypeRepository
	calcRepo      repository.PayrollCalculationRepository
	photoResolver PhotoResolver
}

func NewOverviewUsecase(
	periodRepo repository.PayrollPeriodRepository,
	salaryRepo repository.EmployeeBaseSalaryRepository,
	ovRepo repository.OverviewRepository,
	dedTypeRepo repository.DeductionTypeRepository,
	calcRepo repository.PayrollCalculationRepository,
	photoResolver PhotoResolver,
) *OverviewUsecase {
	return &OverviewUsecase{
		periodRepo:    periodRepo,
		salaryRepo:    salaryRepo,
		ovRepo:        ovRepo,
		dedTypeRepo:   dedTypeRepo,
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

	compMap, err := uc.ovRepo.QueryTotalCompensationsBatch(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query compensations: %w", err)
	}

	salaryCentsMap := make(map[string]int64, len(salaryMap))
	for empID, s := range salaryMap {
		if s != nil {
			salaryCentsMap[empID] = s.Amount.Cents()
		}
	}

	dedMap, err := uc.ovRepo.QueryTotalDeductionsBatch(ctx, employeeIDs, period.StartDate, period.EndDate, salaryCentsMap)
	if err != nil {
		return nil, fmt.Errorf("query deductions: %w", err)
	}

	absentDaysMap, err := uc.ovRepo.QueryAbsentDaysBatch(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query absent days: %w", err)
	}

	workingDaysMap, err := uc.calcRepo.QueryEmployeeWorkingDaysBatch(ctx, employeeIDs, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query working days: %w", err)
	}

	absentDT, err := uc.dedTypeRepo.FindBySlug(ctx, "absent")
	if err != nil {
		return nil, fmt.Errorf("find absent deduction type: %w", err)
	}
	if absentDT != nil && absentDT.IsActive {
		for _, empID := range employeeIDs {
			absentDays := absentDaysMap[empID]
			if absentDays <= 0 {
				continue
			}
			wd := workingDaysMap[empID]
			if wd <= 0 {
				wd = 20
			}
			salaryCents := salaryCentsMap[empID]
			absentCents := absentDT.Calculate(salaryCents, absentDays, wd)
			dedMap[empID] += float64(absentCents) / 100
		}
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
