package entity

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID                    string
	UserID                *string
	FullName              string
	EmployeeNumber        EmployeeNumber
	Phone                 Phone
	PersonalEmail         string
	EmergencyContactName  string
	EmergencyContactPhone Phone
	PlaceOfBirth          string
	DateOfBirth           *Date
	JoinDate              *Date
	Gender                Gender
	Education             string
	Status                Status
	Address               string
	DesignationID         *string
	NationalID            string
	Religion              Religion
	ProfilePhotoID        *string
	BankAccount           BankAccount
	IsActive              bool
	TerminationDate       *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// BankHolder returns the bank account holder name (convenience for serialization).
func (e *Employee) BankHolder() string { return e.BankAccount.Holder() }

// BankName returns the bank name (convenience for serialization).
func (e *Employee) BankName() string { return e.BankAccount.Name() }

// BankNumber returns the bank account number (convenience for serialization).
func (e *Employee) BankNumber() string { return e.BankAccount.Number() }

// UpdateBankInfo replaces the bank account details and touches UpdatedAt.
func (e *Employee) UpdateBankInfo(account BankAccount) {
	e.BankAccount = account
	e.UpdatedAt = time.Now()
}

// UpdateContactInfo replaces contact details and touches UpdatedAt.
func (e *Employee) UpdateContactInfo(phone Phone, personalEmail, emergencyContactName string, emergencyContactPhone Phone) {
	e.Phone = phone
	e.PersonalEmail = personalEmail
	e.EmergencyContactName = emergencyContactName
	e.EmergencyContactPhone = emergencyContactPhone
	e.UpdatedAt = time.Now()
}

// UpdateProfilePhoto sets the profile photo ID and touches UpdatedAt.
func (e *Employee) UpdateProfilePhoto(photoID *string) {
	e.ProfilePhotoID = photoID
	e.UpdatedAt = time.Now()
}

// UpdateIdentityInfo replaces identity fields and touches UpdatedAt.
func (e *Employee) UpdateIdentityInfo(fullName, placeOfBirth string, dateOfBirth *Date, gender Gender, education, address, nationalID string, religion Religion) {
	e.FullName = fullName
	e.PlaceOfBirth = placeOfBirth
	e.DateOfBirth = dateOfBirth
	e.Gender = gender
	e.Education = education
	e.Address = address
	e.NationalID = nationalID
	e.Religion = religion
	e.UpdatedAt = time.Now()
}

func NewEmployee(
	userID *string,
	fullName string,
	employeeNumber EmployeeNumber,
	phone Phone,
	personalEmail string,
	emergencyContactName string,
	emergencyContactPhone Phone,
	placeOfBirth string,
	dateOfBirth *Date,
	joinDate *Date,
	gender Gender,
	education string,
	address string,
	designationID *string,
	nationalID string,
	religion Religion,
) *Employee {
	now := time.Now()
	return &Employee{
		ID:                    uuid.New().String(),
		UserID:                userID,
		FullName:              fullName,
		EmployeeNumber:        employeeNumber,
		Phone:                 phone,
		PersonalEmail:         personalEmail,
		EmergencyContactName:  emergencyContactName,
		EmergencyContactPhone: emergencyContactPhone,
		PlaceOfBirth:          placeOfBirth,
		DateOfBirth:           dateOfBirth,
		JoinDate:              joinDate,
		Gender:                gender,
		Education:             education,
		Status:                StatusPendingContract,
		Address:               address,
		DesignationID:         designationID,
		NationalID:            nationalID,
		Religion:              religion,
		IsActive:              false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func ReconstituteEmployee(
	id string,
	userID *string,
	fullName string,
	employeeNumber EmployeeNumber,
	phone Phone,
	personalEmail string,
	emergencyContactName string,
	emergencyContactPhone Phone,
	placeOfBirth string,
	dateOfBirth *Date,
	joinDate *Date,
	gender Gender,
	education string,
	status Status,
	address string,
	designationID *string,
	nationalID string,
	religion Religion,
	profilePhotoID *string,
	bankAccount BankAccount,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) *Employee {
	return &Employee{
		ID:                    id,
		UserID:                userID,
		FullName:              fullName,
		EmployeeNumber:        employeeNumber,
		Phone:                 phone,
		PersonalEmail:         personalEmail,
		EmergencyContactName:  emergencyContactName,
		EmergencyContactPhone: emergencyContactPhone,
		PlaceOfBirth:          placeOfBirth,
		DateOfBirth:           dateOfBirth,
		JoinDate:              joinDate,
		Gender:                gender,
		Education:             education,
		Status:                status,
		Address:               address,
		DesignationID:         designationID,
		NationalID:            nationalID,
		Religion:              religion,
		ProfilePhotoID:        profilePhotoID,
		BankAccount:           bankAccount,
		IsActive:              isActive,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
	}
}

// Terminate sets the employee's status to inactive and records the termination date.
func (e *Employee) Terminate(terminationDate time.Time) {
	now := time.Now()
	e.Status = StatusInactive
	e.IsActive = false
	e.TerminationDate = &terminationDate
	e.UpdatedAt = now
}
