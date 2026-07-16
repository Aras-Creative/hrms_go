package delivery

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"

	"hrms/internal/leave/adapter"
	"hrms/internal/leave/entity"
	"hrms/internal/leave/models"
	"hrms/internal/leave/usecase"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
)

type AttachmentResolver interface {
	Resolve(ctx context.Context, docID string) (string, error)
}

type Notifier interface {
	Notify(ctx context.Context, userID, ntype, title, body, resource, resourceID string) error
}

type PhotoURLResolver interface {
	ResolveURLs(ctx context.Context, documentIDs []string) (map[string]string, error)
}

const (
	notifTypeLeave = "leave"
)

type LeaveHandler struct {
	uc            *usecase.LeaveUsecase
	resolver      AttachmentResolver
	photoResolver PhotoURLResolver
	auditLogger   *adapter.AuditLogger
	notifUC       Notifier
}

func NewLeaveHandler(uc *usecase.LeaveUsecase, resolver AttachmentResolver, photoResolver PhotoURLResolver, auditLogger *adapter.AuditLogger, notifUC Notifier) *LeaveHandler {
	return &LeaveHandler{uc: uc, resolver: resolver, photoResolver: photoResolver, auditLogger: auditLogger, notifUC: notifUC}
}

func userIDFromCtx(c fiber.Ctx) *string {
	uid, ok := c.Locals("user_id").(string)
	if !ok || uid == "" {
		return nil
	}
	return &uid
}

// ---- Leave Types ----

func (h *LeaveHandler) CreateType(c fiber.Ctx) error {
	var req CreateLeaveTypeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}

	lt, err := h.uc.CreateLeaveType(c.RequestCtx(), models.CreateLeaveTypeInput{
		Name:        req.Name,
		DefaultDays: req.DefaultDays,
		IsPaid:      req.IsPaid,
		IsUnlimited: req.IsUnlimited,
		IsHalfDay:   req.IsHalfDay,
	})
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, "leave_type", lt.ID, "",
				adapter.ActionTypeCreate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name, "default_days": req.DefaultDays, "is_paid": req.IsPaid},
			)
		}
	}
	return response.Created(c, toLeaveTypeResponse(lt))
}

func (h *LeaveHandler) GetType(c fiber.Ctx) error {
	id := c.Params("id")
	lt, err := h.uc.GetLeaveTypeByID(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toLeaveTypeResponse(lt))
}

func (h *LeaveHandler) ListTypes(c fiber.Ctx) error {
	list, err := h.uc.GetAllLeaveTypes(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}
	resp := make([]LeaveTypeResponse, 0, len(list))
	for _, lt := range list {
		resp = append(resp, toLeaveTypeResponse(lt))
	}
	return response.OK(c, resp)
}

func (h *LeaveHandler) ListTypeOptions(c fiber.Ctx) error {
	list, err := h.uc.GetAllLeaveTypes(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}
	opts := make([]response.Option, 0, len(list))
	for _, lt := range list {
		opts = append(opts, response.Option{
			Value: lt.ID,
			Label: lt.Name,
			Extra: lt.IsUnlimited,
		})
	}
	return response.Options(c, opts)
}

func (h *LeaveHandler) UpdateType(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateLeaveTypeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}

	lt, err := h.uc.UpdateLeaveType(c.RequestCtx(), id, models.UpdateLeaveTypeInput{
		Name:        req.Name,
		DefaultDays: req.DefaultDays,
		IsPaid:      req.IsPaid,
		IsUnlimited: req.IsUnlimited,
		IsHalfDay:   req.IsHalfDay,
	})
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, "leave_type", id, "",
				adapter.ActionTypeUpdate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name, "default_days": req.DefaultDays},
			)
		}
	}
	return response.OK(c, toLeaveTypeResponse(lt))
}

func (h *LeaveHandler) DeleteType(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.DeleteLeaveType(c.RequestCtx(), id); err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, "leave_type", id, "",
				adapter.ActionTypeDelete, c.IP(), string(c.RequestCtx().UserAgent()),
				nil,
			)
		}
	}
	return response.NoContent(c)
}

