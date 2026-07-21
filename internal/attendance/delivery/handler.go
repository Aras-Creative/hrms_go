package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/sse"

	"hrms/internal/attendance/adapter"
	"hrms/internal/attendance/models"
	"hrms/internal/attendance/usecase"
	auditUc "hrms/internal/audit/usecase"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
	sselib "hrms/internal/pkg/sse"
)

type PhotoURLResolver interface {
	ResolveURL(ctx context.Context, documentID string) (string, error)
	ResolveURLs(ctx context.Context, documentIDs []string) (map[string]string, error)
}

type Notifier interface {
	Notify(ctx context.Context, userID, ntype, title, body, resource, resourceID string) error
}

const (
	notifTypeEmployment = "employment"
	notifTypeLeave      = "leave"
)

type AttendanceHandler struct {
	punchUc       *usecase.PunchUsecase
	dailyUc       *usecase.DailyAttendanceUsecase
	correctionUc  *usecase.CorrectionUsecase
	meUc          *usecase.MeUsecase
	fetcher       usecase.EmployeeFetcher
	hub           *sselib.Hub
	auditUC       *auditUc.AuditUsecase
	photoResolver PhotoURLResolver
	auditLogger   *adapter.AuditLogger
	notifUC       Notifier
}

func NewAttendanceHandler(
	punchUc *usecase.PunchUsecase,
	dailyUc *usecase.DailyAttendanceUsecase,
	correctionUc *usecase.CorrectionUsecase,
	meUc *usecase.MeUsecase,
	fetcher usecase.EmployeeFetcher,
	hub *sselib.Hub,
	auditUC *auditUc.AuditUsecase,
	photoResolver PhotoURLResolver,
	auditLogger *adapter.AuditLogger,
	notifUC Notifier,
) *AttendanceHandler {
	return &AttendanceHandler{punchUc: punchUc, dailyUc: dailyUc, correctionUc: correctionUc, meUc: meUc, fetcher: fetcher, hub: hub, auditUC: auditUC, photoResolver: photoResolver, auditLogger: auditLogger, notifUC: notifUC}
}

// ---- Punch ----

func getUserID(c fiber.Ctx) (string, error) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return "", errors.NewUnauthorized("user not authenticated")
	}
	return userID, nil
}

func (h *AttendanceHandler) PunchIn(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}
	employeeID, _, err := h.fetcher.FindByUserID(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, errors.NewInternal("failed to resolve employee"))
	}
	if employeeID == "" {
		return response.Error(c, errors.NewNotFound("employee not found for user"))
	}
	p, err := h.punchUc.PunchIn(c.RequestCtx(), models.PunchInput{EmployeeID: employeeID})
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		h.auditLogger.Log(c.RequestCtx(), userID, employeeID, "", adapter.ActionPunchIn,
			c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"employee_id": employeeID, "timestamp": p.Timestamp},
		)
	}
	return response.Created(c, punchToResponse(p))
}

func (h *AttendanceHandler) PunchOut(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}
	employeeID, _, err := h.fetcher.FindByUserID(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, errors.NewInternal("failed to resolve employee"))
	}
	if employeeID == "" {
		return response.Error(c, errors.NewNotFound("employee not found for user"))
	}
	p, err := h.punchUc.PunchOut(c.RequestCtx(), models.PunchInput{EmployeeID: employeeID})
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		h.auditLogger.Log(c.RequestCtx(), userID, employeeID, "", adapter.ActionPunchOut,
			c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"employee_id": employeeID, "timestamp": p.Timestamp},
		)
	}
	return response.OK(c, punchToResponse(p))
}

func (h *AttendanceHandler) PunchToday(c fiber.Ctx) error {
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		return response.Error(c, errors.NewInvalidInput("employee_id is required"))
	}
	list, err := h.punchUc.GetToday(c.RequestCtx(), employeeID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, punchListToResponse(list))
}

func (h *AttendanceHandler) PunchHistory(c fiber.Ctx) error {
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		return response.Error(c, errors.NewInvalidInput("employee_id is required"))
	}
	list, err := h.punchUc.GetHistory(c.RequestCtx(), models.PunchHistoryInput{
		EmployeeID: employeeID, From: c.Query("from"), To: c.Query("to"),
	})
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, punchListToResponse(list))
}

// ---- Daily ----

