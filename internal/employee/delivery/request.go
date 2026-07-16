package delivery

type UpdateProfilePhotoRequest struct {
	ProfilePhotoID *string `json:"profile_photo_id" validate:"omitempty,uuid"`
}

type UpdateMyProfilePhotoRequest struct {
	DocumentID string `json:"document_id" validate:"required,uuid"`
}

type UpdateContactRequest struct {
	Phone                 *string `json:"phone"`
	PersonalEmail         *string `json:"personal_email" validate:"omitempty,email"`
	EmergencyContactName  *string `json:"emergency_contact_name"`
	EmergencyContactPhone *string `json:"emergency_contact_phone"`
}

type UpdateIdentityRequest struct {
	FullName     *string `json:"full_name" validate:"omitempty,min=1,max=255"`
	PlaceOfBirth *string `json:"place_of_birth"`
	DateOfBirth  *string `json:"date_of_birth" validate:"omitempty"`
	Gender       *string `json:"gender" validate:"omitempty,oneof=female male other"`
	Education    *string `json:"education"`
	Address      *string `json:"address"`
	NationalID   *string `json:"national_id"`
	Religion     *string `json:"religion" validate:"omitempty,oneof=islam christian catholic hindu buddhist confucian other"`
}

type UpdateBankRequest struct {
	BankHolder string `json:"bank_holder" validate:"required"`
	BankName   string `json:"bank_name" validate:"required"`
	BankNumber string `json:"bank_number" validate:"required"`
}

type CreateRequest struct {
	FullName              string  `json:"full_name" validate:"required,min=1,max=255"`
	EmployeeNumber        string  `json:"employee_number" validate:"omitempty"`
	Phone                 string  `json:"phone" validate:"omitempty"`
	PersonalEmail         string  `json:"personal_email" validate:"omitempty,email"`
	EmergencyContactName  string  `json:"emergency_contact_name" validate:"omitempty"`
	EmergencyContactPhone string  `json:"emergency_contact_phone" validate:"omitempty"`
	PlaceOfBirth          string  `json:"place_of_birth" validate:"required"`
	DateOfBirth           *string `json:"date_of_birth" validate:"omitempty"`
	JoinDate              *string `json:"join_date" validate:"omitempty"`
	Gender                string  `json:"gender" validate:"required,oneof=female male other"`
	Education             string  `json:"education" validate:"required"`
	Address               string  `json:"address" validate:"omitempty"`
	DesignationID         *string `json:"designation_id" validate:"omitempty,uuid"`
	NationalID            string  `json:"national_id" validate:"required"`
	Religion              string  `json:"religion" validate:"required,oneof=islam christian catholic hindu buddhist confucian other"`
	BankHolder            string  `json:"bank_holder" validate:"omitempty"`
	BankName              string  `json:"bank_name" validate:"omitempty"`
	BankNumber            string  `json:"bank_number" validate:"omitempty"`
}
