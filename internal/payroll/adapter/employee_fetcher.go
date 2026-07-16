package adapter

import (
	"context"

	emplRepo "hrms/internal/employee/repository"
	designationRepo "hrms/internal/designation/repository"
	payrollModels "hrms/internal/payroll/models"
	payrollUc "hrms/internal/payroll/usecase"
)

type EmployeeFetcherAdapter struct {
	repo    emplRepo.EmployeeRepository
	desgRepo designationRepo.DesignationRepository
}

func NewEmployeeFetcherAdapter(repo emplRepo.EmployeeRepository, desgRepo designationRepo.DesignationRepository) *EmployeeFetcherAdapter {
	return &EmployeeFetcherAdapter{repo: repo, desgRepo: desgRepo}
}

func (a *EmployeeFetcherAdapter) ExistsByID(ctx context.Context, id string) (bool, error) {
	e, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	return e != nil, nil
}

func (a *EmployeeFetcherAdapter) FindByUserID(ctx context.Context, userID string) (string, error) {
	e, err := a.repo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if e == nil {
		return "", nil
	}
	return e.ID, nil
}

func (a *EmployeeFetcherAdapter) FindBriefByIDs(ctx context.Context, ids []string) (map[string]payrollUc.EmployeeBrief, error) {
	if len(ids) == 0 {
		return make(map[string]payrollUc.EmployeeBrief), nil
	}
	employees, err := a.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	desgIDs := make(map[string]string)
	for _, e := range employees {
		if e.DesignationID != nil {
			desgIDs[e.ID] = *e.DesignationID
		}
	}

	desgNames := make(map[string]string)
	if len(desgIDs) > 0 {
		uniqueDesgIDs := make([]string, 0, len(desgIDs))
		seen := make(map[string]struct{})
		for _, did := range desgIDs {
			if _, ok := seen[did]; !ok {
				seen[did] = struct{}{}
				uniqueDesgIDs = append(uniqueDesgIDs, did)
			}
		}
		for _, did := range uniqueDesgIDs {
			d, err := a.desgRepo.FindByID(ctx, did)
			if err == nil && d != nil {
				desgNames[did] = d.Name
			}
		}
	}

	result := make(map[string]payrollUc.EmployeeBrief, len(employees))
	for _, e := range employees {
		brief := payrollUc.EmployeeBrief{
			FullName:       e.FullName,
			EmployeeNumber: e.EmployeeNumber.String(),
			ProfilePhotoID: e.ProfilePhotoID,
		}
		if e.DesignationID != nil {
			brief.DesignationName = desgNames[*e.DesignationID]
		}
		result[e.ID] = brief
	}
	return result, nil
}

func (a *EmployeeFetcherAdapter) FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error) {
	e, err := a.repo.FindByID(ctx, employeeID)
	if err != nil {
		return "", err
	}
	if e == nil || e.UserID == nil {
		return "", nil
	}
	return *e.UserID, nil
}

// --- Payslip render fetcher ---

type PayslipEmployeeFetcherAdapter struct {
	empRepo  emplRepo.EmployeeRepository
	desgRepo designationRepo.DesignationRepository
}

func NewPayslipEmployeeFetcherAdapter(empRepo emplRepo.EmployeeRepository, desgRepo designationRepo.DesignationRepository) *PayslipEmployeeFetcherAdapter {
	return &PayslipEmployeeFetcherAdapter{empRepo: empRepo, desgRepo: desgRepo}
}

func (a *PayslipEmployeeFetcherAdapter) FindByID(ctx context.Context, id string) (payrollModels.PayslipEmployeeData, error) {
	e, err := a.empRepo.FindByID(ctx, id)
	if err != nil {
		return payrollModels.PayslipEmployeeData{}, err
	}
	if e == nil {
		return payrollModels.PayslipEmployeeData{}, nil
	}

	data := payrollModels.PayslipEmployeeData{
		FullName:       e.FullName,
		EmployeeNumber: e.EmployeeNumber.String(),
		Status:         string(e.Status),
		BankName:       e.BankName(),
		BankNumber:     e.BankNumber(),
	}

	if e.DesignationID != nil {
		desg, err := a.desgRepo.FindByID(ctx, *e.DesignationID)
		if err == nil && desg != nil {
			data.DesignationName = desg.Name
		}
	}

	return data, nil
}

var _ payrollUc.EmployeeFetcher = (*EmployeeFetcherAdapter)(nil)
var _ payrollUc.PayslipEmployeeFetcher = (*PayslipEmployeeFetcherAdapter)(nil)