func (h *AttendanceHandler) GetDetail(c fiber.Ctx) error {
	id, err := response.ParseParamID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.dailyUc.GetDetail(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}

	corrs := make([]CorrectionResponse, 0, len(result.Corrections))
	for _, c := range result.Corrections {
		corrs = append(corrs, correctionToResponse(c))
	}

	var auditLogs []AuditLogEntryResponse
	if h.auditUC != nil && result.Attendance != nil {
		logs, err := h.auditUC.ListByResourceWithActor(c.RequestCtx(), "employee", result.Attendance.EmployeeID)
		if err == nil {
			for _, a := range logs {
				auditLogs = append(auditLogs, AuditLogEntryResponse{
					ID: a.ID, Action: a.Action, ActorID: a.ActorID,
					ActorName: a.ActorName, Payload: a.Payload,
					IPAddress: a.IPAddress, UserAgent: a.UserAgent,
					CreatedAt: a.CreatedAt,
				})
			}
		}
	}

	return response.OK(c, AttendanceDetailResponse{
		Attendance:  dailyToResponse(result.Attendance),
		Corrections: corrs,
		Punches:     punchListToResponse(result.Punches),
		AuditLogs:   auditLogs,
	})
}

func (h *AttendanceHandler) DailyList(c fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(c.Query("per_page", "20"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 20
	}
	result, err := h.dailyUc.List(c.RequestCtx(), models.ListInput{
		SearchName: c.Query("search"), Status: c.Query("status"),
		DesignationID: c.Query("designation_id"),
		IsLate:        c.Query("is_late"), IsEarlyLeave: c.Query("is_early_leave"),
		From: c.Query("from"), To: c.Query("to"),
		Page: page, PerPage: perPage,
	})
	if err != nil {
		return response.Error(c, err)
	}
	items := make([]AdminAttendanceResponse, 0, len(result.Items))
	for _, row := range result.Items {
		items = append(items, adminAttendanceToResponse(row))
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

// ---- Correction ----

func (h *AttendanceHandler) CorrectionCreate(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}

	// Peek at raw JSON to detect explicit nulls (Go's decoder can't distinguish
	// null from absent for *time.Time).
	rawBody := c.Body()
	var rawMap map[string]any
	hasClockIn := false
	hasClockOut := false
	if err := json.Unmarshal(rawBody, &rawMap); err == nil {
		_, hasClockIn = rawMap["clock_in"]
		_, hasClockOut = rawMap["clock_out"]
	}

	var req CreateCorrectionRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, errors.NewInvalidInput("invalid request body"))
	}
	result, created, err := h.correctionUc.Create(c.RequestCtx(), usecase.CreateCorrectionInput{
		EmployeeID: req.EmployeeID, Date: req.Date, ClockIn: req.ClockIn,
		ClockOut: req.ClockOut, HasClockIn: hasClockIn, HasClockOut: hasClockOut,
		Status: req.Status, IsLate: req.IsLate,
		IsEarlyLeave: req.IsEarlyLeave, Reason: req.Reason, CorrectedBy: userID,
	})
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		action := adapter.ActionCorrection
		if !created {
			action = adapter.ActionCorrectionUpdate
		}
		h.auditLogger.Log(c.RequestCtx(), userID, req.EmployeeID, result.ID, action,
			c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{
				"employee_id": req.EmployeeID,
				"date":        req.Date,
				"reason":      req.Reason,
			},
		)
	}

	if h.notifUC != nil {
		empUserID, _ := h.fetcher.FindUserIDByEmployeeID(c.RequestCtx(), req.EmployeeID)
		if empUserID != "" && empUserID != userID {
			title := "Koreksi Absensi"
			body := fmt.Sprintf("Absensi Anda pada %s telah dikoreksi", req.Date)
			if !created {
				body = fmt.Sprintf("Koreksi absensi Anda pada %s telah diperbarui", req.Date)
			}
			h.notifUC.Notify(c.RequestCtx(), empUserID, notifTypeEmployment,
				title, body,
				"attendance_correction", result.ID,
			)
		}
	}

	resp := correctionToResponse(result)
	if created {
		return response.Created(c, resp)
	}
	return response.OK(c, resp)
}

func (h *AttendanceHandler) CorrectionList(c fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(c.Query("per_page", "20"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 20
	}
	var startDate, endDate *time.Time
	if s := c.Query("start_date"); s != "" {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid start_date"))
		}
		startDate = &d
	}
	if s := c.Query("end_date"); s != "" {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return response.Error(c, errors.NewInvalidInput("invalid end_date"))
		}
		endDate = &d
	}
	result, err := h.correctionUc.List(c.RequestCtx(), usecase.ListCorrectionsInput{
		SearchName: c.Query("search"), StartDate: startDate, EndDate: endDate, Page: page, PerPage: perPage,
	})
	if err != nil {
		return response.Error(c, err)
	}
	return response.Paginate(c, correctionViewListToResponse(result.Items), page, perPage, result.Total)
}

