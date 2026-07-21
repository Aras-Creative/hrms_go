package delivery

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"

	contractAdapter "hrms/internal/contract/adapter"
	"hrms/internal/contract/entity"
	"hrms/internal/contract/models"
	"hrms/internal/contract/usecase"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
)

type Notifier interface {
	Notify(ctx context.Context, userID, ntype, title, body, resource, resourceID string) error
}

const (
	notifTypeEmployment = "employment"
)

type ContractHandler struct {
	uc          *usecase.ContractUsecase
	signUC      *usecase.SigningUsecase
	renderUC    *usecase.RenderUsecase
	docUC       *usecase.DocumentUsecase
	terminateUC *usecase.TerminationUsecase
	auditLogger *contractAdapter.AuditLogger
	notifUC     Notifier
}

func NewContractHandler(uc *usecase.ContractUsecase, signUC *usecase.SigningUsecase, renderUC *usecase.RenderUsecase, docUC *usecase.DocumentUsecase, terminateUC *usecase.TerminationUsecase, auditLogger *contractAdapter.AuditLogger, notifUC Notifier) *ContractHandler {
	return &ContractHandler{uc: uc, signUC: signUC, renderUC: renderUC, docUC: docUC, terminateUC: terminateUC, auditLogger: auditLogger, notifUC: notifUC}
}

func parseID(c fiber.Ctx) (string, error) {
	return response.ParseParamID(c, "id")
}

func (h *ContractHandler) CreateTemplate(c fiber.Ctx) error {
	var req CreateTemplateRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.CreateTemplateInput{
		Name:         req.Name,
		ContractType: entity.ContractType(req.ContractType),
		Description:  req.Description,
		Data:         toEntityData(req.Data),
		Templates:    toEntityPartials(req.Templates),
	}

	e, err := h.uc.CreateTemplate(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		userID, _ := c.Locals("user_id").(string)
		h.auditLogger.Log(c.RequestCtx(), userID, "contract_template", e.ID, "",
			contractAdapter.ActionTemplateCreate, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"name": req.Name, "contract_type": req.ContractType},
		)
	}
	return response.Created(c, toTemplateResponse(e))
}

func (h *ContractHandler) GetTemplate(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return response.Error(c, err)
	}
	e, err := h.uc.GetTemplate(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toTemplateResponse(e))
}

func (h *ContractHandler) GetTemplatePrefill(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return response.Error(c, err)
	}
	e, err := h.uc.GetTemplate(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toTemplatePrefillResponse(e))
}

func (h *ContractHandler) ListTemplates(c fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(c.Query("per_page", "20"))
	if err != nil || perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		b := v == "true"
		isActive = &b
	}

	input := models.ListTemplateInput{
		Page:         page,
		PerPage:      perPage,
		SearchName:   c.Query("search"),
		ContractType: c.Query("contract_type"),
		IsActive:     isActive,
	}

	result, err := h.uc.ListTemplates(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, TemplateListResponse{
		Items: toListItemResponses(result.Entities),
		Total: result.Total,
	})
}

func (h *ContractHandler) UpdateTemplate(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req UpdateTemplateRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	input := models.UpdateTemplateInput{
		ID:           id,
		Name:         req.Name,
		ContractType: entity.ContractType(req.ContractType),
		Description:  req.Description,
		IsActive:     isActive,
		Data:         toEntityData(req.Data),
		Templates:    toEntityPartials(req.Templates),
	}

	e, err := h.uc.UpdateTemplate(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		userID, _ := c.Locals("user_id").(string)
		h.auditLogger.Log(c.RequestCtx(), userID, "contract_template", id, "",
			contractAdapter.ActionTemplateUpdate, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"name": req.Name, "contract_type": req.ContractType, "is_active": isActive},
		)
	}
	return response.OK(c, toTemplateResponse(e))
}