func toLeaveTypeResponse(lt *entity.LeaveType) LeaveTypeResponse {
	return LeaveTypeResponse{
		ID:          lt.ID,
		Name:        lt.Name,
		DefaultDays: lt.DefaultDays,
		IsPaid:      lt.IsPaid,
		IsUnlimited: lt.IsUnlimited,
		IsHalfDay:   lt.IsHalfDay,
		IsActive:    lt.IsActive,
		CreatedAt:   lt.CreatedAt,
		UpdatedAt:   lt.UpdatedAt,
	}
}

// ---- Leave Balances ----

func (h *LeaveHandler) UpdateBalance(c fiber.Ctx) error {
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		return response.Error(c, errors.NewInvalidInput("employee_id is required"))
	}
	leaveTypeID := c.Query("leave_type_id")
	if leaveTypeID == "" {
		return response.Error(c, errors.NewInvalidInput("leave_type_id is required"))
	}
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year == 0 {
		return response.Error(c, errors.NewInvalidInput("year is required"))
	}

	var req UpdateLeaveBalanceRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}

	b, err := h.uc.UpdateBalance(c.RequestCtx(), models.UpdateLeaveBalanceInput{
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
		TotalDays:   req.TotalDays,
	})
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, "leave_balance", b.ID, "",
				adapter.ActionBalanceUpdate, c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"employee_id": employeeID, "leave_type_id": leaveTypeID, "total_days": req.TotalDays, "year": year},
			)
		}
	}
	return response.OK(c, LeaveBalanceResponse{
		ID:          b.ID,
		EmployeeID:  b.EmployeeID,
		LeaveTypeID: b.LeaveTypeID,
		Year:        b.Year,
		TotalDays:   b.TotalDays,
		UsedDays:    b.UsedDays,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	})
}

func (h *LeaveHandler) ListBalances(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	year, _ := strconv.Atoi(c.Query("year"))

	result, err := h.uc.ListBalances(c.RequestCtx(), models.ListBalanceInput{
		LeaveTypeID: c.Query("leave_type_id"),
		Search:      c.Query("search"),
		Year:        year,
		Page:        page,
		PerPage:     perPage,
	})
	if err != nil {
		return response.Error(c, err)
	}

	items := make([]LeaveBalanceResponse, len(result.Items))
	for i, item := range result.Items {
		items[i] = toBalanceResponse(&item)
	}

	var photoIDs []string
	for _, item := range items {
		if item.ProfilePhotoID != nil && *item.ProfilePhotoID != "" {
			photoIDs = append(photoIDs, *item.ProfilePhotoID)
		}
	}
	if len(photoIDs) > 0 && h.photoResolver != nil {
		urls, err := h.photoResolver.ResolveURLs(c.RequestCtx(), photoIDs)
		if err == nil {
			for i, item := range items {
				if item.ProfilePhotoID != nil {
					if url, ok := urls[*item.ProfilePhotoID]; ok {
						items[i].ProfilePhotoURL = url
					}
				}
			}
		}
	}

	return response.Paginate(c, items, page, perPage, result.Total)
}

func (h *LeaveHandler) GetBalance(c fiber.Ctx) error {
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		return response.Error(c, errors.NewInvalidInput("employee_id is required"))
	}
	leaveTypeID := c.Query("leave_type_id")
	if leaveTypeID == "" {
		return response.Error(c, errors.NewInvalidInput("leave_type_id is required"))
	}

	b, err := h.uc.GetEmployeeBalance(c.RequestCtx(), employeeID, leaveTypeID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toBalanceResponse(b))
}

func toBalanceResponse(b *models.LeaveBalance) LeaveBalanceResponse {
	return LeaveBalanceResponse{
		ID:             b.ID,
		EmployeeID:     b.EmployeeID,
		EmployeeName:   b.EmployeeName,
		EmployeeNumber: b.EmployeeNumber,
		ProfilePhotoID: b.ProfilePhotoID,
		LeaveTypeID:    b.LeaveTypeID,
		LeaveTypeName:  b.LeaveTypeName,
		Year:           b.Year,
		TotalDays:      b.TotalDays,
		UsedDays:       b.UsedDays,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
	}
}

