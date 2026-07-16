package usecase

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

const lockID int64 = 0x48524D53

type Scheduler struct {
	db        *sqlx.DB
	processor *DailyProcessor
	loc       *time.Location
	targets   []int
	mu        sync.Mutex
	done      map[string]bool
	cancel    context.CancelFunc
}

func NewScheduler(db *sqlx.DB, processor *DailyProcessor, timezone string) *Scheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return &Scheduler{
		db:        db,
		processor: processor,
		loc:       loc,
		targets:   []int{720, 960, 975, 1080, 1380},
		done:      make(map[string]bool),
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
	ticker := time.NewTicker(30 * time.Second)
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
	minutes := now.Hour()*60 + now.Minute()

	s.mu.Lock()
	for k := range s.done {
		if k[:strings.IndexByte(k, ':')] != today {
			delete(s.done, k)
		}
	}

	for _, tgt := range s.targets {
		key := today + ":" + strconv.Itoa(tgt)
		if s.done[key] {
			continue
		}
		if minutes < tgt {
			continue
		}
		s.mu.Unlock()

		s.runOnce(ctx, tgt, today, now, key)

		s.mu.Lock()
	}
	s.mu.Unlock()
}

func (s *Scheduler) runOnce(ctx context.Context, tgt int, today string, now time.Time, key string) {
	conn, acquired := s.tryLock(ctx)
	if !acquired {
		slog.Info("scheduler: advisory lock not acquired, skipping run", "target", tgt)
		return
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockID)
		conn.Close()
	}()

	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.loc)
	if err := s.processor.ProcessRange(ctx, date, date); err != nil {
		slog.Error("scheduler: ProcessRange failed", "target", tgt, "error", err)
		return
	}

	// Insert after success — prevents job_runs row from blocking retry on failure
	target := "attendance.process_daily." + strconv.Itoa(tgt)
	result, err := conn.ExecContext(ctx, `INSERT INTO job_runs (target, run_date) VALUES ($1, $2) ON CONFLICT DO NOTHING`, target, today)
	if err != nil {
		slog.Error("scheduler: failed to insert job_run", "target", tgt, "error", err)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		slog.Info("scheduler: job already ran (another instance)", "target", tgt, "date", today)
	}

	s.mu.Lock()
	s.done[key] = true
	s.mu.Unlock()
	slog.Info("scheduler: daily attendance processed", "target", tgt, "date", today)
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
