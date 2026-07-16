package delivery

import (
	"hrms/internal/employee/models"
	"time"
)

type EmployeeResponse struct {
	ID                    string      `json:"id"`
	UserID                *string     `json:"user_id"`
	FullName              string      `json:"full_name"`
	EmployeeNumber        string      `json:"employee_number"`
	Phone                 string      `json:"phone"`
	PersonalEmail         string      `json:"personal_email"`
	EmergencyContactName  string      `json:"emergency_contact_name"`
	EmergencyContactPhone string      `json:"emergency_contact_phone"`
	PlaceOfBirth          string      `json:"place_of_birth"`
	DateOfBirth           *string     `json:"date_of_birth"`
	JoinDate              *string     `json:"join_date"`
	Gender                string      `json:"gender"`
	Education             string      `json:"education"`
	Status                string      `json:"status"`
	Address               string      `json:"address"`
	DesignationID         *string     `json:"designation_id"`
	DesignationName       *string     `json:"designation_name"`
	NationalID            string      `json:"national_id"`
	Religion              string      `json:"religion"`
	BankHolder            string      `json:"bank_holder"`
	BankName              string      `json:"bank_name"`
	BankNumber            string      `json:"bank_number"`
	ProfilePhotoURL       string      `json:"profile_photo_url"`
	IsActive              bool        `json:"is_active"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
	Device                *DeviceInfo `json:"device"`
}

type DeviceInfo struct {
	ID         string    `json:"id"`
	Platform   string    `json:"platform"`
	Name       string    `json:"name"`
	IsActive   bool      `json:"is_active"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func fromEmployeeResult(r *models.EmployeeResult) *EmployeeResponse {
	if r == nil {
		return nil
	}

	resp := &EmployeeResponse{
		ID:                    r.ID,
		UserID:                r.UserID,
		FullName:              r.FullName,
		EmployeeNumber:        r.EmployeeNumber,
		Phone:                 r.Phone,
		PersonalEmail:         r.PersonalEmail,
		EmergencyContactName:  r.EmergencyContactName,
		EmergencyContactPhone: r.EmergencyContactPhone,
		PlaceOfBirth:          r.PlaceOfBirth,
		DateOfBirth:           r.DateOfBirth,
		JoinDate:              r.JoinDate,
		Gender:                r.Gender,
		Education:             r.Education,
		Status:                r.Status,
		Address:               r.Address,
		DesignationID:         r.DesignationID,
		DesignationName:       r.DesignationName,
		NationalID:            r.NationalID,
		Religion:              r.Religion,
		BankHolder:            r.BankHolder,
		BankName:              r.BankName,
		BankNumber:            r.BankNumber,
		ProfilePhotoURL:       r.ProfilePhotoURL,
		IsActive:              r.IsActive,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}

	if r.Device != nil {
		resp.Device = &DeviceInfo{
			ID:         r.Device.ID,
			Platform:   r.Device.Platform,
			Name:       r.Device.Name,
			IsActive:   r.Device.IsActive,
			LastUsedAt: r.Device.LastUsedAt,
			CreatedAt:  r.Device.CreatedAt,
		}
	}

	return resp
}