func (h *AttendanceHandler) CorrectionDelete(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}
	id, err := response.ParseParamID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	corr, err := h.correctionUc.Delete(c.RequestCtx(), id)
	if err != nil {
		return response.Error(c, err)
	}
	if h.auditLogger != nil {
		h.auditLogger.Log(c.RequestCtx(), userID, corr.EmployeeID, id, adapter.ActionCorrectionDelete,
			c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"correction_id": id, "employee_id": corr.EmployeeID},
		)
	}
	return response.NoContent(c)
}

// ---- My Attendance ----

func (h *AttendanceHandler) MyAttendanceHistory(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		return response.Error(c, errors.NewInvalidInput("from and to query params are required"))
	}
	result, err := h.meUc.GetMyAttendanceHistory(c.RequestCtx(), userID, from, to)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, result)
}

func (h *AttendanceHandler) EmployeeAttendanceHistory(c fiber.Ctx) error {
	employeeID, err := response.ParseParamID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	// Security: only allow admin or requesting own data
	userID, err := getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}
	role, _ := c.Locals("role").(string)
	if role != "admin" {
		myEmpID, _, err := h.fetcher.FindByUserID(c.RequestCtx(), userID)
		if err != nil || myEmpID != employeeID {
			return response.Error(c, errors.NewForbidden("access denied"))
		}
	}
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		return response.Error(c, errors.NewInvalidInput("from and to query params are required"))
	}
	result, err := h.dailyUc.GetAttendanceHistoryByEmployeeID(c.RequestCtx(), employeeID, from, to)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, result)
}

func (h *AttendanceHandler) MyAttendanceStats(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}
	stats, err := h.meUc.GetMyStats(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, stats)
}

func (h *AttendanceHandler) MyAttendance(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Error(c, err)
	}
	result, err := h.meUc.GetMyAttendance(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, result)
}

func (h *AttendanceHandler) Recap(c fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		return response.Error(c, errors.NewInvalidInput("from and to query params are required"))
	}
	designationID := c.Query("designation_id")

	result, err := h.dailyUc.Recap(c.RequestCtx(), from, to, designationID)
	if err != nil {
		return response.Error(c, err)
	}

	output := h.dailyUc.BuildRecapOutput(result)

	// Resolve profile photo URLs
	photoIDs := make([]string, 0, len(output.Employees))
	for _, emp := range output.Employees {
		if emp.ProfilePhotoID != nil && *emp.ProfilePhotoID != "" {
			photoIDs = append(photoIDs, *emp.ProfilePhotoID)
		}
	}
	photoURLs := make(map[string]string)
	if len(photoIDs) > 0 && h.photoResolver != nil {
		urls, err := h.photoResolver.ResolveURLs(c.RequestCtx(), photoIDs)
		if err == nil {
			photoURLs = urls
		}
	}

	headers := make([]RecapHeader, len(output.Headers))
	for i, h := range output.Headers {
		headers[i] = RecapHeader{Key: h.Key, Label: h.Label}
	}

	employees := make([]RecapEmployee, 0, len(output.Employees))
	for _, emp := range output.Employees {
		var photoURL *string
		if emp.ProfilePhotoID != nil {
			if url, ok := photoURLs[*emp.ProfilePhotoID]; ok {
				photoURL = &url
			}
		}

		metrics := make([]RecapMetric, 0, len(headers))
		for _, h := range headers {
			metrics = append(metrics, RecapMetric{Key: h.Key, Value: emp.MetricValues[h.Key]})
		}

		employees = append(employees, RecapEmployee{
			ID:                      emp.ID,
			EmployeeNumber:          emp.EmployeeNumber,
			Name:                    emp.Name,
			ProfilePictureURL:       photoURL,
			Department:              emp.Department,
			WorkingDays:             emp.WorkingDays,
			LateMinutes:             emp.LateMinutes,
			AttendanceMetrics:       metrics,
			TotalAttendanceIncident: emp.TotalAttendanceIncident,
		})
	}

	return response.OK(c, RecapResponse{
		Headers:   headers,
		Employees: employees,
	})
}

// ---- SSE ----

func (h *AttendanceHandler) PunchEvents(c fiber.Ctx) error {
	return sse.New(sse.Config{
		Handler: func(c fiber.Ctx, stream *sse.Stream) error {
			events, err := h.hub.Subscribe(stream.Context(), "punches")
			if err != nil {
				return err
			}
			for {
				select {
				case msg, ok := <-events:
					if !ok {
						return nil
					}
					var data interface{}
					if err := json.Unmarshal([]byte(msg), &data); err == nil {
						if err := stream.Event(sse.Event{Name: "punch", Data: data}); err != nil {
							return err
						}
					}
				case <-stream.Done():
					return nil
				}
			}
		},
	})(c)
}
