package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"hrms/internal/leave/entity"
	"hrms/internal/leave/models"
	"hrms/internal/leave/repository"
	errors "hrms/internal/pkg/apperror"
	"hrms/internal/pkg/timeutil"
)

func (uc *LeaveUsecase) SubmitLeave(ctx context.Context, input models.CreateLeaveSubmissionInput) (*entity.LeaveSubmission, error) {
	employeeID, err := uc.employeeFetcher.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup employee: %w", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("no employee record linked to this user")
	}

	lt, err := uc.leaveTypeRepo.FindByID(ctx, input.LeaveTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find leave type: %w", err)
	}
	if lt == nil || !lt.IsActive {
		return nil, errors.NewInvalidInput("leave type not found or inactive")
	}

	loc := timeutil.LoadDefaultLocation()
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	if input.StartDate.Before(today) {
		return nil, errors.NewInvalidInput("start date cannot be in the past")
	}
	if input.EndDate.Before(input.StartDate) {
		return nil, errors.NewInvalidInput("end date must be after start date")
	}

	overlap, err := uc.submissionRepo.HasOverlap(ctx, employeeID, input.StartDate, input.EndDate, "")
	if err != nil {
		return nil, fmt.Errorf("failed to check overlap: %w", err)
	}
	if overlap {
		return nil, errors.NewInvalidInput("leave dates overlap with an existing submission")
	}

	days := entity.CountWeekdays(input.StartDate, input.EndDate)
	if days == 0 {
		return nil, errors.NewInvalidInput("selected dates contain no working days")
	}

	if !lt.IsUnlimited {
		year := input.StartDate.Year()
		balance, err := uc.leaveBalanceRepo.FindByEmployeeAndTypeYear(ctx, employeeID, input.LeaveTypeID, year)
		if err != nil {
			return nil, fmt.Errorf("failed to find balance: %w", err)
		}
		if balance == nil {
			return nil, errors.NewInvalidInput("leave balance not yet available for this year, please try again later")
		}
		if !balance.SufficientQuota(days) {
			return nil, errors.NewInvalidInput("insufficient leave balance")
		}
	}

	s := entity.NewLeaveSubmission(
		employeeID, input.LeaveTypeID,
		input.StartDate, input.EndDate, days,
		input.Reason, input.AttachmentID,
	)
	if err := uc.submissionRepo.Create(ctx, s); err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}
	return s, nil
}

func (uc *LeaveUsecase) ApproveSubmission(ctx context.Context, submissionID, approvedBy string) (*entity.LeaveSubmission, error) {
	s, err := uc.submissionRepo.FindByID(ctx, submissionID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find submission", err)
	}
	if s == nil {
		return nil, errors.NewNotFound("submission not found")
	}

	if err := s.Approve(approvedBy); err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	lt, err := uc.leaveTypeRepo.FindByID(ctx, s.LeaveTypeID)
	if err != nil {
		return nil, errors.WrapInternal("failed to find leave type", err)
	}
	if lt == nil {
		return nil, errors.NewNotFound("leave type not found")
	}

	tx, err := uc.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.WrapInternal("failed to begin transaction", err)
	}
	defer tx.Rollback()

	submissionRepo := uc.submissionRepo.WithTx(tx)
	balanceRepo := uc.leaveBalanceRepo.WithTx(tx)

	if err := submissionRepo.Update(ctx, s); err != nil {
		return nil, errors.WrapInternal("failed to update submission", err)
	}

	if lt.IsHalfDay {
		uc.ensureHalfDayPunches(ctx, s.EmployeeID, s.StartDate, s.EndDate, lt.Name)
	}

	if lt.IsUnlimited {
		uc.reprocessAttendance(ctx, s.EmployeeID, s.StartDate, s.EndDate)
		if err := tx.Commit(); err != nil {
			return nil, errors.WrapInternal("failed to commit transaction", err)
		}
		return s, nil
	}

	year := s.StartDate.Year()
	balance, err := balanceRepo.FindByEmployeeAndTypeYear(ctx, s.EmployeeID, s.LeaveTypeID, year)
	if err != nil {
		return nil, errors.WrapInternal("failed to find balance", err)
	}
	if balance == nil {
		return nil, errors.NewInvalidInput("no leave balance found for this year")
	}

	// Atomic consume — uses UPDATE with WHERE clause to prevent race conditions
	if err := balanceRepo.ConsumeBalance(ctx, balance.ID, s.Days); err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.WrapInternal("failed to commit transaction", err)
	}

	uc.reprocessAttendance(ctx, s.EmployeeID, s.StartDate, s.EndDate)
	return s, nil
}

func (uc *LeaveUsecase) RejectSubmission(ctx context.Context, submissionID string) (*entity.LeaveSubmission, error) {
	s, err := uc.submissionRepo.FindByID(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find submission: %w", err)
	}
	if s == nil {
		return nil, errors.NewNotFound("submission not found")
	}

	if err := s.Reject(); err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}
	if err := uc.submissionRepo.Update(ctx, s); err != nil {
		return nil, fmt.Errorf("failed to update submission: %w", err)
	}

	uc.reprocessAttendance(ctx, s.EmployeeID, s.StartDate, s.EndDate)
	return s, nil
}

