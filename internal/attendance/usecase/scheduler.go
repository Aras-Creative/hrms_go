package usecase

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
)

const lockID int64 = 0x48524D53

type Scheduler struct {
	db        *sqlx.DB
	processor *DailyProcessor
	loc       *time.Location
	c         *cron.Cron
}

func NewScheduler(db *sqlx.DB, processor *DailyProcessor, loc *time.Location) *Scheduler {
	s := &Scheduler{
		db:        db,
		processor: processor,
		loc:       loc,
	}
	s.c = cron.New(
		cron.WithLocation(loc),
	)
	return s
}

func (s *Scheduler) Start(_ context.Context) {
	s.c.AddFunc("0 12,14,16,18,20,22 * * *", s.runSweep)
	s.c.AddFunc("30 0 * * *", s.runSweep)
	s.c.Start()
	slog.Info("scheduler: attendance finalize started", "tz", s.loc.String())
}

func (s *Scheduler) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
}

func (s *Scheduler) runSweep() {
	ctx := context.Background()
	now := time.Now().In(s.loc)
	hour := now.Hour()
	minute := now.Minute()

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.loc)

	// At 00:30 the sweep finalises yesterday's day — the 00:30-next-day cutoff
	// has just passed, so yesterday's no_punch entries become absent.
	date := today
	if hour == 0 && minute == 30 {
		date = today.AddDate(0, 0, -1)
	}

	conn, acquired := s.tryLock(ctx)
	if !acquired {
		slog.Info("scheduler: advisory lock not acquired, skipping finalize")
		return
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockID)
		conn.Close()
	}()

	if err := s.processor.ProcessRange(ctx, date, date); err != nil {
		slog.Error("scheduler: ProcessRange failed", "error", err)
		return
	}

	slog.Info("scheduler: finalize sweep completed", "date", date.Format("2006-01-02"))
}

func (s *Scheduler) tryLock(ctx context.Context) (*sql.Conn, bool) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		slog.Error("scheduler: failed to get connection for lock", "error", err)
		return nil, false
	}
	var acquired bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired); err != nil {
		slog.Error("scheduler: advisory lock query failed", "error", err)
		conn.Close()
		return nil, false
	}
	if !acquired {
		conn.Close()
		return nil, false
	}
	return conn, true
}
