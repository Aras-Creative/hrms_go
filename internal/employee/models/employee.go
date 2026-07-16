package models

import "time"

type MeResult struct {
	ID                    string  `json:"id"`
	UserID                *string `json:"user_id"`
	FullName              string  `json:"full_name"`
	EmployeeNumber        string  `json:"employee_number"`
	Phone                 string  `json:"phone"`
	PersonalEmail         string  `json:"personal_email"`
	EmergencyContactName  string  `json:"emergency_contact_name"`
	EmergencyContactPhone string  `json:"emergency_contact_phone"`
	PlaceOfBirth          string  `json:"place_of_birth"`
	DateOfBirth           *string `json:"date_of_birth"`
	JoinDate              *string `json:"join_date"`
	Gender                string  `json:"gender"`
	Education             string  `json:"education"`
	Status                string  `json:"status"`
	Address               string  `json:"address"`
	DesignationID         *string `json:"designation_id"`
	DesignationName       *string `json:"designation_name"`
	NationalID            string  `json:"national_id"`
	Religion              string  `json:"religion"`
	ProfilePhotoID        *string     `json:"profile_photo_id"`
	ProfilePhotoURL       string      `json:"profile_photo_url"`
	IsActive              bool        `json:"is_active"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
	Device                *DeviceInfo `json:"device"`
	BankHolder            string      `json:"bank_holder"`
	BankName              string      `json:"bank_name"`
	BankNumber            string      `json:"bank_number"`
}

type CreateEmployeeInput struct {
	FullName              string
	EmployeeNumber        string
	Phone                 string
	PersonalEmail         string
	EmergencyContactName  string
	EmergencyContactPhone string
	PlaceOfBirth          string
	DateOfBirth           *string
	JoinDate              *string
	Gender                string
	Education             string
	Address               string
	DesignationID         *string
	NationalID            string
	Religion              string
	BankHolder            string
	BankName              string
	BankNumber            string
}

type UpdateProfilePhotoInput struct {
	EmployeeID     string
	ProfilePhotoID *string
}

type UpdateMyProfilePhotoInput struct {
	UserID     string
	DocumentID string
}

type UpdateContactInput struct {
	EmployeeID            string
	Phone                 *string
	PersonalEmail         *string
	EmergencyContactName  *string
	EmergencyContactPhone *string
}

type UpdateIdentityInput struct {
	EmployeeID   string
	FullName     *string
	PlaceOfBirth *string
	DateOfBirth  *string
	Gender       *string
	Education    *string
	Address      *string
	NationalID   *string
	Religion     *string
}

type UpdateBankInput struct {
	EmployeeID string
	BankHolder string
	BankName   string
	BankNumber string
}

type ListEmployeeInput struct {
	Page          int
	PerPage       int
	SearchName    string
	Status        string
	Gender        string
	DesignationID string
}

type ListEmployeeResult struct {
	Items []*EmployeeListItem
	Total int64
}

type EmployeeListItem struct {
	ID                    string  `json:"id"`
	FullName              string  `json:"full_name"`
	EmployeeNumber        string  `json:"employee_number"`
	Phone                 string  `json:"phone"`
	PersonalEmail         string  `json:"personal_email"`
	EmergencyContactName  string  `json:"emergency_contact_name"`
	EmergencyContactPhone string  `json:"emergency_contact_phone"`
	JoinDate              *string `json:"join_date"`
	Gender                string  `json:"gender"`
	Status                string  `json:"status"`
	DesignationID         *string `json:"designation_id"`
	DesignationName       *string `json:"designation_name"`
	ProfilePhotoID        *string `json:"profile_photo_id"`
	ProfilePhotoURL       string `json:"profile_photo_url"`
	IsActive              bool   `json:"is_active"`
}

type ProfileCompletionResult struct {
	Identity   SectionCompletion `json:"identity"`
	Contact    SectionCompletion `json:"contact"`
	Employment SectionCompletion `json:"employment"`
	Bank       SectionCompletion `json:"bank"`
}

type SectionCompletion struct {
	Complete bool            `json:"complete"`
	Fields   map[string]bool `json:"fields"`
}

type PeekNextNumberInput struct {
	DesignationID string
}

type PeekNextNumberResult struct {
	EmployeeNumber string `json:"employee_number"`
}

type EmployeeResult struct {
	ID                    string
	UserID                *string
	FullName              string
	EmployeeNumber        string
	Phone                 string
	PersonalEmail         string
	EmergencyContactName  string
	EmergencyContactPhone string
	PlaceOfBirth          string
	DateOfBirth           *string
	JoinDate              *string
	Gender                string
	Education             string
	Status                string
	Address               string
	DesignationID         *string
	DesignationName       *string
	NationalID            string
	Religion              string
	ProfilePhotoID        *string
	ProfilePhotoURL       string
	IsActive              bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	Device                *DeviceInfo
	BankHolder            string
	BankName              string
	BankNumber            string
}

type DeviceInfo struct {
	ID         string
	Platform   string
	Name       string
	IsActive   bool
	LastUsedAt time.Time
	CreatedAt  time.Time
}