func (h *ContractHandler) DeleteTemplate(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return response.Error(c, err)
	}
	if err := h.uc.DeleteTemplate(c.RequestCtx(), id); err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		userID, _ := c.Locals("user_id").(string)
		h.auditLogger.Log(c.RequestCtx(), userID, "contract_template", id, "",
			contractAdapter.ActionTemplateDelete, c.IP(), string(c.RequestCtx().UserAgent()),
			nil,
		)
	}
	return response.NoContent(c)
}

func (h *ContractHandler) CreateContract(c fiber.Ctx) error {
	var req CreateContractRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input, err := req.ToInput()
	if err != nil {
		return response.Error(c, err)
	}

	result, err := h.uc.CreateContract(c.RequestCtx(), *input)
	if err != nil {
		return response.Error(c, err)
	}

	items := make([]*ContractCreatedItem, len(result.Contracts))
	for i, e := range result.Contracts {
		items[i] = toContractCreatedItem(e)

		if h.auditLogger != nil {
			userID, _ := c.Locals("user_id").(string)
			h.auditLogger.Log(c.RequestCtx(), userID, "contract", e.ID, "",
				contractAdapter.ActionCreate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{
					"employee_id":     e.EmployeeID,
					"template_id":     e.TemplateID,
					"contract_number": e.Number,
					"status":          "draft",
				},
			)
		}

		if h.notifUC != nil {
			empUserID, _ := h.uc.FindUserIDByEmployeeID(c.RequestCtx(), e.EmployeeID)
			if empUserID != "" {
				h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeEmployment,
					"Kontrak Baru",
					fmt.Sprintf("Kontrak %s telah dibuat untuk Anda", e.Number),
					"contract", e.ID,
				)
			}
		}
	}
	return response.Created(c, ContractCreatedResponse{Signed: len(items), Contracts: items})
}

func (h *ContractHandler) DeleteContract(c fiber.Ctx) error {
	id := c.Params("id")

	if err := h.uc.DeleteContract(c.RequestCtx(), id); err != nil {
		return response.Error(c, err)
	}
	return response.NoContent(c)
}

func (h *ContractHandler) CheckActiveContracts(c fiber.Ctx) error {
	var req CheckActiveContractRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.CheckActiveContractInput{
		EmployeeIDs: req.EmployeeIDs,
	}

	result, err := h.uc.CheckActiveContracts(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, result)
}

func (h *ContractHandler) ListContracts(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	input := models.ListContractInput{
		Page:          page,
		PerPage:       perPage,
		Status:        c.Query("status"),
		ContractType:  c.Query("contract_type"),
		DesignationID: c.Query("designation_id"),
	}

	result, err := h.uc.ListContractsWithDetail(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, toListContractResponse(result.Items, result.Total, result.SigningsByContract, result.EmployeeBriefs))
}

func (h *ContractHandler) MyActiveContract(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("unauthorized"))
	}

	e, err := h.uc.GetMyActiveContract(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, err)
	}
	if e == nil {
		return response.OK(c, nil)
	}

	return response.OK(c, toActiveContractResponse(e))
}

func (h *ContractHandler) GetEmployeeContract(c fiber.Ctx) error {
	employeeID := c.Params("id")
	if employeeID == "" {
		return response.Error(c, errors.NewInvalidInput("employee id is required"))
	}

	e, err := h.uc.GetEmployeeContract(c.RequestCtx(), employeeID)
	if err != nil {
		return response.Error(c, err)
	}
	if e == nil {
		return response.OK(c, nil)
	}

	return response.OK(c, toEmployeeContractResponse(e))
}

func (h *ContractHandler) GetContract(c fiber.Ctx) error {
	id := c.Params("id")

	e, templateName, contractType, signings, err := h.uc.GetContractDetail(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, toContractDetailResponse(e, templateName, contractType, signings))
}

