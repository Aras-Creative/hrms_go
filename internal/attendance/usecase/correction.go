package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"

	"hrms/internal/attendance/entity"
	"hrms/internal/attendance/models"
	"hrms/internal/attendance/repository"
	errors "hrms/internal/pkg/apperror"
)

type CorrectionUsecase struct {
	db            *sqlx.DB
	correctionRepo repository.CorrectionRepository
	dailyRepo      repository.DailyAttendanceRepository
	empChecker     EmployeeExistenceChecker
	resolver       ScheduleResolver
}

func NewCorrectionUsecase(
	db *sqlx.DB,
	correctionRepo repository.CorrectionRepository,
	dailyRepo repository.DailyAttendanceRepository,
	empChecker EmployeeExistenceChecker,
	resolver ScheduleResolver,
) *CorrectionUsecase {
	return &CorrectionUsecase{
		db:             db,
		correctionRepo: correctionRepo,
		dailyRepo:      dailyRepo,
		empChecker:     empChecker,
		resolver:       resolver,
	}
}

type CreateCorrectionInput struct {
	EmployeeID   string
	Date         string
	ClockIn      *time.Time
	ClockOut     *time.Time
	Status       *string
	IsLate       *bool
	IsEarlyLeave *bool
	Reason       string
	CorrectedBy  string
}

type ListCorrectionsInput struct {
	SearchName string
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PerPage    int
}

type ListCorrectionsResult struct {
	Items []*models.CorrectionViewItem
	Total int64
}

func (uc *CorrectionUsecase) Create(ctx context.Context, input CreateCorrectionInput) (*entity.AttendanceCorrection, bool, error) {
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, false, errors.NewInvalidInput("invalid date format, expected YYYY-MM-DD")
	}

	correction := entity.NewAttendanceCorrection(
		input.EmployeeID, date, input.ClockIn, input.ClockOut, input.Status,
		input.IsLate, input.IsEarlyLeave,
		input.Reason, input.CorrectedBy,
	)

	if err := correction.Validate(); err != nil {
		return nil, false, errors.NewInvalidInput(err.Error())
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if date.After(today) {
		return nil, false, errors.NewInvalidInput("correction date cannot be in the future")
	}

	exists, err := uc.empChecker.ExistsByID(ctx, input.EmployeeID)
	if err != nil {
		return nil, false, errors.WrapInternal("failed to check employee", err)
	}
	if !exists {
		return nil, false, errors.NewNotFound("employee not found")
	}

	existing, err := uc.correctionRepo.FindByEmployeeAndDate(ctx, input.EmployeeID, date)
	if err != nil {
		return nil, false, errors.WrapInternal("failed to check existing correction", err)
	}

	tx, err := uc.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, false, errors.WrapInternal("failed to begin transaction", err)
	}
	defer tx.Rollback()

	dailyRepo := uc.dailyRepo.WithTx(tx)
	correctionRepo := uc.correctionRepo.WithTx(tx)
	processor := NewDailyProcessor(dailyRepo, uc.resolver)

	da, err := processor.ComputeDaily(ctx, input.EmployeeID, date)
	if err != nil {
		return nil, false, errors.WrapInternal("failed to compute base attendance", err)
	}

	created := existing == nil

	if created {
		if err := correctionRepo.Create(ctx, correction); err != nil {
			return nil, false, errors.WrapInternal("failed to save correction", err)
		}
	} else {
		existing.ClockIn = input.ClockIn
		existing.ClockOut = input.ClockOut
		existing.Status = input.Status
		existing.IsLate = input.IsLate
		existing.IsEarlyLeave = input.IsEarlyLeave
		existing.Reason = input.Reason
		existing.CorrectedBy = input.CorrectedBy
		if err := correctionRepo.Update(ctx, existing); err != nil {
			return nil, false, errors.WrapInternal("failed to update correction", err)
		}
		correction = existing
	}

	correction.ApplyTo(da)

	if err := dailyRepo.Upsert(ctx, da); err != nil {
		slog.Error("correction daily upsert failed",
			"employee_id", input.EmployeeID,
			"date", input.Date,
			"error", err,
		)
		return nil, false, errors.WrapInternal("failed to update daily attendance", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, false, errors.WrapInternal("failed to commit transaction", err)
	}

	return correction, created, nil
}

func (uc *CorrectionUsecase) List(ctx context.Context, input ListCorrectionsInput) (*ListCorrectionsResult, error) {
	var startDate, endDate time.Time
	if input.StartDate != nil {
		startDate = *input.StartDate
	} else {
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if input.EndDate != nil {
		endDate = *input.EndDate
	} else {
		endDate = time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	}
	rows, total, err := uc.correctionRepo.FindAllPaginated(ctx, input.SearchName, startDate, endDate, input.Page, input.PerPage)
	if err != nil {
		return nil, errors.WrapInternal("failed to list corrections", err)
	}
	items := make([]*models.CorrectionViewItem, len(rows))
	for i, r := range rows {
		items[i] = correctionRowToItem(r)
	}
	return &ListCorrectionsResult{Items: items, Total: total}, nil
}

func (uc *CorrectionUsecase) Delete(ctx context.Context, id string) (*entity.AttendanceCorrection, error) {
	c, err := uc.correctionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.WrapInternal("failed to find correction", err)
	}
	if c == nil {
		return nil, errors.NewNotFound("correction not found")
	}

	tx, err := uc.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.WrapInternal("failed to begin transaction", err)
	}
	defer tx.Rollback()

	correctionRepo := uc.correctionRepo.WithTx(tx)
	dailyRepo := uc.dailyRepo.WithTx(tx)
	processor := NewDailyProcessor(dailyRepo, uc.resolver)

	if err := correctionRepo.Delete(ctx, id); err != nil {
		return nil, errors.WrapInternal("failed to delete correction", err)
	}

	if _, err := processor.ProcessDaily(ctx, c.EmployeeID, c.Date); err != nil {
		return nil, errors.WrapInternal("failed to restore attendance", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.WrapInternal("failed to commit transaction", err)
	}

	return c, nil
}
