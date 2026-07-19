package processor

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
)

const contractExpiryLockID int64 = 0x434F4E45 // "CONE" in hex

type Scheduler struct {
	db        *sqlx.DB
	processor *ExpiryProcessor
	loc       *time.Location
	c         *cron.Cron
}

func NewScheduler(db *sqlx.DB, processor *ExpiryProcessor, loc *time.Location) *Scheduler {
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
	s.c.AddFunc("0 0 * * *", s.runOnce)
	s.c.Start()
	slog.Info("scheduler: contract expiry started", "tz", s.loc.String())
}

func (s *Scheduler) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
}

func (s *Scheduler) runOnce() {
	ctx := context.Background()
	now := time.Now().In(s.loc)
	today := now.Format("2006-01-02")

	conn, acquired := s.tryLock(ctx)
	if !acquired {
		return
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", contractExpiryLockID)
		conn.Close()
	}()

	if err := s.processor.Process(ctx, now); err != nil {
		slog.Error("contract expiry scheduler: Process failed", "error", err)
		return
	}

	slog.Info("contract expiry scheduler: completed", "date", today)
}

func (s *Scheduler) tryLock(ctx context.Context) (*sql.Conn, bool) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		slog.Error("contract expiry scheduler: failed to get connection", "error", err)
		return nil, false
	}
	var acquired bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", contractExpiryLockID).Scan(&acquired); err != nil {
		slog.Error("contract expiry scheduler: advisory lock query failed", "error", err)
		conn.Close()
		return nil, false
	}
	if !acquired {
		conn.Close()
		return nil, false
	}
	return conn, true
}
