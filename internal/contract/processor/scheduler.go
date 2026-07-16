package processor

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

const contractExpiryLockID int64 = 0x434F4E45 // "CONE" in hex

type Scheduler struct {
	db        *sqlx.DB
	processor *ExpiryProcessor
	loc       *time.Location
	mu        sync.Mutex
	doneDate  string
	cancel    context.CancelFunc
}

func NewScheduler(db *sqlx.DB, processor *ExpiryProcessor, timezone string) *Scheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return &Scheduler{
		db:        db,
		processor: processor,
		loc:       loc,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)
	go s.loop(ctx)
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	now := time.Now().In(s.loc)
	today := now.Format("2006-01-02")

	s.mu.Lock()
	if s.doneDate == today {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	s.runOnce(ctx, today, now)
}

func (s *Scheduler) runOnce(ctx context.Context, today string, now time.Time) {
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

	result, err := conn.ExecContext(ctx,
		`INSERT INTO job_runs (target, run_date) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		"contract.expiry", today)
	if err != nil {
		slog.Error("contract expiry scheduler: failed to insert job_run", "error", err)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		slog.Info("contract expiry scheduler: already ran (another instance)", "date", today)
	}

	s.mu.Lock()
	s.doneDate = today
	s.mu.Unlock()
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
