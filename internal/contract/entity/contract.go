package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Contract struct {
	ID               string
	TemplateID       string
	EmployeeID       string
	Number           string
	StartDate        *time.Time
	EndDate          *time.Time
	Salary           string
	DesignationID    *string
	DesignationTitle string
	Status           ContractStatus
	ContractType     string
	Data             ContractTemplateData
	Templates        ContractTemplatePartials
	SentAt           *time.Time
	DocumentID       *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (c *Contract) AddSignature(party, signedBy, signedByName, signedByTitle, place, signatureBase64 string) *ContractSigning {
	c.UpdatedAt = time.Now()
	return NewContractSigning(party, c.ID, signedBy, signedByName, signedByTitle, place, signatureBase64)
}

// MarkSent transitions the contract to Sent status (first party has signed).
func (c *Contract) MarkSent() error {
	if c.Status != ContractStatusDraft {
		return fmt.Errorf("contract must be in draft status to mark as sent")
	}
	now := time.Now()
	c.Status = ContractStatusSent
	c.UpdatedAt = now
	return nil
}

// MarkActive transitions the contract to Active status (both parties have signed).
func (c *Contract) MarkActive() error {
	if c.Status != ContractStatusDraft && c.Status != ContractStatusSent {
		return fmt.Errorf("contract must be in draft or sent status to activate")
	}
	now := time.Now()
	c.Status = ContractStatusActive
	c.SentAt = &now
	c.UpdatedAt = now
	return nil
}

// CanSign reports whether the contract is in a status that allows the given
// party to sign.  Returns nil when signing is permitted, or an error describing
// why it is not.
func (c *Contract) CanSign(party string) error {
	switch party {
	case "first":
		if c.Status != ContractStatusDraft {
			return fmt.Errorf("contract must be in draft status to sign as first party")
		}
	case "second":
		if c.Status != ContractStatusDraft && c.Status != ContractStatusSent {
			return fmt.Errorf("contract must be in draft or sent status to sign as second party")
		}
	default:
		return fmt.Errorf("unknown party: %s", party)
	}
	return nil
}

// AttachDocument records that the final signed PDF has been stored.
func (c *Contract) AttachDocument() {
	c.UpdatedAt = time.Now()
}

// EvaluateSigningState applies the current signings to determine the contract's
// next status.  It returns true when both parties have signed and the final PDF
// should be generated.
func (c *Contract) EvaluateSigningState(signings []*ContractSigning) (bool, error) {
	hasFirst, hasSecond := false, false
	for _, s := range signings {
		if s.Party == "first" {
			hasFirst = true
		}
		if s.Party == "second" {
			hasSecond = true
		}
	}

	if hasFirst && hasSecond {
		if err := c.MarkActive(); err != nil {
			return false, err
		}
		return true, nil
	}
	if hasFirst {
		if err := c.MarkSent(); err != nil {
			return false, err
		}
	}
	return false, nil
}

// UpdateDraft applies content changes to a draft contract.
// Only call this when the contract is still in draft status.
func (c *Contract) UpdateDraft(
	number string,
	startDate *time.Time,
	endDate *time.Time,
	salary string,
	designationID *string,
	designationTitle string,
	data ContractTemplateData,
	templates ContractTemplatePartials,
) error {
	if c.Status != ContractStatusDraft {
		return fmt.Errorf("contract must be in draft status to update")
	}
	c.Number = number
	c.StartDate = startDate
	c.EndDate = endDate
	c.Salary = salary
	c.DesignationID = designationID
	c.DesignationTitle = designationTitle
	c.Data = data
	c.Templates = templates
	c.UpdatedAt = time.Now()
	return nil
}

// Expire transitions the contract to Expired status (superseded by a new active contract).
func (c *Contract) Expire() error {
	if c.Status != ContractStatusActive {
		return fmt.Errorf("contract must be active to expire")
	}
	now := time.Now()
	c.Status = ContractStatusExpired
	c.UpdatedAt = now
	return nil
}

func (c *Contract) Terminate() error {
	if c.Status != ContractStatusActive {
		return fmt.Errorf("contract must be active to terminate")
	}
	now := time.Now()
	c.Status = ContractStatusTerminated
	c.UpdatedAt = now
	return nil
}

func (c *Contract) CanDelete() error {
	if c.Status == ContractStatusActive {
		return fmt.Errorf("cannot delete an active contract")
	}
	if c.Status == ContractStatusSent {
		return fmt.Errorf("cannot delete a contract that has been sent for signing")
	}
	return nil
}

func NewContract(
	templateID string,
	employeeID string,
	number string,
	startDate *time.Time,
	endDate *time.Time,
	salary string,
	designationID *string,
	designationTitle string,
	data ContractTemplateData,
	templates ContractTemplatePartials,
) *Contract {
	now := time.Now()
	return &Contract{
		ID:               uuid.New().String(),
		TemplateID:       templateID,
		EmployeeID:       employeeID,
		Number:           number,
		StartDate:        startDate,
		EndDate:          endDate,
		Salary:           salary,
		DesignationID:    designationID,
		DesignationTitle: designationTitle,
		Status:           ContractStatusDraft,
		Data:             data,
		Templates:        templates,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func ReconstituteContract(
	id string,
	templateID string,
	employeeID string,
	number string,
	startDate *time.Time,
	endDate *time.Time,
	salary string,
	designationID *string,
	designationTitle string,
	status ContractStatus,
	contractType string,
	data ContractTemplateData,
	templates ContractTemplatePartials,
	sentAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Contract {
	return &Contract{
		ID:               id,
		TemplateID:       templateID,
		EmployeeID:       employeeID,
		Number:           number,
		StartDate:        startDate,
		EndDate:          endDate,
		Salary:           salary,
		DesignationID:    designationID,
		DesignationTitle: designationTitle,
		Status:           status,
		ContractType:     contractType,
		Data:             data,
		Templates:        templates,
		SentAt:           sentAt,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}
