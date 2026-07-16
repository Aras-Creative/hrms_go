package repository

import "time"

type EmployeeModel struct {
	ID                    string     `db:"id"`
	UserID                *string    `db:"user_id"`
	FullName              string     `db:"full_name"`
	EmployeeNumber        string     `db:"employee_number"`
	Phone                 string     `db:"phone"`
	PersonalEmail         *string    `db:"personal_email"`
	EmergencyContactName  string     `db:"emergency_contact_name"`
	EmergencyContactPhone string     `db:"emergency_contact_phone"`
	PlaceOfBirth          string     `db:"place_of_birth"`
	DateOfBirth           *string    `db:"date_of_birth"`
	JoinDate              *string    `db:"join_date"`
	Gender                string     `db:"gender"`
	Education             string     `db:"education"`
	Status                string     `db:"status"`
	Address               string     `db:"address"`
	DesignationID         *string    `db:"designation_id"`
	NationalID            string     `db:"national_id"`
	Religion              string     `db:"religion"`
	ProfilePhotoID        *string    `db:"profile_photo_id"`
	BankHolder            string     `db:"bank_holder"`
	BankName              string     `db:"bank_name"`
	BankNumber            string     `db:"bank_number"`
	IsActive              bool       `db:"is_active"`
	TerminationDate       *time.Time `db:"termination_date"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"`
}

type EmployeeWithDetailsRow struct {
	EmployeeModel
	DesignationName *string    `db:"designation_name"`
	DeviceID        *string    `db:"device_id"`
	DevicePlatform  *string    `db:"device_platform"`
	DeviceName      *string    `db:"device_name"`
	DeviceIsActive  *bool      `db:"device_is_active"`
	DeviceLastUsed  *time.Time `db:"device_last_used"`
	DeviceCreatedAt *time.Time `db:"device_created_at"`
}

type NumberSequenceModel struct {
	DesignationCode string    `db:"designation_code"`
	Prefix          string    `db:"prefix"`
	LastSequence    int       `db:"last_sequence"`
	UpdatedAt       time.Time `db:"updated_at"`
}
