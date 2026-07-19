package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/employee/entity"
	"hrms/internal/employee/models"
	"hrms/internal/employee/numbergen"
	"hrms/internal/employee/repository"
	errors "hrms/internal/pkg/apperror"
)

type parsedEmployeeInput struct {
	phone          entity.Phone
	emergencyPhone entity.Phone
	employeeNumber entity.EmployeeNumber
	gender         entity.Gender
	religion       entity.Religion
	dateOfBirth    *entity.Date
	joinDate       *entity.Date
}

type DesignationFetcher interface {
	FindCodeByID(ctx context.Context, id string) (string, error)
}

type AccountCreator interface {
	CreateUser(ctx context.Context, username, fullName string) (userID string, err error)
}

type BalanceAssigner interface {
	AssignBalances(ctx context.Context, employeeID string) error
}

type ProfilePhotoResolver interface {
	ResolveURL(ctx context.Context, documentID string) (string, error)
	ResolveURLs(ctx context.Context, documentIDs []string) (map[string]string, error)
}

type ContractBrief struct {
	ContractType string
	StartDate    *time.Time
	EndDate      *time.Time
}

type CurrentContractFetcher interface {
	FindCurrentByEmployeeID(ctx context.Context, employeeID string) (*ContractBrief, error)
}

type EmployeeUsecase struct {
	repo            repository.EmployeeRepository
	fetcher         DesignationFetcher
	generator       *numbergen.Generator
	accountCreator  AccountCreator
	balanceAssigner BalanceAssigner
	photoResolver   ProfilePhotoResolver
	contractFetcher CurrentContractFetcher
}

func NewEmployeeUsecase(repo repository.EmployeeRepository, fetcher DesignationFetcher, generator *numbergen.Generator, accountCreator AccountCreator, balanceAssigner BalanceAssigner, photoResolver ProfilePhotoResolver, contractFetcher CurrentContractFetcher) *EmployeeUsecase {
	return &EmployeeUsecase{repo: repo, fetcher: fetcher, generator: generator, accountCreator: accountCreator, balanceAssigner: balanceAssigner, photoResolver: photoResolver, contractFetcher: contractFetcher}
}

func (uc *EmployeeUsecase) SetContractFetcher(f CurrentContractFetcher) {
	uc.contractFetcher = f
}

func (uc *EmployeeUsecase) parseEmployeeInput(ctx context.Context, input models.CreateEmployeeInput) (*parsedEmployeeInput, error) {
	var (
		phone          entity.Phone
		emergencyPhone entity.Phone
		err            error
	)
	if input.Phone != "" {
		phone, err = entity.NewPhone(input.Phone)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
	}
	if input.EmergencyContactPhone != "" {
		emergencyPhone, err = entity.NewPhone(input.EmergencyContactPhone)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
	}

	var code string
	if input.DesignationID != nil {
		code, err = uc.fetcher.FindCodeByID(ctx, *input.DesignationID)
		if err != nil {
			return nil, fmt.Errorf("resolve designation: %w", err)
		}
	}

	var employeeNumber entity.EmployeeNumber
	if input.EmployeeNumber != "" {
		employeeNumber, err = entity.NewEmployeeNumber(input.EmployeeNumber)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		if code != "" {
			if err := uc.generator.EnsureAtLeast(ctx, code, employeeNumber.Sequence()); err != nil {
				return nil, fmt.Errorf("sync sequence: %w", err)
			}
		}
	} else {
		if code == "" {
			return nil, errors.NewInvalidInput("designation_id is required when employee_number is not specified")
		}
		num, err := uc.generator.Generate(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("generate number: %w", err)
		}
		employeeNumber, err = entity.NewEmployeeNumber(num)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
	}
	gender, err := entity.ParseGender(input.Gender)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}
	religion, err := entity.ParseReligion(input.Religion)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}
	var dateOfBirth *entity.Date
	if input.DateOfBirth != nil {
		dob, err := entity.ParseDate(*input.DateOfBirth)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		dateOfBirth = &dob
	}
	var joinDate *entity.Date
	if input.JoinDate != nil {
		jd, err := entity.ParseDate(*input.JoinDate)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		joinDate = &jd
	}
	return &parsedEmployeeInput{
		phone: phone, emergencyPhone: emergencyPhone, employeeNumber: employeeNumber,
		gender: gender, religion: religion, dateOfBirth: dateOfBirth, joinDate: joinDate,
	}, nil
}

