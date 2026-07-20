package entity

type EmployeeRenderData struct {
	Name           string
	IdentityNumber string
	BirthInfo      string
	Address        string
	Education      string
	Gender         string
	Religion       string
	Phone          string
}

type SignatoryRenderData struct {
	Name        string
	Designation string
}

type ContractRenderData struct {
	Number           string
	StartDate        string
	EndDate          string
	Salary           string
	DesignationTitle string
	ShiftStart       string
	ShiftEnd         string
}

type ContractSigningRenderData struct {
	Party           string
	SignedByName    string
	SignedByTitle   string
	SignatureBase64 string
	Place           string
	SignedAt        string
}