func (uc *LeaveUsecase) CancelSubmission(ctx context.Context, submissionID, userID string) (*entity.LeaveSubmission, error) {
	s, err := uc.submissionRepo.FindByID(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find submission: %w", err)
	}
	if s == nil {
		return nil, errors.NewNotFound("submission not found")
	}

	employeeID, err := uc.employeeFetcher.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup employee: %w", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("no employee record linked to this user")
	}

	if s.EmployeeID != employeeID {
		return nil, errors.NewForbidden("you can only cancel your own submissions")
	}

	if s.Status == entity.LeaveStatusApproved {
		if s.EndDate.Before(time.Now().Truncate(24 * time.Hour)) {
			return nil, errors.NewInvalidInput("cannot cancel approved leave that has already passed")
		}
	}

	wasApproved := s.Status == entity.LeaveStatusApproved

	if err := s.Cancel(); err != nil {
		return nil, errors.NewInvalidInput(err.Error())
	}

	if wasApproved {
		lt, err := uc.leaveTypeRepo.FindByID(ctx, s.LeaveTypeID)
		if err != nil {
			return nil, fmt.Errorf("failed to find leave type: %w", err)
		}
		if lt != nil && !lt.IsUnlimited {
			year := s.StartDate.Year()
			balance, err := uc.leaveBalanceRepo.FindByEmployeeAndTypeYear(ctx, s.EmployeeID, s.LeaveTypeID, year)
			if err != nil {
				return nil, fmt.Errorf("failed to find balance: %w", err)
			}
			if balance != nil {
				balance.Restore(s.Days)
				if err := uc.leaveBalanceRepo.Update(ctx, balance); err != nil {
					return nil, fmt.Errorf("failed to update balance: %w", err)
				}
			}
		}
	}

	if err := uc.submissionRepo.Update(ctx, s); err != nil {
		return nil, fmt.Errorf("failed to update submission: %w", err)
	}

	uc.reprocessAttendance(ctx, s.EmployeeID, s.StartDate, s.EndDate)
	return s, nil
}

func (uc *LeaveUsecase) ListMySubmissions(ctx context.Context, input models.ListSubmissionInput) (*models.ListSubmissionResult, error) {

	employeeID, err := uc.employeeFetcher.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup employee: %w", err)
	}
	if employeeID == "" {
		return nil, errors.NewNotFound("no employee record linked to this user")
	}
	filter := repository.LeaveSubmissionFilter{
		EmployeeID: employeeID,
		Status:     input.Status,
		Page:       input.Page,
		PerPage:    input.PerPage,
	}

	rows, total, err := uc.submissionRepo.FindByEmployeeID(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}

	items := make([]models.LeaveSubmission, 0, len(rows))
	for _, r := range rows {
		items = append(items, submissionToModel(r))
	}
	uc.enrichLeaveTypeNames(ctx, items)
	return &models.ListSubmissionResult{Items: items, Total: total}, nil
}

func (uc *LeaveUsecase) ListAllSubmissions(ctx context.Context, filter models.ListAllSubmissionInput) (*models.ListSubmissionResult, error) {
	rows, total, err := uc.submissionRepo.FindAll(ctx, repository.SubmissionFilter{
		Status:    filter.Status,
		Search:    filter.Search,
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Page:      filter.Page,
		PerPage:   filter.PerPage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list all submissions: %w", err)
	}
	items := make([]models.LeaveSubmission, 0, len(rows))
	for _, r := range rows {
		items = append(items, *r)
	}
	return &models.ListSubmissionResult{Items: items, Total: total}, nil
}

func (uc *LeaveUsecase) GetSubmissionByID(ctx context.Context, id string) (*models.LeaveSubmission, error) {
	s, err := uc.submissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find submission: %w", err)
	}
	if s == nil {
		return nil, errors.NewNotFound("submission not found")
	}
	m := submissionToModel(s)
	uc.enrichLeaveTypeNames(ctx, []models.LeaveSubmission{m})
	return &m, nil
}

func (uc *LeaveUsecase) enrichLeaveTypeNames(ctx context.Context, items []models.LeaveSubmission) {
	types, err := uc.leaveTypeRepo.FindAllActive(ctx)
	if err != nil {
		return
	}
	nameByID := make(map[string]string, len(types))
	for _, t := range types {
		nameByID[t.ID] = t.Name
	}
	for i := range items {
		if n, ok := nameByID[items[i].LeaveTypeID]; ok {
			items[i].LeaveTypeName = n
		}
	}
}

func submissionToModel(s *entity.LeaveSubmission) models.LeaveSubmission {
	return models.LeaveSubmission{
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

func (uc *LeaveUsecase) ensureHalfDayPunches(ctx context.Context, employeeID string, startDate, endDate time.Time, leaveTypeName string) {
	if uc.halfDayPunchHandler == nil {
		return
	}
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		in, out, err := uc.halfDayPunchHandler.EnsureHalfDayPunches(ctx, employeeID, d)
		if err != nil {
			uc.log.WithFields(logrus.Fields{
				"employee_id": employeeID,
				"date":        d.Format("2006-01-02"),
				"error":       err,
			}).Warn("failed to create half-day punches")
			continue
		}
		if in || out {
			uc.log.WithFields(logrus.Fields{
				"employee_id":     employeeID,
				"date":            d.Format("2006-01-02"),
				"clock_in_auto":   in,
				"clock_out_auto":  out,
				"leave_type":      leaveTypeName,
			}).Info("auto-created half-day attendance punches")
		}
	}
}

func (uc *LeaveUsecase) reprocessAttendance(ctx context.Context, employeeID string, startDate, endDate time.Time) {
	if uc.attendanceProcessor == nil {
		return
	}
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		skipped, err := uc.attendanceProcessor.ReprocessDay(ctx, employeeID, d)
		if err != nil {
			uc.log.WithFields(logrus.Fields{
				"employee_id": employeeID,
				"date":        dateStr,
				"error":       err,
			}).Warn("failed to reprocess attendance after leave status change")
			continue
		}
		if skipped {
			uc.log.WithFields(logrus.Fields{
				"employee_id": employeeID,
				"date":        dateStr,
			}).Warn("leave status changed but date has manual correction — attendance not overridden. Delete the correction to apply leave changes.")
		}
	}
}