func (h *ContractHandler) GetDraftContract(c fiber.Ctx) error {
	id := c.Params("id")

	e, _, _, _, err := h.uc.GetContractDetail(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}

	if e.Status != entity.ContractStatusDraft {
		return response.Error(c, errors.NewInvalidInput("contract is not in draft status"))
	}

	return response.OK(c, toContractResponse(e))
}

func (h *ContractHandler) UpdateDraftContract(c fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateDraftContractRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input, err := req.ToInput(id)
	if err != nil {
		return response.Error(c, err)
	}

	e, err := h.uc.UpdateDraftContract(c.RequestCtx(), *input)
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		userID, _ := c.Locals("user_id").(string)
		h.auditLogger.Log(c.RequestCtx(), userID, "contract", e.ID, "",
			contractAdapter.ActionUpdate, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{
				"status": "draft",
			},
		)
	}

	return response.OK(c, toContractResponse(e))
}

func (h *ContractHandler) CountSoonExpired(c fiber.Ctx) error {
	count, err := h.uc.CountSoonExpired(c.RequestCtx(), 14)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, CountSoonExpiredResponse{SoonExpired: count})
}

func (h *ContractHandler) PendingContracts(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("unauthorized"))
	}

	count, err := h.uc.CountPendingContracts(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, PendingContractsResponse{Pending: count})
}

func (h *ContractHandler) MyContracts(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("unauthorized"))
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	input := models.ListContractInput{
		Page:         page,
		PerPage:      perPage,
		Status:       c.Query("status"),
		ExcludeDraft: true,
	}

	result, err := h.uc.ListMyContractsWithDetail(c.RequestCtx(), input, userID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, toListContractResponse(result.Items, result.Total, result.SigningsByContract, result.EmployeeBriefs))
}

func (h *ContractHandler) PreviewContract(c fiber.Ctx) error {
	id := c.Params("id")

	pdf, err := h.renderUC.Preview(c.RequestCtx(), id, c.Query("signatory_name", "Pihak Pertama"), c.Query("signatory_title", "Direktur"))
	if err != nil {
		return response.Error(c, err)
	}

	c.Type("pdf")
	return c.Send(pdf)
}

func (h *ContractHandler) DownloadContract(c fiber.Ctx) error {
	id := c.Params("id")

	reader, meta, err := h.docUC.DownloadPDF(c.Context(), id)
	if err != nil {
		return response.Error(c, err)
	}

	c.Set("Content-Type", meta.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, meta.OriginalName))
	c.Set("Content-Length", strconv.FormatInt(meta.SizeBytes, 10))

	return c.SendStream(reader, int(meta.SizeBytes))
}

func (h *ContractHandler) SignBySecondParty(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("unauthorized"))
	}

	var req SignContractRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.BulkSignContractInput{
		ContractIDs:     req.ContractIDs,
		SignedBy:        req.SignedBy,
		SignedByName:    req.SignedByName,
		SignedByTitle:   req.SignedByTitle,
		Place:           req.Place,
		SignatureBase64: req.SignatureBase64,
	}

	result, err := h.signUC.BulkSignAsSecondParty(c.RequestCtx(), input, userID)
	if err != nil {
		return response.Error(c, err)
	}

	items := make([]*ContractResponse, len(result))
	for i, e := range result {
		items[i] = toContractResponse(e)

		if h.auditLogger != nil {
			h.auditLogger.Log(c.RequestCtx(), userID, "contract", e.ID, "",
				contractAdapter.ActionSignSecondParty, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{
					"party":           "second",
					"signed_by":       req.SignedBy,
					"signed_by_name":  req.SignedByName,
					"signed_by_title": req.SignedByTitle,
					"place":           req.Place,
					"contract_number": e.Number,
				},
			)
		}

		// Notify employee that a new contract is ready to sign
		if h.notifUC != nil {
			empUserID, _ := h.uc.FindUserIDByEmployeeID(c.RequestCtx(), e.EmployeeID)
			if empUserID != "" {
				h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeEmployment,
					"Kontrak Baru",
					fmt.Sprintf("Kontrak %s siap ditandatangani", e.Number),
					"contract", e.ID,
				)
			}
		}
	}
	return response.Created(c, SignContractsResponse{Signed: len(items), Contracts: items})
}

