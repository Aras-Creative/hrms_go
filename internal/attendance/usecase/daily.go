package usecase

import (
	"context"
	"fmt"
	"time"

	"hrms/internal/attendance/entity"
	"hrms/internal/attendance/models"
	"hrms/internal/attendance/repository"
	"hrms/internal/pkg/cache"
	"hrms/internal/pkg/timeutil"
	errors "hrms/internal/pkg/apperror"
)

type DailyAttendanceUsecase struct {
	dailyRepo      repository.DailyAttendanceRepository
	correctionRepo repository.CorrectionRepository
	punchRepo      repository.PunchRepository
	processor      *DailyProcessor
	recapCache     *cache.Cache[*RecapResult]
}

func NewDailyAttendanceUsecase(dailyRepo repository.DailyAttendanceRepository, correctionRepo repository.CorrectionRepository, punchRepo repository.PunchRepository, processor *DailyProcessor) *DailyAttendanceUsecase {
	return &DailyAttendanceUsecase{
		dailyRepo:      dailyRepo,
		correctionRepo: correctionRepo,
		punchRepo:      punchRepo,
		processor:      processor,
		recapCache:     cache.New[*RecapResult](2 * time.Minute),
	}
}

func (uc *DailyAttendanceUsecase) List(ctx context.Context, input models.ListInput) (*models.ListResult, error) {
	fromStr := input.From
	toStr := input.To
	if fromStr == "" {
		fromStr = time.Now().Format("2006-01-02")
	}
	if toStr == "" {
		toStr = fromStr
	}
	from, to, err := timeutil.ParseDateRange(fromStr, toStr)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	rows, total, err := uc.dailyRepo.FindAllPaginated(ctx,
		input.SearchName, input.Status, input.DesignationID,
		input.IsLate, input.IsEarlyLeave,
		from, to, input.Page, input.PerPage,
	)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to list attendance: %v", err))
	}

	items := make([]*models.AdminAttendanceItem, len(rows))
	for i, row := range rows {
		items[i] = adminRowToItem(row)
	}
	return &models.ListResult{Items: items, Total: total}, nil
}

func adminRowToItem(r *repository.AdminAttendanceRow) *models.AdminAttendanceItem {
	if r == nil {
		return nil
	}
	return &models.AdminAttendanceItem{
		ID:                r.ID,
		EmployeeID:        r.EmployeeID,
		EmployeeName:      r.EmployeeName,
		EmployeeNumber:    r.EmployeeNumber,
		DesignationName:   r.DesignationName,
		ProfilePhotoID:    r.ProfilePhotoID,
		Date:              r.Date,
		Status:            r.Status,
		IsLate:            r.IsLate,
		IsEarlyLeave:      r.IsEarlyLeave,
		ExpectedStartTime: r.ExpectedStartTime,
		ExpectedEndTime:   r.ExpectedEndTime,
		Source:            r.Source,
		FirstPunchIn:      r.FirstPunchIn,
		LastPunchOut:      r.LastPunchOut,
		TotalWorkSeconds:  r.TotalWorkSeconds,
		LeaveSubmissionID: r.LeaveSubmissionID,
		LeaveTypeName:     r.LeaveTypeName,
		ScheduleOverrideID: r.ScheduleOverrideID,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

func correctionRowToItem(r *repository.CorrectionViewRow) *models.CorrectionViewItem {
	if r == nil {
		return nil
	}
	return &models.CorrectionViewItem{
		ID:           r.ID,
		EmployeeID:   r.EmployeeID,
		EmployeeName: r.EmployeeName,
		Date:         r.Date,
		ClockIn:      r.ClockIn,
		ClockOut:     r.ClockOut,
		Status:       r.Status,
		Reason:       r.Reason,
		CorrectedBy:  r.CorrectedBy,
		CreatedAt:    r.CreatedAt,
	}
}

func (uc *DailyAttendanceUsecase) Query(ctx context.Context, input models.DailyQueryInput) ([]*entity.DailyAttendance, error) {
	from, to, err := timeutil.ParseDateRange(input.From, input.To)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}
	return uc.processor.ProcessDailyRange(ctx, input.EmployeeID, from, to)
}

type AttendanceDetail struct {
	Attendance  *entity.DailyAttendance       `json:"attendance"`
	Corrections []*entity.AttendanceCorrection `json:"corrections"`
	Punches     []*entity.Punch                `json:"punches"`
}

func (uc *DailyAttendanceUsecase) GetDetail(ctx context.Context, id string) (*AttendanceDetail, error) {
	da, err := uc.dailyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to find daily attendance: %v", err))
	}
	if da == nil {
		return nil, errors.NewNotFound("daily attendance not found")
	}

	correction, err := uc.correctionRepo.FindByEmployeeAndDate(ctx, da.EmployeeID, da.Date)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to find corrections: %v", err))
	}
	var corrections []*entity.AttendanceCorrection
	if correction != nil {
		corrections = append(corrections, correction)
	}

	startOfDay := time.Date(da.Date.Year(), da.Date.Month(), da.Date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Second)
	punches, err := uc.punchRepo.FindByEmployeeAndDateRange(ctx, da.EmployeeID, startOfDay, endOfDay)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to find punches: %v", err))
	}

	return &AttendanceDetail{
		Attendance:  da,
		Corrections: corrections,
		Punches:     punches,
	}, nil
}

func (uc *DailyAttendanceUsecase) GetAttendanceHistoryByEmployeeID(ctx context.Context, employeeID, fromStr, toStr string) ([]models.MyAttendanceHistoryItem, error) {
	from, to, err := timeutil.ParseDateRange(fromStr, toStr)
	if err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	records, err := uc.processor.ComputeDailyRange(ctx, employeeID, from, to)
	if err != nil {
		return nil, err
	}

	items := make([]models.MyAttendanceHistoryItem, len(records))
	for i, da := range records {
		items[i] = toHistoryItem(da)
	}
	return items, nil
}
