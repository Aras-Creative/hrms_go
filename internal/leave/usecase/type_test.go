package usecase

import (
	"context"
	"testing"

	"hrms/internal/leave/entity"
	"hrms/internal/leave/models"
)

type mockLeaveTypeRepo struct {
	types map[string]*entity.LeaveType
}

func (m *mockLeaveTypeRepo) Create(ctx context.Context, lt *entity.LeaveType) error {
	m.types[lt.ID] = lt
	return nil
}

func (m *mockLeaveTypeRepo) FindByID(ctx context.Context, id string) (*entity.LeaveType, error) {
	return m.types[id], nil
}

func (m *mockLeaveTypeRepo) FindAllActive(ctx context.Context) ([]*entity.LeaveType, error) {
	var list []*entity.LeaveType
	for _, lt := range m.types {
		if lt.IsActive {
			list = append(list, lt)
		}
	}
	return list, nil
}

func (m *mockLeaveTypeRepo) Update(ctx context.Context, lt *entity.LeaveType) error {
	m.types[lt.ID] = lt
	return nil
}

func TestLeaveTypeUsecase_IncludeInAttendanceAllowance(t *testing.T) {
	repo := &mockLeaveTypeRepo{types: make(map[string]*entity.LeaveType)}
	uc := &LeaveUsecase{leaveTypeRepo: repo}

	// 1. Create with IncludeInAttendanceAllowance = true
	created, err := uc.CreateLeaveType(context.Background(), models.CreateLeaveTypeInput{
		Name:                         "Sakit",
		DefaultDays:                  0,
		IsPaid:                       true,
		IsUnlimited:                  true,
		IncludeInAttendanceAllowance: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType failed: %v", err)
	}
	if !created.IncludeInAttendanceAllowance {
		t.Errorf("expected IncludeInAttendanceAllowance to be true")
	}

	// 2. Update to false
	updated, err := uc.UpdateLeaveType(context.Background(), created.ID, models.UpdateLeaveTypeInput{
		Name:                         "Sakit",
		DefaultDays:                  0,
		IsPaid:                       true,
		IsUnlimited:                  true,
		IncludeInAttendanceAllowance: false,
	})
	if err != nil {
		t.Fatalf("UpdateLeaveType failed: %v", err)
	}
	if updated.IncludeInAttendanceAllowance {
		t.Errorf("expected IncludeInAttendanceAllowance to be false after update")
	}
}