func (uc *EmployeeUsecase) Create(ctx context.Context, input models.CreateEmployeeInput) (*entity.Employee, error) {
	parsed, err := uc.parseEmployeeInput(ctx, input)
	if err != nil {
		return nil, err
	}
	userID, err := uc.accountCreator.CreateUser(ctx, parsed.employeeNumber.String(), input.FullName)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	e := entity.NewEmployee(&userID, input.FullName, parsed.employeeNumber, parsed.phone,
		input.PersonalEmail, input.EmergencyContactName, parsed.emergencyPhone,
		input.PlaceOfBirth, parsed.dateOfBirth, parsed.joinDate, parsed.gender,
		input.Education, input.Address, input.DesignationID, input.NationalID, parsed.religion)

	if input.BankHolder != "" || input.BankName != "" || input.BankNumber != "" {
		account, err := entity.NewBankAccount(input.BankHolder, input.BankName, input.BankNumber)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		e.UpdateBankInfo(account)
	}

	if err := uc.repo.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("create employee: %w", err)
	}
	if err := uc.balanceAssigner.AssignBalances(ctx, e.ID); err != nil {
		return nil, fmt.Errorf("assign balances: %w", err)
	}
	return e, nil
}

func (uc *EmployeeUsecase) Upsert(ctx context.Context, id string, input models.CreateEmployeeInput) (*entity.Employee, error) {
	parsed, err := uc.parseEmployeeInput(ctx, input)
	if err != nil {
		return nil, err
	}
	existing, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}
	if existing != nil {
		existing.EmployeeNumber = parsed.employeeNumber
		existing.JoinDate = parsed.joinDate
		existing.DesignationID = input.DesignationID
		existing.UpdatedAt = time.Now()
		existing.UpdateContactInfo(parsed.phone, input.PersonalEmail, input.EmergencyContactName, parsed.emergencyPhone)
		existing.UpdateIdentityInfo(input.FullName, input.PlaceOfBirth, parsed.dateOfBirth, parsed.gender, input.Education, input.Address, input.NationalID, parsed.religion)
		if input.BankHolder != "" || input.BankName != "" || input.BankNumber != "" {
			account, err := entity.NewBankAccount(input.BankHolder, input.BankName, input.BankNumber)
			if err != nil {
				return nil, errors.NewInvalidInput(err.Error())
			}
			existing.UpdateBankInfo(account)
		}
		if err := uc.repo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("upsert employee: %w", err)
		}
		return existing, nil
	}
	userID, err := uc.accountCreator.CreateUser(ctx, parsed.employeeNumber.String(), input.FullName)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	e := entity.ReconstituteEmployee(id, &userID, input.FullName, parsed.employeeNumber,
		parsed.phone, input.PersonalEmail, input.EmergencyContactName, parsed.emergencyPhone,
		input.PlaceOfBirth, parsed.dateOfBirth, parsed.joinDate, parsed.gender,
		input.Education, entity.StatusPendingContract, input.Address, input.DesignationID,
		input.NationalID, parsed.religion, nil, entity.BankAccount{}, false, time.Now(), time.Now())
	if input.BankHolder != "" || input.BankName != "" || input.BankNumber != "" {
		account, err := entity.NewBankAccount(input.BankHolder, input.BankName, input.BankNumber)
		if err != nil {
			return nil, errors.NewInvalidInput(err.Error())
		}
		e.UpdateBankInfo(account)
	}
	if err := uc.repo.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("create on upsert: %w", err)
	}
	if err := uc.balanceAssigner.AssignBalances(ctx, e.ID); err != nil {
		return nil, fmt.Errorf("assign balances: %w", err)
	}
	return e, nil
}

func (uc *EmployeeUsecase) PeekNextEmployeeNumber(ctx context.Context, input models.PeekNextNumberInput) (*models.PeekNextNumberResult, error) {
	code, err := uc.fetcher.FindCodeByID(ctx, input.DesignationID)
	if err != nil {
		return nil, fmt.Errorf("peek next number: %w", err)
	}
	next, err := uc.generator.Peek(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("peek next number: %w", err)
	}
	return &models.PeekNextNumberResult{EmployeeNumber: next}, nil
}

// ---- Read ----

func (uc *EmployeeUsecase) List(ctx context.Context, input models.ListEmployeeInput) (*models.ListEmployeeResult, error) {
	items, total, err := uc.repo.FindAllWithDetails(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	var photoIDs []string
	for _, item := range items {
		if item.ProfilePhotoID != nil {
			photoIDs = append(photoIDs, *item.ProfilePhotoID)
		}
	}
	if len(photoIDs) > 0 {
		urls, err := uc.photoResolver.ResolveURLs(ctx, photoIDs)
		if err == nil {
			for _, item := range items {
				if item.ProfilePhotoID != nil {
					item.ProfilePhotoURL = urls[*item.ProfilePhotoID]
				}
			}
		}
	}
	return &models.ListEmployeeResult{Items: items, Total: total}, nil
}

func (uc *EmployeeUsecase) GetByUserID(ctx context.Context, userID string) (*entity.Employee, error) {
	e, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find by user id: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("employee not found for user")
	}
	return e, nil
}

