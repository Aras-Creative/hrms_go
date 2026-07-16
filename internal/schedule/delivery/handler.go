package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
	scheduleAdapter "hrms/internal/schedule/adapter"
	"hrms/internal/schedule/entity"
	"hrms/internal/schedule/usecase"
)

type Notifier interface {
	Notify(ctx context.Context, userID, ntype, title, body, resource, resourceID string) error
}

const notifTypeEmployment = "employment"

type ScheduleHandler struct {
	wpUc        *usecase.WorkPatternUsecase
	epUc        *usecase.EmployeePatternUsecase
	ovUc        *usecase.ScheduleOverrideUsecase
	ovVUC       *usecase.ScheduleOverviewUsecase
	empFetcher  usecase.EmployeeFetcher
	auditLogger *scheduleAdapter.AuditLogger
	notifUC     Notifier
}

func NewScheduleHandler(
	wpUc *usecase.WorkPatternUsecase,
	epUc *usecase.EmployeePatternUsecase,
	ovUc *usecase.ScheduleOverrideUsecase,
	ovVUC *usecase.ScheduleOverviewUsecase,
	empFetcher usecase.EmployeeFetcher,
	auditLogger *scheduleAdapter.AuditLogger,
	notifUC Notifier,
) *ScheduleHandler {
	return &ScheduleHandler{wpUc: wpUc, epUc: epUc, ovUc: ovUc, ovVUC: ovVUC, empFetcher: empFetcher, auditLogger: auditLogger, notifUC: notifUC}
}

// ---- Work Patterns ----

func (h *ScheduleHandler) CreatePattern(c fiber.Ctx) error {
	var req CreateWorkingPatternRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}
	wp, err := h.wpUc.Create(c.RequestCtx(), toCreateInput(req))
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, "work_pattern", wp.ID, "", scheduleAdapter.ActionPatternCreate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name},
			)
		}
	}
	return response.Created(c, toResponse(wp))
}

func (h *ScheduleHandler) GetPattern(c fiber.Ctx) error {
	wp, err := h.wpUc.GetByID(c.RequestCtx(), c.Params("id"))
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, toResponse(wp))
}

func (h *ScheduleHandler) ListPatterns(c fiber.Ctx) error {
	list, err := h.wpUc.GetAll(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}
	resp := make([]WorkPatternResponse, 0, len(list))
	for _, wp := range list {
		resp = append(resp, toResponse(wp))
	}
	return response.OK(c, resp)
}

func (h *ScheduleHandler) UpdatePattern(c fiber.Ctx) error {
	var req UpdateWorkingPatternRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}
	wp, err := h.wpUc.Update(c.RequestCtx(), c.Params("id"), toUpdateInput(req))
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, "work_pattern", wp.ID, "", scheduleAdapter.ActionPatternUpdate,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"name": req.Name},
			)
		}
	}
	return response.OK(c, toResponse(wp))
}

func (h *ScheduleHandler) DeletePattern(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.wpUc.Delete(c.RequestCtx(), id); err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, "work_pattern", id, "", scheduleAdapter.ActionPatternDelete,
				c.IP(), string(c.RequestCtx().UserAgent()),
				nil,
			)
		}
	}
	return response.NoContent(c)
}

func (h *ScheduleHandler) ListPatternOptions(c fiber.Ctx) error {
	opts, err := h.wpUc.GetOptions(c.RequestCtx())
	if err != nil {
		return response.Error(c, err)
	}
	return response.Options(c, opts)
}

func toCreateInput(req CreateWorkingPatternRequest) usecase.CreateWorkingPatternInput {
	details := make([]usecase.WorkingPatternDetailInput, 0, len(req.Details))
	for _, d := range req.Details {
		details = append(details, usecase.WorkingPatternDetailInput{DayOfWeek: d.DayOfWeek, Type: d.Type, StartTime: d.StartTime, EndTime: d.EndTime})
	}
	return usecase.CreateWorkingPatternInput{Name: req.Name, Description: req.Description, Details: details}
}

func toUpdateInput(req UpdateWorkingPatternRequest) usecase.UpdateWorkingPatternInput {
	details := make([]usecase.WorkingPatternDetailInput, 0, len(req.Details))
	for _, d := range req.Details {
		details = append(details, usecase.WorkingPatternDetailInput{DayOfWeek: d.DayOfWeek, Type: d.Type, StartTime: d.StartTime, EndTime: d.EndTime})
	}
	return usecase.UpdateWorkingPatternInput{Name: req.Name, Description: req.Description, Details: details}
}

func toResponse(wp *entity.WorkingPattern) WorkPatternResponse {
	details := make([]WorkingPatternDetailResponse, 0, len(wp.Details))
	for _, d := range wp.Details {
		details = append(details, WorkingPatternDetailResponse{DayOfWeek: int(d.DayOfWeek), Type: string(d.Type), StartTime: d.StartTime, EndTime: d.EndTime})
	}
	return WorkPatternResponse{
		ID: wp.ID, Name: wp.Name, Description: wp.Description,
		IsActive: wp.IsActive, Details: details,
		CreatedAt: wp.CreatedAt, UpdatedAt: wp.UpdatedAt,
	}
}