func (h *ContractHandler) GeneratePDF(c fiber.Ctx) error {
	id := c.Params("id")

	e, _, _, signings, err := h.uc.GetContractDetail(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}

	if len(signings) == 0 {
		return response.Error(c, errors.NewInvalidInput("contract has no signings"))
	}

	signatoryName := signings[0].SignedByName
	signatoryTitle := signings[0].SignedByTitle

	docID, contentHash, err := h.docUC.StorePDF(c.RequestCtx(), e.ID, e.Number, signatoryName, signatoryTitle)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, GeneratePDFResponse{DocumentID: docID, ContentHash: contentHash})
}

func (h *ContractHandler) TerminateContract(c fiber.Ctx) error {
	id := c.Params("id")

	var req TerminateContractRequest
	body := c.Body()
	if len(body) > 0 {
		if err := c.Bind().Body(&req); err != nil {
			return err
		}
	}

	terminationDate, err := req.GetTerminationDate()
	if err != nil {
		return response.Error(c, err)
	}

	e, _, _, _, err := h.uc.GetContractDetail(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}

	input := usecase.TerminateContractInput{
		ContractID:      id,
		TerminationDate: terminationDate,
	}

	if err := h.terminateUC.TerminateContract(c.RequestCtx(), input); err != nil {
		return response.Error(c, err)
	}

	if h.notifUC != nil {
		empUserID, _ := h.uc.FindUserIDByEmployeeID(c.RequestCtx(), e.EmployeeID)
		if empUserID != "" {
			h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeEmployment,
				"Kontrak Diputus",
				fmt.Sprintf("Kontrak %s telah diputus per %s", e.Number, req.TerminationDate),
				"contract", e.ID,
			)
		}
	}

	if h.auditLogger != nil {
		userID, _ := c.Locals("user_id").(string)
		h.auditLogger.Log(c.RequestCtx(), userID, "contract", id, "",
			contractAdapter.ActionTerminate, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{
				"termination_date": req.TerminationDate,
			},
		)
	}

	return response.OK(c, nil)
}

func (h *ContractHandler) SignByFirstParty(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("unauthorized"))
	}

	var req SignContractRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.BulkSignContractInput{
		ContractIDs:     req.ContractIDs,
		Party:           req.Party,
		SignedBy:        req.SignedBy,
		SignedByName:    req.SignedByName,
		SignedByTitle:   req.SignedByTitle,
		Place:           req.Place,
		SignatureBase64: req.SignatureBase64,
	}

	result, err := h.signUC.BulkSign(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}

	items := make([]*ContractResponse, len(result))
	for i, e := range result {
		items[i] = toContractResponse(e)

		if h.auditLogger != nil {
			h.auditLogger.Log(c.RequestCtx(), userID, "contract", e.ID, "",
				contractAdapter.ActionSignFirstParty, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{
					"party":           input.Party,
					"signed_by":       req.SignedBy,
					"signed_by_name":  req.SignedByName,
					"signed_by_title": req.SignedByTitle,
					"place":           req.Place,
					"contract_number": e.Number,
				},
			)
		}

		if h.notifUC != nil {
			empUserID, _ := h.uc.FindUserIDByEmployeeID(c.RequestCtx(), e.EmployeeID)
			if empUserID != "" {
				h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeEmployment,
					"Kontrak Ditandatangani",
					fmt.Sprintf("Kontrak %s telah ditandatangani oleh pihak pertama", e.Number),
					"contract", e.ID,
				)
			}
		}
	}
	return response.Created(c, SignContractsResponse{Signed: len(items), Contracts: items})
}