func (uc *EmployeeUsecase) GetByID(ctx context.Context, id string) (*entity.Employee, error) {
	e, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find employee: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("employee not found")
	}
	return e, nil
}

func (uc *EmployeeUsecase) GetByIDWithDetails(ctx context.Context, id string) (*models.EmployeeResult, error) {
	result, err := uc.repo.FindByIDWithDetails(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find with details: %w", err)
	}
	if result == nil {
		return nil, errors.NewNotFound("employee not found")
	}
	if result.ProfilePhotoID != nil {
		url, err := uc.photoResolver.ResolveURL(ctx, *result.ProfilePhotoID)
		if err == nil {
			result.ProfilePhotoURL = url
		}
	}
	return result, nil
}

func (uc *EmployeeUsecase) GetMe(ctx context.Context, userID string) (*models.MeResult, error) {
	result, err := uc.repo.FindByUserIDWithDetails(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find me: %w", err)
	}
	if result == nil {
		return nil, errors.NewNotFound("employee not found for user")
	}
	if result.ProfilePhotoID != nil {
		url, err := uc.photoResolver.ResolveURL(ctx, *result.ProfilePhotoID)
		if err == nil {
			result.ProfilePhotoURL = url
		}
	}
	return result, nil
}

func (uc *EmployeeUsecase) ChangeDesignation(ctx context.Context, input models.ChangeDesignationInput) ([]*entity.Employee, error) {
	if input.DesignationID != nil {
		if _, err := uc.fetcher.FindCodeByID(ctx, *input.DesignationID); err != nil {
			return nil, errors.NewInvalidInput("invalid designation_id")
		}
	}

	employees := make([]*entity.Employee, 0, len(input.EmployeeIDs))
	for _, empID := range input.EmployeeIDs {
		e, err := uc.repo.FindByID(ctx, empID)
		if err != nil {
			return nil, fmt.Errorf("find employee %s: %w", empID, err)
		}
		if e == nil {
			return nil, errors.NewNotFound("employee not found: " + empID)
		}
		e.DesignationID = input.DesignationID
		e.UpdatedAt = time.Now()
		if err := uc.repo.Update(ctx, e); err != nil {
			return nil, fmt.Errorf("update employee %s: %w", empID, err)
		}
		employees = append(employees, e)
	}
	return employees, nil
}

func (uc *EmployeeUsecase) GetProfileCompletion(ctx context.Context, userID string) (*models.ProfileCompletionResult, error) {
	e, err := uc.GetMe(ctx, userID)
	if err != nil {
		return nil, err
	}

	dobComplete := e.DateOfBirth != nil && *e.DateOfBirth != ""
	desComplete := e.DesignationID != nil && *e.DesignationID != ""

	identityComplete := e.FullName != "" && e.Gender != "" && e.PlaceOfBirth != "" &&
		dobComplete && e.NationalID != "" && e.Religion != "" && e.Education != ""
	contactComplete := e.Phone != "" && e.PersonalEmail != "" &&
		e.EmergencyContactName != "" && e.EmergencyContactPhone != ""
	employmentComplete := desComplete && e.EmployeeNumber != "" && e.JoinDate != nil && *e.JoinDate != ""
	bankComplete := e.BankHolder != "" && e.BankName != "" && e.BankNumber != ""

	return &models.ProfileCompletionResult{
		Identity: models.SectionCompletion{
			Complete: identityComplete,
			Fields: map[string]bool{
				"full_name":      e.FullName != "",
				"gender":         e.Gender != "",
				"place_of_birth": e.PlaceOfBirth != "",
				"date_of_birth":  dobComplete,
				"national_id":    e.NationalID != "",
				"religion":       e.Religion != "",
				"education":      e.Education != "",
			},
		},
		Contact: models.SectionCompletion{
			Complete: contactComplete,
			Fields: map[string]bool{
				"phone":                   e.Phone != "",
				"personal_email":          e.PersonalEmail != "",
				"emergency_contact_name":  e.EmergencyContactName != "",
				"emergency_contact_phone": e.EmergencyContactPhone != "",
			},
		},
		Employment: models.SectionCompletion{
			Complete: employmentComplete,
			Fields: map[string]bool{
				"designation":     desComplete,
				"employee_number": e.EmployeeNumber != "",
				"join_date":       e.JoinDate != nil && *e.JoinDate != "",
			},
		},
		Bank: models.SectionCompletion{
			Complete: bankComplete,
			Fields: map[string]bool{
				"bank_holder": e.BankHolder != "",
				"bank_name":   e.BankName != "",
				"bank_number": e.BankNumber != "",
			},
		},
	}, nil
}
