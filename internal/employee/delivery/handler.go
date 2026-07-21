package delivery

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	emplAdapter "hrms/internal/employee/adapter"
	"hrms/internal/employee/entity"
	"hrms/internal/employee/models"
	"hrms/internal/employee/usecase"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
)

type EmployeeHandler struct {
	uc          *usecase.EmployeeUsecase
	auditLogger *emplAdapter.AuditLogger
}

func NewEmployeeHandler(uc *usecase.EmployeeUsecase, auditLogger *emplAdapter.AuditLogger) *EmployeeHandler {
	return &EmployeeHandler{uc: uc, auditLogger: auditLogger}
}

func userIDFromCtx(c fiber.Ctx) *string {
	uid, ok := c.Locals("user_id").(string)
	if !ok || uid == "" {
		return nil
	}
	return &uid
}

func (h *EmployeeHandler) Create(c fiber.Ctx) error {
	var req CreateRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.CreateEmployeeInput{
		FullName:              req.FullName,
		EmployeeNumber:        req.EmployeeNumber,
		Phone:                 req.Phone,
		PersonalEmail:         req.PersonalEmail,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		PlaceOfBirth:          req.PlaceOfBirth,
		DateOfBirth:           req.DateOfBirth,
		JoinDate:              req.JoinDate,
		Gender:                req.Gender,
		Education:             req.Education,
		Address:               req.Address,
		DesignationID:         req.DesignationID,
		NationalID:            req.NationalID,
		Religion:              req.Religion,
		BankHolder:            req.BankHolder,
		BankName:              req.BankName,
		BankNumber:            req.BankNumber,
	}

	e, err := h.uc.Create(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, e.ID, "", emplAdapter.ActionCreate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"full_name": req.FullName, "employee_number": req.EmployeeNumber, "designation_id": req.DesignationID},
			)
		}
	}
	return response.Created(c, toResponse(e))
}

func (h *EmployeeHandler) Upsert(c fiber.Ctx) error {
	id := c.Params("id")
	var req CreateRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.CreateEmployeeInput{
		FullName:              req.FullName,
		EmployeeNumber:        req.EmployeeNumber,
		Phone:                 req.Phone,
		PersonalEmail:         req.PersonalEmail,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		PlaceOfBirth:          req.PlaceOfBirth,
		DateOfBirth:           req.DateOfBirth,
		JoinDate:              req.JoinDate,
		Gender:                req.Gender,
		Education:             req.Education,
		Address:               req.Address,
		DesignationID:         req.DesignationID,
		NationalID:            req.NationalID,
		Religion:              req.Religion,
		BankHolder:            req.BankHolder,
		BankName:              req.BankName,
		BankNumber:            req.BankNumber,
	}

	e, err := h.uc.Upsert(c.RequestCtx(), id, input)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toResponse(e))
}

func (h *EmployeeHandler) UpdateMyProfilePhoto(c fiber.Ctx) error {
	userID := userIDFromCtx(c)
	if userID == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	var req UpdateMyProfilePhotoRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.UpdateMyProfilePhotoInput{
		UserID:     *userID,
		DocumentID: req.DocumentID,
	}

	e, err := h.uc.UpdateMyProfilePhoto(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		h.auditLogger.Log(c.RequestCtx(), *userID, e.ID, "", emplAdapter.ActionPhotoUpdate,
			c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"document_id": req.DocumentID, "scope": "self"},
		)
	}
	return response.OK(c, toResponse(e))
}

func (h *EmployeeHandler) UpdateProfilePhoto(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateProfilePhotoRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.UpdateProfilePhotoInput{
		EmployeeID:     id,
		ProfilePhotoID: req.ProfilePhotoID,
	}

	e, err := h.uc.UpdateProfilePhoto(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, id, "", emplAdapter.ActionPhotoUpdate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"document_id": req.ProfilePhotoID, "scope": "admin"},
			)
		}
	}
	return response.OK(c, toResponse(e))
}

func (h *EmployeeHandler) UpdateContact(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateContactRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.UpdateContactInput{
		EmployeeID:            id,
		Phone:                 req.Phone,
		PersonalEmail:         req.PersonalEmail,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
	}

	e, err := h.uc.UpdateContact(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, id, "", emplAdapter.ActionUpdate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"section": "contact"},
			)
		}
	}
	return response.OK(c, toResponse(e))
}

func (h *EmployeeHandler) UpdateIdentity(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateIdentityRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.UpdateIdentityInput{
		EmployeeID:   id,
		FullName:     req.FullName,
		PlaceOfBirth: req.PlaceOfBirth,
		DateOfBirth:  req.DateOfBirth,
		Gender:       req.Gender,
		Education:    req.Education,
		Address:      req.Address,
		NationalID:   req.NationalID,
		Religion:     req.Religion,
	}

	e, err := h.uc.UpdateIdentity(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, id, "", emplAdapter.ActionUpdate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"section": "identity"},
			)
		}
	}
	return response.OK(c, toResponse(e))
}