// ---- Assign ----

func (h *ScheduleHandler) Assign(c fiber.Ctx) error {
	var req AssignPatternRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}
	validTo, err := parseOptionalDate(req.ValidTo)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.epUc.Assign(c.RequestCtx(), usecase.AssignPatternInput{
		EmployeeIDs: req.EmployeeIDs, WorkPatternID: req.WorkPatternID,
		ValidFrom: req.ValidFrom, ValidTo: validTo,
	})
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid != "" {
			h.auditLogger.Log(c.RequestCtx(), uid, "employee_work_pattern", "", "", scheduleAdapter.ActionAssign,
				c.IP(), string(c.RequestCtx().UserAgent()),
				map[string]any{"employee_ids": req.EmployeeIDs, "work_pattern_id": req.WorkPatternID},
			)
		}
	}
	return response.OK(c, result)
}

func (h *ScheduleHandler) GetActivePattern(c fiber.Ctx) error {
	ewp, err := h.epUc.GetActive(c.RequestCtx(), c.Params("employeeId"))
	if err != nil {
		return response.Error(c, err)
	}
	if ewp == nil {
		return response.Error(c, errors.NewNotFound("no active pattern for this employee"))
	}
	return response.OK(c, ewpToResponse(ewp))
}

func (h *ScheduleHandler) GetPatternHistory(c fiber.Ctx) error {
	list, err := h.epUc.GetHistory(c.RequestCtx(), c.Params("employeeId"))
	if err != nil {
		return response.Error(c, err)
	}
	resp := make([]EmployeeWorkPatternResponse, 0, len(list))
	for _, ewp := range list {
		resp = append(resp, ewpToResponse(ewp))
	}
	return response.OK(c, resp)
}

func ewpToResponse(ewp *entity.EmployeeWorkPattern) EmployeeWorkPatternResponse {
	var validTo *string
	if ewp.ValidTo != nil {
		s := ewp.ValidTo.Format("2006-01-02")
		validTo = &s
	}
	return EmployeeWorkPatternResponse{
		ID: ewp.ID, EmployeeID: ewp.EmployeeID, WorkPatternID: ewp.WorkPatternID,
		ValidFrom: ewp.ValidFrom.Format("2006-01-02"), ValidTo: validTo,
		IsActive: ewp.IsActive, CreatedAt: ewp.CreatedAt, UpdatedAt: ewp.UpdatedAt,
	}
}

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, errors.NewInvalidInput("invalid date, expected yyyy-mm-dd")
	}
	return &t, nil
}

// ---- Overrides ----

func (h *ScheduleHandler) SetOverride(c fiber.Ctx) error {
	var req SetOverrideRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body: "+err.Error()))
	}
	dateFrom, err := parseDate(req.DateFrom)
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid date_from: "+err.Error()))
	}
	dateTo, err := parseDate(req.DateTo)
	if err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid date_to: "+err.Error()))
	}
	results, err := h.ovUc.SetRange(c.RequestCtx(), usecase.SetOverrideInputRange{
		EmployeeID: req.EmployeeID, DateFrom: dateFrom, DateTo: dateTo,
		IsWorkingDay: req.IsWorkingDay, StartTime: req.StartTime,
		EndTime: req.EndTime, Reason: req.Reason,
	})
	if err != nil {
		return response.Error(c, err)
	}
	resp := make([]OverrideResponse, len(results))
	for i, o := range results {
		resp[i] = overrideToResponse(o)
	}

	if h.notifUC != nil {
		empUserID, _ := h.empFetcher.FindUserIDByEmployeeID(c.RequestCtx(), req.EmployeeID)
		if empUserID != "" {
			h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeEmployment,
				"Jadwal Diubah",
				fmt.Sprintf("Jadwal Anda pada %s hingga %s telah diubah", req.DateFrom, req.DateTo),
				"schedule_override", req.EmployeeID,
			)
		}
	}

	return response.OK(c, resp)
}

func (h *ScheduleHandler) GetOverride(c fiber.Ctx) error {
	o, err := h.ovUc.GetByID(c.RequestCtx(), c.Params("id"))
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, overrideToResponse(o))
}

func (h *ScheduleHandler) ListOverrides(c fiber.Ctx) error {
	employeeID := c.Query("employee_id")
	from, to := c.Query("from"), c.Query("to")
	var list []*entity.EmployeeScheduleOverride
	var err error
	if employeeID != "" {
		list, err = h.ovUc.ListByEmployee(c.RequestCtx(), employeeID, from, to)
	} else {
		list, err = h.ovUc.ListAll(c.RequestCtx(), from, to)
	}
	if err != nil {
		return response.Error(c, err)
	}
	resp := make([]OverrideResponse, 0, len(list))
	for _, o := range list {
		resp = append(resp, overrideToResponse(o))
	}
	return response.OK(c, resp)
}