// ---- Leave Submissions ----

func (h *LeaveHandler) Submit(c fiber.Ctx) error {
	var req CreateLeaveSubmissionRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}

	authUserID := userIDFromCtx(c)
	if authUserID == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	startDate, err := models.ParseDate(req.StartDate)
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid start_date: "+err.Error()))
	}
	endDate, err := models.ParseDate(req.EndDate)
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid end_date: "+err.Error()))
	}

	input := models.CreateLeaveSubmissionInput{
		UserID:       *authUserID,
		LeaveTypeID:  req.LeaveTypeID,
		StartDate:    startDate,
		EndDate:      endDate,
		Reason:       req.Reason,
		AttachmentID: req.AttachmentID,
	}

	s, err := h.uc.SubmitLeave(c.RequestCtx(), input)
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorName, _ := h.uc.ResolveActorName(c.RequestCtx(), *authUserID)
		h.auditLogger.Log(c.RequestCtx(), *authUserID, "leave_submission", s.ID, "",
			adapter.ActionSubmit, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{
				"actor_name":  actorName,
				"employee_id": s.EmployeeID,
				"start_date":  s.StartDate.Format("2006-01-02"),
				"end_date":    s.EndDate.Format("2006-01-02"),
				"days":        s.Days,
			},
		)
	}

	if h.notifUC != nil {
		actorName, _ := h.uc.ResolveActorName(c.RequestCtx(), *authUserID)
		lt, _ := h.uc.GetLeaveTypeByID(c.RequestCtx(), s.LeaveTypeID)
		typeName := "cuti"
		if lt != nil {
			typeName = lt.Name
		}
		dateStr := fmt.Sprintf("%s - %s", s.StartDate.Format("02 Jan 2006"), s.EndDate.Format("02 Jan 2006"))

		// Notify the submitter
		h.notifUC.Notify(c.RequestCtx(), *authUserID, notifTypeLeave,
			"Pengajuan Cuti",
			fmt.Sprintf("Anda mengajukan %s pada %s", typeName, dateStr),
			"leave_submission", s.ID,
		)

		// Notify all admins
		adminIDs, _ := h.uc.FindAdminIDs(c.RequestCtx())
		for _, adminID := range adminIDs {
			if adminID != *authUserID {
				h.notifUC.Notify(c.RequestCtx(), adminID, notifTypeLeave,
					"Pengajuan Cuti Baru",
					fmt.Sprintf("%s mengajukan %s pada %s", actorName, typeName, dateStr),
					"leave_submission", s.ID,
				)
			}
		}
	}

	return response.Created(c, toSubmissionResponse(s))
}

func (h *LeaveHandler) ListAllSubmissions(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		d, err := models.ParseDate(s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid start_date: "+err.Error()))
		}
		startDate = &d
	}
	if s := c.Query("end_date"); s != "" {
		d, err := models.ParseDate(s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid end_date: "+err.Error()))
		}
		endDate = &d
	}

	result, err := h.uc.ListAllSubmissions(c.RequestCtx(), models.ListAllSubmissionInput{
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		PerPage:   perPage,
	})
	if err != nil {
		return response.Error(c, err)
	}

	items := make([]LeaveSubmissionResponse, len(result.Items))
	for i, item := range result.Items {
		items[i] = toSubmissionResponseFromModel(item)
	}

	// Resolve profile photo URLs
	var photoIDs []string
	for _, item := range items {
		if item.ProfilePhotoID != nil && *item.ProfilePhotoID != "" {
			photoIDs = append(photoIDs, *item.ProfilePhotoID)
		}
	}
	if len(photoIDs) > 0 && h.photoResolver != nil {
		urls, err := h.photoResolver.ResolveURLs(c.RequestCtx(), photoIDs)
		if err == nil {
			for i := range items {
				if items[i].ProfilePhotoID != nil {
					if url, ok := urls[*items[i].ProfilePhotoID]; ok {
						items[i].ProfilePhotoURL = url
					}
				}
			}
		}
	}

	return response.Paginate(c, items, page, perPage, result.Total)
}

