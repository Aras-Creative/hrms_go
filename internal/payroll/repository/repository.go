package repository

import (
	"context"
	"time"

	"hrms/internal/payroll/entity"
)

type EmployeeBaseSalaryRepository interface {
	Create(ctx context.Context, s *entity.EmployeeBaseSalary) error
	FindByID(ctx context.Context, id string) (*entity.EmployeeBaseSalary, error)
	FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.EmployeeBaseSalary, error)
	FindCurrentByEmployeeID(ctx context.Context, employeeID string) (*entity.EmployeeBaseSalary, error)
	FindCurrentByEmployeeIDs(ctx context.Context, employeeIDs []string) (map[string]*entity.EmployeeBaseSalary, error)
	Update(ctx context.Context, s *entity.EmployeeBaseSalary) error
	Delete(ctx context.Context, id string) error
}

type CompensationItemRepository interface {
	Create(ctx context.Context, ci *entity.CompensationItem) error
	FindByID(ctx context.Context, id string) (*entity.CompensationItem, error)
	FindByCode(ctx context.Context, code string) (*entity.CompensationItem, error)
	FindAll(ctx context.Context, filter CompItemFilter) ([]*entity.CompensationItem, int64, error)
	CountByItemID(ctx context.Context, itemID string) (int64, error)
	Update(ctx context.Context, ci *entity.CompensationItem) error
	Delete(ctx context.Context, id string) error
}

type EmployeeCompensationRepository interface {
	Create(ctx context.Context, ec *entity.EmployeeCompensation) error
	FindByID(ctx context.Context, id string) (*entity.EmployeeCompensation, error)
	FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.EmployeeCompensation, error)
	FindAll(ctx context.Context, filter EmpCompFilter) ([]*entity.EmployeeCompensation, int64, error)
	Update(ctx context.Context, ec *entity.EmployeeCompensation) error
	Delete(ctx context.Context, id string) error
}

type BenefitTypeRepository interface {
	Create(ctx context.Context, bt *entity.BenefitType) error
	FindByID(ctx context.Context, id string) (*entity.BenefitType, error)
	FindByCode(ctx context.Context, code string) (*entity.BenefitType, error)
	FindAll(ctx context.Context, filter BenefitTypeFilter) ([]*entity.BenefitType, int64, error)
	CountByTypeID(ctx context.Context, typeID string) (int64, error)
	Update(ctx context.Context, bt *entity.BenefitType) error
	Delete(ctx context.Context, id string) error
}

type EmployeeBenefitRepository interface {
	Create(ctx context.Context, eb *entity.EmployeeBenefit) error
	FindByID(ctx context.Context, id string) (*entity.EmployeeBenefit, error)
	FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.EmployeeBenefit, error)
	FindAll(ctx context.Context, filter EmpBenefitFilter) ([]*entity.EmployeeBenefit, int64, error)
	Update(ctx context.Context, eb *entity.EmployeeBenefit) error
	Delete(ctx context.Context, id string) error
}

type DeductionTypeRepository interface {
	Create(ctx context.Context, dt *entity.DeductionType) error
	FindByID(ctx context.Context, id string) (*entity.DeductionType, error)
	FindByCode(ctx context.Context, code string) (*entity.DeductionType, error)
	FindBySlug(ctx context.Context, slug string) (*entity.DeductionType, error)
	FindAll(ctx context.Context, filter DeductionTypeFilter) ([]*entity.DeductionType, int64, error)
	CountByTypeID(ctx context.Context, typeID string) (int64, error)
	Update(ctx context.Context, dt *entity.DeductionType) error
	Delete(ctx context.Context, id string) error
}

type EmployeeDeductionRepository interface {
	Create(ctx context.Context, ed *entity.EmployeeDeduction) error
	FindByID(ctx context.Context, id string) (*entity.EmployeeDeduction, error)
	FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.EmployeeDeduction, error)
	FindAll(ctx context.Context, filter EmpDeductionFilter) ([]*entity.EmployeeDeduction, int64, error)
	Update(ctx context.Context, ed *entity.EmployeeDeduction) error
	Delete(ctx context.Context, id string) error
}

type PayrollPeriodRepository interface {
	Create(ctx context.Context, p *entity.PayrollPeriod) error
	FindByID(ctx context.Context, id string) (*entity.PayrollPeriod, error)
	FindByOverlap(ctx context.Context, startDate, endDate time.Time, excludeID string) (*entity.PayrollPeriod, error)
	FindAll(ctx context.Context, page, perPage int) ([]*entity.PayrollPeriod, int64, error)
	Update(ctx context.Context, p *entity.PayrollPeriod) error
	Delete(ctx context.Context, id string) error
}

type PaySlipRepository interface {
	Upsert(ctx context.Context, ps *entity.PaySlip) error
	FindByPeriodID(ctx context.Context, periodID string) ([]*entity.PaySlip, error)
	FindByID(ctx context.Context, id string) (*entity.PaySlip, error)
	FindByEmployeeID(ctx context.Context, employeeID string) ([]*entity.PaySlip, error)
	FindByEmployeeAndPeriod(ctx context.Context, employeeID, periodID string) (*entity.PaySlip, error)
	DeleteByPeriodID(ctx context.Context, periodID string) error
}

type PayrollCalculationRepository interface {
	QueryActiveSalaries(ctx context.Context, startDate, endDate time.Time) ([]CalcSalaryRow, error)
	QueryActiveSalariesByIDs(ctx context.Context, startDate, endDate time.Time, employeeIDs []string) ([]CalcSalaryRow, error)
	QueryAbsentDays(ctx context.Context, employeeID string, startDate, endDate time.Time) (int, error)
	QueryEmployeeCompensations(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]CalcCompRow, error)
	QueryEmployeeDeductions(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]CalcDedRow, error)
	QueryEmployeeWorkingDaysBatch(ctx context.Context, employeeIDs []string, startDate, endDate time.Time) (map[string]int, error)
}

type OverviewRepository interface {
	QueryEmployees(ctx context.Context, startDate, endDate interface{}) ([]OverviewEmployee, error)
	QueryTotalCompensationsBatch(ctx context.Context, employeeIDs []string, startDate, endDate interface{}) (map[string]float64, error)
	QueryTotalDeductionsBatch(ctx context.Context, employeeIDs []string, startDate, endDate interface{}, salaryCentsMap map[string]int64) (map[string]float64, error)
	QueryAbsentDaysBatch(ctx context.Context, employeeIDs []string, startDate, endDate interface{}) (map[string]int, error)
}

type CompItemFilter struct {
	IsActive *bool
	Page     int
	PerPage  int
}

type EmpCompFilter struct {
	EmployeeID string
	Page       int
	PerPage    int
}

type BenefitTypeFilter struct {
	IsActive *bool
	Page     int
	PerPage  int
}

type EmpBenefitFilter struct {
	EmployeeID string
	Page       int
	PerPage    int
}

type DeductionTypeFilter struct {
	IsActive   *bool
	IsMandatory *bool
	Page       int
	PerPage    int
}

type EmpDeductionFilter struct {
	EmployeeID string
	Page       int
	PerPage    int
}