func (h *ScheduleHandler) DeleteOverride(c fiber.Ctx) error {
	id := c.Params("id")

	o, _ := h.ovUc.GetByID(c.RequestCtx(), id)

	if err := h.ovUc.Remove(c.RequestCtx(), id); err != nil {
		return response.Error(c, err)
	}

	if h.notifUC != nil && o != nil {
		empUserID, _ := h.empFetcher.FindUserIDByEmployeeID(c.RequestCtx(), o.EmployeeID)
		if empUserID != "" {
			h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeEmployment,
				"Jadwal Dihapus",
				fmt.Sprintf("Jadwal Anda pada %s telah dihapus", o.Date.Format("02 Jan 2006")),
				"schedule_override", o.EmployeeID,
			)
		}
	}

	return response.NoContent(c)
}

func (h *ScheduleHandler) Overview(c fiber.Ctx) error {
	list, err := h.ovVUC.Query(c.RequestCtx(), usecase.ScheduleOverviewParams{
		EmployeeID: c.Query("employee_id"), DesignationID: c.Query("designation_id"),
		Search: c.Query("search"), From: c.Query("from"), To: c.Query("to"),
	})
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, list)
}

func parseDate(s string) (time.Time, error) {
	layouts := []string{"2006-01-02", time.RFC3339}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.NewInvalidInput("unable to parse date, expected format: yyyy-mm-dd")
}

// ---- My Schedule ----

func (h *ScheduleHandler) MyToday(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.NewUnauthorized("user not authenticated"))
	}

	employeeID, err := h.empFetcher.FindByUserID(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, errors.NewInternal("failed to resolve employee"))
	}
	if employeeID == "" {
		return response.Error(c, errors.NewNotFound("employee not found for user"))
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Get active pattern assignment
	ewp, err := h.epUc.GetActive(c.RequestCtx(), employeeID)
	if err != nil {
		return response.Error(c, errors.NewInternal("failed to get active pattern"))
	}

	var wp *entity.WorkingPattern
	if ewp != nil {
		wp, err = h.wpUc.GetByID(c.RequestCtx(), ewp.WorkPatternID)
		if err != nil {
			return response.Error(c, errors.NewInternal("failed to get work pattern"))
		}
	}

	var detail *entity.WorkingPatternDetail
	dow := int(today.Weekday())
	if wp != nil {
		for _, d := range wp.Details {
			if int(d.DayOfWeek) == dow {
				detail = &d
				break
			}
		}
	}

	// Get override for today
	overrides, err := h.ovUc.ListByEmployee(c.RequestCtx(), employeeID, today.Format("2006-01-02"), today.Format("2006-01-02"))
	if err != nil {
		return response.Error(c, errors.NewInternal("failed to get overrides"))
	}
	var override *entity.EmployeeScheduleOverride
	if len(overrides) > 0 {
		override = overrides[0]
	}

	// Resolve final schedule
	isWorkingDay := detail != nil && detail.Type != entity.WorkingTypeOff && (detail.Type == entity.WorkingTypeDynamic || (detail.StartTime != nil && detail.EndTime != nil))
	var startTime, endTime *string
	var source string
	var workingType *string
	var overrideReason *string

	if override != nil {
		source = "override"
		isWorkingDay = override.IsWorkingDay
		startTime = detail.StartTime
		endTime = detail.EndTime
		if override.StartTime != nil {
			startTime = override.StartTime
		}
		if override.EndTime != nil {
			endTime = override.EndTime
		}
		overrideReason = override.Reason
		if override.StartTime != nil && *override.StartTime != "" {
			s := "fixed"
			workingType = &s
		} else if detail != nil {
			s := string(detail.Type)
			workingType = &s
		}
	} else if detail != nil {
		source = "working_pattern"
		startTime = detail.StartTime
		endTime = detail.EndTime
		s := string(detail.Type)
		workingType = &s
	} else if ewp != nil {
		source = "no_pattern"
	} else {
		source = "none"
	}

	resp := fiber.Map{
		"date":            today.Format("2006-01-02"),
		"day_of_week":     dow,
		"is_working_day":  isWorkingDay,
		"source":          source,
		"working_type":    workingType,
		"start_time":      startTime,
		"end_time":        endTime,
		"override_reason": overrideReason,
	}

	return response.OK(c, resp)
}

func overrideToResponse(o *entity.EmployeeScheduleOverride) OverrideResponse {
	return OverrideResponse{
		ID: o.ID, EmployeeID: o.EmployeeID, Date: o.Date.Format("2006-01-02"),
		IsWorkingDay: o.IsWorkingDay, StartTime: o.StartTime, EndTime: o.EndTime,
		Reason: o.Reason, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
}