func (h *LeaveHandler) ListMySubmissions(c fiber.Ctx) error {
	userID := userIDFromCtx(c)
	if userID == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	result, err := h.uc.ListMySubmissions(c.RequestCtx(), models.ListSubmissionInput{
		UserID:  *userID,
		Status:  c.Query("status"),
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return response.Error(c, err)
	}

	items := make([]LeaveSubmissionResponse, len(result.Items))
	for i, item := range result.Items {
		items[i] = toSubmissionResponseFromModel(item)
	}
	return response.Paginate(c, items, page, perPage, result.Total)
}

func (h *LeaveHandler) GetSubmission(c fiber.Ctx) error {
	id := c.Params("id")
	s, err := h.uc.GetSubmissionByID(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}

	submissionRes := toSubmissionResponseFromModel(*s)

	var history []AuditEntryResponse
	if h.auditLogger != nil {
		entries, err := h.auditLogger.ListByResource(c.RequestCtx(), "leave_submission", id)
		if err == nil {
			history = make([]AuditEntryResponse, 0, len(entries))
			for _, e := range entries {
				actorName, _ := e.Payload["actor_name"].(string)
				history = append(history, AuditEntryResponse{
					ID:        e.ID,
					Action:    e.Action,
					ActorID:   e.ActorID,
					ActorName: actorName,
					Payload:   e.Payload,
					CreatedAt: e.CreatedAt,
				})
			}
		}
	}

	return response.OK(c, SubmissionDetailResponse{
		Submission: submissionRes,
		History:    history,
	})
}

func (h *LeaveHandler) GetAttachment(c fiber.Ctx) error {
	id := c.Params("id")
	s, err := h.uc.GetSubmissionByID(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}
	if s.AttachmentID == nil || *s.AttachmentID == "" {
		return response.Error(c, errors.NewNotFound("submission has no attachment"))
	}

	url, err := h.resolver.Resolve(c.RequestCtx(), *s.AttachmentID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, AttachmentResponse{URL: url})
}

func (h *LeaveHandler) ApproveSubmission(c fiber.Ctx) error {
	id := c.Params("id")
	adminID := userIDFromCtx(c)
	if adminID == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	s, err := h.uc.ApproveSubmission(c.RequestCtx(), id, *adminID)
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorName, _ := h.uc.ResolveActorName(c.RequestCtx(), *adminID)
		h.auditLogger.Log(c.RequestCtx(), *adminID, "leave_submission", id, "",
			adapter.ActionApprove, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{
				"actor_name":  actorName,
				"employee_id": s.EmployeeID,
				"start_date":  s.StartDate.Format("2006-01-02"),
				"end_date":    s.EndDate.Format("2006-01-02"),
				"days":        s.Days,
			},
		)
	}

	if h.notifUC != nil {
		actorName, _ := h.uc.ResolveActorName(c.RequestCtx(), *adminID)
		empUserID, _ := h.uc.FindUserIDByEmployeeID(c.RequestCtx(), s.EmployeeID)
		if empUserID != "" {
			h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeLeave,
				"Cuti Disetujui",
				fmt.Sprintf("%s menyetujui cuti %s - %s", actorName, s.StartDate.Format("02 Jan 2006"), s.EndDate.Format("02 Jan 2006")),
				"leave_submission", s.ID,
			)
		}
	}

	return response.OK(c, toSubmissionResponse(s))
}

