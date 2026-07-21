package usecase

import (
	"context"
	"fmt"

	"hrms/internal/employee/entity"
	"hrms/internal/employee/models"
	errors "hrms/internal/pkg/apperror"
)

func (uc *EmployeeUsecase) UpdateMyProfilePhoto(ctx context.Context, input models.UpdateMyProfilePhotoInput) (*entity.Employee, error) {
	e, err := uc.repo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("employee not found for user")
	}
	e.UpdateProfilePhoto(&input.DocumentID)
	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update profile photo: %w", err)
	}
	return e, nil
}

func (uc *EmployeeUsecase) UpdateProfilePhoto(ctx context.Context, input models.UpdateProfilePhotoInput) (*entity.Employee, error) {
	e, err := uc.repo.FindByID(ctx, input.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("employee not found")
	}
	e.UpdateProfilePhoto(input.ProfilePhotoID)
	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update profile photo: %w", err)
	}
	return e, nil
}

func (uc *EmployeeUsecase) UpdateContact(ctx context.Context, input models.UpdateContactInput) (*entity.Employee, error) {
	e, err := uc.repo.FindByID(ctx, input.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("employee not found")
	}

	phone := e.Phone
	if input.Phone != nil {
		p, err := entity.NewPhone(*input.Phone)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		phone = p
	}
	personalEmail := e.PersonalEmail
	if input.PersonalEmail != nil {
		personalEmail = *input.PersonalEmail
	}
	emergencyContactName := e.EmergencyContactName
	if input.EmergencyContactName != nil {
		emergencyContactName = *input.EmergencyContactName
	}
	emergencyPhone := e.EmergencyContactPhone
	if input.EmergencyContactPhone != nil {
		ep, err := entity.NewPhone(*input.EmergencyContactPhone)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		emergencyPhone = ep
	}

	e.UpdateContactInfo(phone, personalEmail, emergencyContactName, emergencyPhone)
	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update contact: %w", err)
	}
	return e, nil
}

func (uc *EmployeeUsecase) UpdateBank(ctx context.Context, input models.UpdateBankInput) (*entity.Employee, error) {
	e, err := uc.repo.FindByID(ctx, input.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("employee not found")
	}
	if input.BankHolder == "" && input.BankName == "" && input.BankNumber == "" {
		e.UpdateBankInfo(entity.BankAccount{})
	} else {
		account, err := entity.NewBankAccount(input.BankHolder, input.BankName, input.BankNumber)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		e.UpdateBankInfo(account)
	}
	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update bank info: %w", err)
	}
	return e, nil
}

func (uc *EmployeeUsecase) UpdateIdentity(ctx context.Context, input models.UpdateIdentityInput) (*entity.Employee, error) {
	e, err := uc.repo.FindByID(ctx, input.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("employee not found")
	}

	fullName := e.FullName
	if input.FullName != nil {
		fullName = *input.FullName
	}
	placeOfBirth := e.PlaceOfBirth
	if input.PlaceOfBirth != nil {
		placeOfBirth = *input.PlaceOfBirth
	}
	dateOfBirth := e.DateOfBirth
	if input.DateOfBirth != nil {
		dob, err := entity.ParseDate(*input.DateOfBirth)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		dateOfBirth = &dob
	}
	gender := e.Gender
	if input.Gender != nil {
		g, err := entity.ParseGender(*input.Gender)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		gender = g
	}
	education := e.Education
	if input.Education != nil {
		education = *input.Education
	}
	address := e.Address
	if input.Address != nil {
		address = *input.Address
	}
	nationalID := e.NationalID
	if input.NationalID != nil {
		nationalID = *input.NationalID
	}
	religion := e.Religion
	if input.Religion != nil {
		r, err := entity.ParseReligion(*input.Religion)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		religion = r
	}

	e.UpdateIdentityInfo(fullName, placeOfBirth, dateOfBirth, gender, education, address, nationalID, religion)
	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update identity: %w", err)
	}
	return e, nil
}

func (uc *EmployeeUsecase) UpdateEmployeeNumber(ctx context.Context, input models.UpdateEmployeeNumberInput) (*entity.Employee, error) {
	e, err := uc.repo.FindByID(ctx, input.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("employee not found")
	}

	num, err := entity.NewEmployeeNumber(input.EmployeeNumber)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	e.UpdateEmployeeNumber(num)
	if err := uc.repo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update employee number: %w", err)
	}
	return e, nil
}
