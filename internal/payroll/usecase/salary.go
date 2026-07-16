package usecase

import (
	"context"

	"hrms/internal/payroll/entity"
	"hrms/internal/payroll/repository"
)

type SalaryUsecase struct {
	salaryRepo     repository.EmployeeBaseSalaryRepository
	employeeFetcher EmployeeFetcher
}

func NewSalaryUsecase(salaryRepo repository.EmployeeBaseSalaryRepository, employeeFetcher EmployeeFetcher) *SalaryUsecase {
	return &SalaryUsecase{salaryRepo: salaryRepo, employeeFetcher: employeeFetcher}
}

func (uc *SalaryUsecase) ListByEmployee(ctx context.Context, employeeID string) ([]*entity.EmployeeBaseSalary, error) {
	return uc.salaryRepo.FindByEmployeeID(ctx, employeeID)
}