func (h *EmployeeHandler) UpdateBank(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateBankRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.UpdateBankInput{
		EmployeeID: id,
		BankHolder: req.BankHolder,
		BankName:   req.BankName,
		BankNumber: req.BankNumber,
	}

	e, err := h.uc.UpdateBank(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, id, "", emplAdapter.ActionUpdate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"section": "bank"},
			)
		}
	}
	return response.OK(c, toResponse(e))
}

func (h *EmployeeHandler) UpdateEmployeeNumber(c fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateEmployeeNumberRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.UpdateEmployeeNumberInput{
		EmployeeID:     id,
		EmployeeNumber: req.EmployeeNumber,
	}

	e, err := h.uc.UpdateEmployeeNumber(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			h.auditLogger.Log(c.RequestCtx(), *actorID, id, "", emplAdapter.ActionUpdate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"section": "employee_number"},
			)
		}
	}
	return response.OK(c, toResponse(e))
}

func (h *EmployeeHandler) ChangeDesignation(c fiber.Ctx) error {
	var req ChangeDesignationRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	input := models.ChangeDesignationInput{
		EmployeeIDs:   req.EmployeeIDs,
		DesignationID: req.DesignationID,
	}

	employees, err := h.uc.ChangeDesignation(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID := userIDFromCtx(c)
		if actorID != nil {
			for _, e := range employees {
				h.auditLogger.Log(c.RequestCtx(), *actorID, e.ID, "", emplAdapter.ActionUpdate,
					c.IP(), string(c.RequestCtx().UserAgent()),
					map[string]any{"section": "designation", "designation_id": req.DesignationID},
				)
			}
		}
	}

	items := make([]EmployeeResponse, 0, len(employees))
	for _, e := range employees {
		items = append(items, toResponse(e))
	}
	return response.OK(c, items)
}

func (h *EmployeeHandler) MyProfileCompletion(c fiber.Ctx) error {
	userID := userIDFromCtx(c)
	if userID == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	result, err := h.uc.GetProfileCompletion(c.RequestCtx(), *userID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, result)
}

func (h *EmployeeHandler) PeekNextNumber(c fiber.Ctx) error {
	designationID := c.Query("designation_id")
	if designationID == "" {
		return response.Error(c, errors.NewInvalidInput("designation_id is required"))
	}

	input := models.PeekNextNumberInput{DesignationID: designationID}
	result, err := h.uc.PeekNextEmployeeNumber(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, result)
}

func (h *EmployeeHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	input := models.ListEmployeeInput{
		Page:          page,
		PerPage:       perPage,
		SearchName:    c.Query("search"),
		Status:        c.Query("status"),
		Gender:        c.Query("gender"),
		DesignationID: c.Query("designation_id"),
	}

	result, err := h.uc.List(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginate(c, result.Items, page, perPage, result.Total)
}

func (h *EmployeeHandler) FindByUserID(c fiber.Ctx) error {
	userID := userIDFromCtx(c)
	if userID == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	result, err := h.uc.GetMe(c.RequestCtx(), *userID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, result)
}

func (h *EmployeeHandler) FindByID(c fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.uc.GetByIDWithDetails(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, fromEmployeeResult(result))
}

func toResponse(e *entity.Employee) EmployeeResponse {
	var dateOfBirth *string
	if e.DateOfBirth != nil {
		s := e.DateOfBirth.String()
		dateOfBirth = &s
	}
	var joinDate *string
	if e.JoinDate != nil {
		s := e.JoinDate.String()
		joinDate = &s
	}

	return EmployeeResponse{
		ID:                    e.ID,
		UserID:                e.UserID,
		FullName:              e.FullName,
		EmployeeNumber:        e.EmployeeNumber.String(),
		Phone:                 e.Phone.String(),
		PersonalEmail:         e.PersonalEmail,
		EmergencyContactName:  e.EmergencyContactName,
		EmergencyContactPhone: e.EmergencyContactPhone.String(),
		PlaceOfBirth:          e.PlaceOfBirth,
		DateOfBirth:           dateOfBirth,
		JoinDate:              joinDate,
		Gender:                string(e.Gender),
		Education:             e.Education,
		Status:                string(e.Status),
		Address:               e.Address,
		DesignationID:         e.DesignationID,
		NationalID:            e.NationalID,
		Religion:              string(e.Religion),
		BankHolder:            e.BankHolder(),
		BankName:              e.BankName(),
		BankNumber:            e.BankNumber(),
		ProfilePhotoURL:       "",
		IsActive:              e.IsActive,
		CreatedAt:             e.CreatedAt,
		UpdatedAt:             e.UpdatedAt,
	}
}
