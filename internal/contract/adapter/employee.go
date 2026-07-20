package adapter

import (
	"context"
	"fmt"
	"time"

	contractEntity "hrms/internal/contract/entity"
	emplEntity "hrms/internal/employee/entity"
	emplRepo "hrms/internal/employee/repository"
	emplUc "hrms/internal/employee/usecase"
	contractUc "hrms/internal/contract/usecase"
)

type EmployeeFetcherAdapter struct {
	repo          emplRepo.EmployeeRepository
	photoResolver emplUc.ProfilePhotoResolver
}

func NewEmployeeFetcherAdapter(repo emplRepo.EmployeeRepository, photoResolver emplUc.ProfilePhotoResolver) *EmployeeFetcherAdapter {
	return &EmployeeFetcherAdapter{repo: repo, photoResolver: photoResolver}
}

func (a *EmployeeFetcherAdapter) FindEmployeeRenderData(ctx context.Context, id string) (*contractEntity.EmployeeRenderData, error) {
	emp, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, nil
	}
	return mapToRenderData(emp), nil
}

func (a *EmployeeFetcherAdapter) FindEmployeeIDByUserID(ctx context.Context, userID string) (string, error) {
	emp, err := a.repo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if emp == nil {
		return "", nil
	}
	return emp.ID, nil
}

func (a *EmployeeFetcherAdapter) FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error) {
	employees, err := a.repo.FindByIDs(ctx, []string{employeeID})
	if err != nil {
		return "", err
	}
	if len(employees) == 0 || employees[0] == nil || employees[0].UserID == nil {
		return "", nil
	}
	return *employees[0].UserID, nil
}

func (a *EmployeeFetcherAdapter) FindDesignationIDs(ctx context.Context, ids []string) (map[string]*string, error) {
	employees, err := a.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*string, len(employees))
	for _, e := range employees {
		result[e.ID] = e.DesignationID
	}
	return result, nil
}

func (a *EmployeeFetcherAdapter) FindBriefByIDs(ctx context.Context, ids []string) (map[string]contractUc.EmployeeBrief, error) {
	employees, err := a.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	var photoIDs []string
	for _, e := range employees {
		if e.ProfilePhotoID != nil {
			photoIDs = append(photoIDs, *e.ProfilePhotoID)
		}
	}
	photoURLs := make(map[string]string)
	if len(photoIDs) > 0 {
		resolved, err := a.photoResolver.ResolveURLs(ctx, photoIDs)
		if err == nil {
			photoURLs = resolved
		}
	}

	result := make(map[string]contractUc.EmployeeBrief, len(employees))
	for _, e := range employees {
		brief := contractUc.EmployeeBrief{Name: e.FullName}
		if e.ProfilePhotoID != nil {
			brief.ProfilePhotoURL = photoURLs[*e.ProfilePhotoID]
		}
		result[e.ID] = brief
	}
	return result, nil
}

var _ contractUc.EmployeeFetcher = (*EmployeeFetcherAdapter)(nil)

// ---- TerminateEmployeeAdapter ----

type UserDeactivator interface {
	Deactivate(ctx context.Context, userID string) error
}

type TerminateEmployeeAdapter struct {
	employeeRepo emplRepo.EmployeeRepository
	userDeactiv  UserDeactivator
}

func NewTerminateEmployeeAdapter(employeeRepo emplRepo.EmployeeRepository, userDeactiv UserDeactivator) *TerminateEmployeeAdapter {
	return &TerminateEmployeeAdapter{employeeRepo: employeeRepo, userDeactiv: userDeactiv}
}

func (a *TerminateEmployeeAdapter) FindEmployeeUserID(ctx context.Context, employeeID string) (string, error) {
	employees, err := a.employeeRepo.FindByIDs(ctx, []string{employeeID})
	if err != nil {
		return "", err
	}
	if len(employees) == 0 || employees[0] == nil || employees[0].UserID == nil {
		return "", nil
	}
	return *employees[0].UserID, nil
}

func (a *TerminateEmployeeAdapter) TerminateEmployee(ctx context.Context, employeeID string, terminationDate time.Time) error {
	emp, err := a.employeeRepo.FindByID(ctx, employeeID)
	if err != nil {
		return err
	}
	if emp == nil {
		return nil
	}
	emp.Terminate(terminationDate)
	return a.employeeRepo.Update(ctx, emp)
}

var _ contractUc.EmployeeTerminator = (*TerminateEmployeeAdapter)(nil)

// ---- helpers ----

func normalizeGender(g string) string {
	switch g {
	case "female":
		return "Perempuan"
	case "male":
		return "Laki-laki"
	case "other":
		return "Lainnya"
	}
	return g
}

func normalizeReligion(r string) string {
	switch r {
	case "islam":
		return "Islam"
	case "christian":
		return "Kristen"
	case "catholic":
		return "Katolik"
	case "hindu":
		return "Hindu"
	case "buddhist":
		return "Buddha"
	case "confucian":
		return "Konghucu"
	case "other":
		return "Lainnya"
	}
	return r
}

var idMonthNames = map[time.Month]string{
	time.January: "Januari", time.February: "Februari", time.March: "Maret",
	time.April: "April", time.May: "Mei", time.June: "Juni",
	time.July: "Juli", time.August: "Agustus", time.September: "September",
	time.October: "Oktober", time.November: "November", time.December: "Desember",
}

func formatDateID(d emplEntity.Date) string {
	return fmt.Sprintf("%d %s %d", d.Day, idMonthNames[d.Month], d.Year)
}

func mapToRenderData(emp *emplEntity.Employee) *contractEntity.EmployeeRenderData {
	birthInfo := emp.PlaceOfBirth
	if emp.DateOfBirth != nil {
		birthInfo = emp.PlaceOfBirth + ", " + formatDateID(*emp.DateOfBirth)
	}
	return &contractEntity.EmployeeRenderData{
		Name:           emp.FullName,
		IdentityNumber: emp.NationalID,
		BirthInfo:      birthInfo,
		Address:        emp.Address,
		Education:      emp.Education,
		Gender:         normalizeGender(string(emp.Gender)),
		Religion:       normalizeReligion(string(emp.Religion)),
		Phone:          emp.Phone.String(),
	}
}
