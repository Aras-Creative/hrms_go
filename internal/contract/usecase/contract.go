package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/contract/entity"
	"hrms/internal/contract/models"
	errors "hrms/internal/pkg/apperror"
)

var romanMonths = map[time.Month]string{
	time.January: "I", time.February: "II", time.March: "III",
	time.April: "IV", time.May: "V", time.June: "VI",
	time.July: "VII", time.August: "VIII", time.September: "IX",
	time.October: "X", time.November: "XI", time.December: "XII",
}

func (uc *ContractUsecase) CreateContract(ctx context.Context, input models.CreateContractInput) (*models.BulkCreateContractResult, error) {
	tmpl, err := uc.tmplRepo.FindByID(ctx, input.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("find template: %w", err)
	}
	if tmpl == nil {
		return nil, errors.NewNotFound("contract template not found")
	}
	if !tmpl.IsActive {
		return nil, errors.NewInvalidInput("contract template is not active")
	}

	employeeIDs := input.EmployeeIDs

	empDesIDs, err := uc.empFetcher.FindDesignationIDs(ctx, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("find employee designations: %w", err)
	}

	desIDSet := make(map[string]struct{})
	for _, desID := range empDesIDs {
		if desID != nil {
			desIDSet[*desID] = struct{}{}
		}
	}
	desIDs := make([]string, 0, len(desIDSet))
	for id := range desIDSet {
		desIDs = append(desIDs, id)
	}

	desNames := make(map[string]string)
	if len(desIDs) > 0 {
		names, err := uc.desFetcher.FindNamesByIDs(ctx, desIDs)
		if err != nil {
			return nil, fmt.Errorf("find designation names: %w", err)
		}
		desNames = names
	}

	data := tmpl.Data
	templates := tmpl.Templates

	startDate := &input.StartDate

	var endDate *time.Time
	if input.EndDate != nil {
		endDate = input.EndDate
	} else if tmpl.ContractType == entity.ContractTypePKWT {
		return nil, errors.NewInvalidInput("end_date is required for PKWT contracts")
	}

	salaries, err := uc.salFetcher.FindCurrentByEmployeeIDs(ctx, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch salaries: %w", err)
	}

	contracts := make([]*entity.Contract, 0, len(employeeIDs))

	for _, empID := range employeeIDs {
		designationTitle := ""
		if desID := empDesIDs[empID]; desID != nil {
			designationTitle = desNames[*desID]
		}

		salary, ok := salaries[empID]
		if !ok || salary == "" {
			return nil, errors.NewInvalidInput(fmt.Sprintf("no base salary found for employee %s", empID))
		}

		seq, err := uc.numGen.NextSequence(ctx, "CTR")
		if err != nil {
			return nil, fmt.Errorf("generate contract number: %w", err)
		}
		number := fmt.Sprintf("%03d/HRD-ARAS/%s/%s/%d",
			seq, tmpl.ContractType, romanMonths[startDate.Month()], startDate.Year())

		e := entity.NewContract(
			input.TemplateID,
			empID,
			number,
			startDate,
			endDate,
			salary,
			empDesIDs[empID],
			designationTitle,
			data,
			templates,
		)

		contracts = append(contracts, e)
	}

	if err := uc.contractRepo.BulkCreateContracts(ctx, contracts); err != nil {
		return nil, fmt.Errorf("bulk create contracts: %w", err)
	}

	return &models.BulkCreateContractResult{Contracts: contracts}, nil
}

func (uc *ContractUsecase) CheckActiveContracts(ctx context.Context, input models.CheckActiveContractInput) (*models.CheckActiveContractResult, error) {
	activeMap, err := uc.contractRepo.FindActiveContractEmployeeIDs(ctx, input.EmployeeIDs)
	if err != nil {
		return nil, fmt.Errorf("check active contracts: %w", err)
	}

	items := make([]models.CheckActiveContractItem, 0, len(activeMap))
	for _, empID := range input.EmployeeIDs {
		if endDate, ok := activeMap[empID]; ok {
			items = append(items, models.CheckActiveContractItem{
				EmployeeID: empID,
				EndDate:    endDate,
			})
		}
	}

	return &models.CheckActiveContractResult{Items: items}, nil
}

type ListContractDetailResult struct {
	Items              []*entity.Contract
	Total              int64
	SigningsByContract map[string][]*entity.ContractSigning
	EmployeeBriefs     map[string]EmployeeBrief
}

func (uc *ContractUsecase) ListContracts(ctx context.Context, input models.ListContractInput) (*models.ListContractResult, error) {
	items, total, err := uc.contractRepo.FindAllContracts(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	return &models.ListContractResult{Items: items, Total: total}, nil
}

func (uc *ContractUsecase) ListContractsWithDetail(ctx context.Context, input models.ListContractInput) (*ListContractDetailResult, error) {
	result, err := uc.ListContracts(ctx, input)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(result.Items))
	empIDs := make([]string, len(result.Items))
	for i, e := range result.Items {
		ids[i] = e.ID
		empIDs[i] = e.EmployeeID
	}

	signings, err := uc.signingRepo.FindSigningsByContractIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("find signings: %w", err)
	}
	briefs, err := uc.empFetcher.FindBriefByIDs(ctx, empIDs)
	if err != nil {
		briefs = nil
	}

	return &ListContractDetailResult{
		Items:              result.Items,
		Total:              result.Total,
		SigningsByContract: signings,
		EmployeeBriefs:     briefs,
	}, nil
}