func (h *LeaveHandler) RejectSubmission(c fiber.Ctx) error {
	id := c.Params("id")
	s, err := h.uc.RejectSubmission(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}

	adminID := userIDFromCtx(c)
	if h.auditLogger != nil && adminID != nil {
		actorName, _ := h.uc.ResolveActorName(c.RequestCtx(), *adminID)
		h.auditLogger.Log(c.RequestCtx(), *adminID, "leave_submission", id, "",
			adapter.ActionReject, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{
				"actor_name":  actorName,
				"employee_id": s.EmployeeID,
				"start_date":  s.StartDate.Format("2006-01-02"),
				"end_date":    s.EndDate.Format("2006-01-02"),
				"days":        s.Days,
			},
		)
	}

	if h.notifUC != nil {
		actorName, _ := h.uc.ResolveActorName(c.RequestCtx(), *adminID)
		empUserID, _ := h.uc.FindUserIDByEmployeeID(c.RequestCtx(), s.EmployeeID)
		if empUserID != "" {
			h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeLeave,
				"Cuti Ditolak",
				fmt.Sprintf("%s menolak cuti %s - %s", actorName, s.StartDate.Format("02 Jan 2006"), s.EndDate.Format("02 Jan 2006")),
				"leave_submission", s.ID,
			)
		}
	}

	return response.OK(c, toSubmissionResponse(s))
}

func (h *LeaveHandler) CancelSubmission(c fiber.Ctx) error {
	id := c.Params("id")
	userID := userIDFromCtx(c)
	if userID == nil {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	s, err := h.uc.CancelSubmission(c.RequestCtx(), id, *userID)
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorName, _ := h.uc.ResolveActorName(c.RequestCtx(), *userID)
		h.auditLogger.Log(c.RequestCtx(), *userID, "leave_submission", id, "",
			adapter.ActionCancel, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{
				"actor_name":  actorName,
				"employee_id": s.EmployeeID,
				"start_date":  s.StartDate.Format("2006-01-02"),
				"end_date":    s.EndDate.Format("2006-01-02"),
				"days":        s.Days,
			},
		)
	}

	if h.notifUC != nil {
		actorName, _ := h.uc.ResolveActorName(c.RequestCtx(), *userID)
		dateStr := fmt.Sprintf("%s - %s", s.StartDate.Format("02 Jan 2006"), s.EndDate.Format("02 Jan 2006"))

		adminIDs, _ := h.uc.FindAdminIDs(c.RequestCtx())
		for _, adminID := range adminIDs {
			if adminID != *userID {
				h.notifUC.Notify(c.RequestCtx(), adminID, notifTypeLeave,
					"Cuti Dibatalkan",
					fmt.Sprintf("%s membatalkan cuti pada %s", actorName, dateStr),
					"leave_submission", s.ID,
				)
			}
		}
	}

	return response.OK(c, toSubmissionResponse(s))
}

func toSubmissionResponse(s *entity.LeaveSubmission) LeaveSubmissionResponse {
	return LeaveSubmissionResponse{
		ID:           s.ID,
		EmployeeID:   s.EmployeeID,
		LeaveTypeID:  s.LeaveTypeID,
		StartDate:    s.StartDate,
		EndDate:      s.EndDate,
		Days:         s.Days,
		Reason:       s.Reason,
		AttachmentID: s.AttachmentID,
		Status:       string(s.Status),
		ApprovedBy:   s.ApprovedBy,
		ApprovedAt:   s.ApprovedAt,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

func toSubmissionResponseFromModel(m models.LeaveSubmission) LeaveSubmissionResponse {
	return LeaveSubmissionResponse{
		ID:             m.ID,
		EmployeeID:     m.EmployeeID,
		EmployeeName:   m.EmployeeName,
		EmployeeNumber: m.EmployeeNumber,
		ProfilePhotoID: m.ProfilePhotoID,
		LeaveTypeID:    m.LeaveTypeID,
		LeaveTypeName:  m.LeaveTypeName,
		StartDate:      m.StartDate,
		EndDate:        m.EndDate,
		Days:           m.Days,
		Reason:         m.Reason,
		AttachmentID:   m.AttachmentID,
		Status:         m.Status,
		ApprovedBy:     m.ApprovedBy,
		ApprovedAt:     m.ApprovedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
