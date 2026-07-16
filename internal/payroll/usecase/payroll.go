package usecase

import "context"

type EmployeeFetcher interface {
	ExistsByID(ctx context.Context, id string) (bool, error)
	FindByUserID(ctx context.Context, userID string) (string, error)
	FindBriefByIDs(ctx context.Context, ids []string) (map[string]EmployeeBrief, error)
	FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error)
}

type EmployeeBrief struct {
	FullName       string
	EmployeeNumber string
	DesignationName string
	ProfilePhotoID *string
}
