package adapter

import (
	"context"
	"time"

	emplEntity "hrms/internal/employee/entity"
	emplRepo "hrms/internal/employee/repository"
	contractUc "hrms/internal/contract/usecase"
)

type EmployeeActivatorAdapter struct {
	employeeRepo emplRepo.EmployeeRepository
}

func NewEmployeeActivatorAdapter(employeeRepo emplRepo.EmployeeRepository) *EmployeeActivatorAdapter {
	return &EmployeeActivatorAdapter{employeeRepo: employeeRepo}
}

func (a *EmployeeActivatorAdapter) ActivateEmployee(ctx context.Context, employeeID string) error {
	emp, err := a.employeeRepo.FindByID(ctx, employeeID)
	if err != nil {
		return err
	}
	if emp == nil {
		return nil
	}
	emp.Status = emplEntity.StatusActive
	emp.IsActive = true
	emp.UpdatedAt = time.Now()
	return a.employeeRepo.Update(ctx, emp)
}

var _ contractUc.EmployeeActivator = (*EmployeeActivatorAdapter)(nil)
