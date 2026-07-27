package adapter

import (
	"context"
	"time"

	emplRepo "hrms/internal/employee/repository"
	contractUc "hrms/internal/contract/usecase"
)

type DesignationAssignerAdapter struct {
	employeeRepo emplRepo.EmployeeRepository
}

func NewDesignationAssignerAdapter(employeeRepo emplRepo.EmployeeRepository) *DesignationAssignerAdapter {
	return &DesignationAssignerAdapter{employeeRepo: employeeRepo}
}

func (a *DesignationAssignerAdapter) AssignDesignation(ctx context.Context, employeeID string, designationID *string) error {
	emp, err := a.employeeRepo.FindByID(ctx, employeeID)
	if err != nil {
		return err
	}
	if emp == nil {
		return nil
	}
	emp.DesignationID = designationID
	emp.UpdatedAt = time.Now()
	return a.employeeRepo.Update(ctx, emp)
}

var _ contractUc.DesignationAssigner = (*DesignationAssignerAdapter)(nil)