func (uc *ContractUsecase) ListMyContractsWithDetail(ctx context.Context, input models.ListContractInput, userID string) (*ListContractDetailResult, error) {
	employeeID, err := uc.empFetcher.FindEmployeeIDByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("employee not found for authenticated user")
	}
	input.EmployeeID = employeeID
	return uc.ListContractsWithDetail(ctx, input)
}

func (uc *ContractUsecase) FindSigningsByContractIDs(ctx context.Context, contractIDs []string) (map[string][]*entity.ContractSigning, error) {
	return uc.signingRepo.FindSigningsByContractIDs(ctx, contractIDs)
}

func (uc *ContractUsecase) ListMyContracts(ctx context.Context, input models.ListContractInput, userID string) (*models.ListContractResult, error) {
	employeeID, err := uc.empFetcher.FindEmployeeIDByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("employee not found for authenticated user")
	}

	input.EmployeeID = employeeID
	return uc.ListContracts(ctx, input)
}

func (uc *ContractUsecase) CountSoonExpired(ctx context.Context, withinDays int) (int64, error) {
	return uc.contractRepo.CountSoonExpired(ctx, withinDays)
}

func (uc *ContractUsecase) CountPendingContracts(ctx context.Context, userID string) (int64, error) {
	employeeID, err := uc.empFetcher.FindEmployeeIDByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("resolve user: %w", err)
	}
	if employeeID == "" {
		return 0, nil
	}
	return uc.contractRepo.CountByEmployeeIDAndStatus(ctx, employeeID, "sent")
}

func (uc *ContractUsecase) GetMyActiveContract(ctx context.Context, userID string) (*entity.Contract, error) {
	employeeID, err := uc.empFetcher.FindEmployeeIDByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}
	if employeeID == "" {
		return nil, nil
	}

	items, _, err := uc.contractRepo.FindAllContracts(ctx, models.ListContractInput{
		EmployeeID: employeeID,
		Status:     "active",
		PerPage:    1,
	})
	if err != nil {
		return nil, fmt.Errorf("find active contract: %w", err)
	}
	if len(items) == 0 {
		return nil, nil
	}

	return items[0], nil
}

func (uc *ContractUsecase) GetEmployeeContract(ctx context.Context, employeeID string) (*entity.Contract, error) {
	return uc.contractRepo.FindCurrentByEmployeeID(ctx, employeeID)
}

func (uc *ContractUsecase) FindUserIDByEmployeeID(ctx context.Context, employeeID string) (string, error) {
	return uc.empFetcher.FindUserIDByEmployeeID(ctx, employeeID)
}

func (uc *ContractUsecase) FindEmployeeBriefs(ctx context.Context, employeeIDs []string) (map[string]EmployeeBrief, error) {
	return uc.empFetcher.FindBriefByIDs(ctx, employeeIDs)
}

func (uc *ContractUsecase) GetContractDetail(ctx context.Context, contractID string) (*entity.Contract, string, string, []*entity.ContractSigning, error) {
	e, err := uc.contractRepo.FindContractByID(ctx, contractID)
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("find contract: %w", err)
	}
	if e == nil {
		return nil, "", "", nil, errors.NewNotFound("contract not found")
	}

	templateName := ""
	contractType := ""
	tmpl, err := uc.tmplRepo.FindByID(ctx, e.TemplateID)
	if err == nil && tmpl != nil {
		templateName = tmpl.Name
		contractType = string(tmpl.ContractType)
	}

	signings, err := uc.signingRepo.FindSigningsByContractID(ctx, contractID)
	if err != nil {
		return nil, "", "", nil, fmt.Errorf("find signings: %w", err)
	}

	return e, templateName, contractType, signings, nil
}

func (uc *ContractUsecase) DeleteContract(ctx context.Context, contractID string) error {
	e, err := uc.contractRepo.FindContractByID(ctx, contractID)
	if err != nil {
		return fmt.Errorf("find contract: %w", err)
	}
	if e == nil {
		return errors.NewNotFound("contract not found")
	}

	if err := e.CanDelete(); err != nil {
		return errors.NewInvalidInput(err.Error())
	}

	if err := uc.contractRepo.DeleteContract(ctx, contractID); err != nil {
		return fmt.Errorf("delete contract: %w", err)
	}
	return nil
}

func (uc *ContractUsecase) UpdateDraftContract(ctx context.Context, input models.UpdateContractInput) (*entity.Contract, error) {
	e, err := uc.contractRepo.FindContractByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("find contract: %w", err)
	}
	if e == nil {
		return nil, errors.NewNotFound("contract not found")
	}

	if err := e.UpdateDraft(
		input.Number,
		&input.StartDate,
		input.EndDate,
		input.Salary,
		input.DesignationID,
		input.DesignationTitle,
		input.Data,
		input.Templates,
	); err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	if err := uc.contractRepo.UpdateContractData(ctx, e); err != nil {
		return nil, fmt.Errorf("update contract data: %w", err)
	}
	return e, nil
}
