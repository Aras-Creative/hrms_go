package repository

import (
	"hrms/internal/employee/models"
	"hrms/internal/pkg/ptr"
	"hrms/internal/pkg/timeutil"
)

func toEmployeeResult(r *EmployeeWithDetailsRow) *models.EmployeeResult {
	result := &models.EmployeeResult{
		ID:                    r.ID,
		UserID:                r.UserID,
		FullName:              r.FullName,
		EmployeeNumber:        r.EmployeeNumber,
		Phone:                 r.Phone,
		PersonalEmail:         ptr.Val(r.PersonalEmail),
		EmergencyContactName:  r.EmergencyContactName,
		EmergencyContactPhone: r.EmergencyContactPhone,
		PlaceOfBirth:          r.PlaceOfBirth,
		DateOfBirth:           timeutil.ReformatDate(r.DateOfBirth),
		JoinDate:              timeutil.ReformatDate(r.JoinDate),
		Gender:                r.Gender,
		Education:             r.Education,
		Status:                r.Status,
		Address:               r.Address,
		DesignationID:         r.DesignationID,
		DesignationName:       r.DesignationName,
		NationalID:            r.NationalID,
		Religion:              r.Religion,
		ProfilePhotoID:        r.ProfilePhotoID,
		IsActive:              r.IsActive,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
		BankHolder:            r.BankHolder,
		BankName:              r.BankName,
		BankNumber:            r.BankNumber,
	}

	if r.DeviceID != nil {
		result.Device = &models.DeviceInfo{
			ID:         *r.DeviceID,
			Platform:   ptr.Val(r.DevicePlatform),
			Name:       ptr.Val(r.DeviceName),
			IsActive:   r.DeviceIsActive != nil && *r.DeviceIsActive,
			LastUsedAt: derefTime(r.DeviceLastUsed),
			CreatedAt:  derefTime(r.DeviceCreatedAt),
		}
	}

	return result
}
