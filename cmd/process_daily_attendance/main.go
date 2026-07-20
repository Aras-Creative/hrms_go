package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	attendanceAdapter "hrms/internal/attendance/adapter"
	attendanceRepo "hrms/internal/attendance/repository"
	attendanceUc "hrms/internal/attendance/usecase"
	"hrms/internal/pkg/config"
	"hrms/internal/pkg/database"
	"hrms/internal/pkg/timeutil"
	scheduleRepo "hrms/internal/schedule/repository"
)

func main() {
	cfgPath := flag.String("config", "config/config.yaml", "path to configuration file")
	dateFlag := flag.String("date", "", "process a single date (YYYY-MM-DD, defaults to today)")
	fromFlag := flag.String("from", "", "range start (YYYY-MM-DD)")
	toFlag := flag.String("to", "", "range end (YYYY-MM-DD)")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	timeutil.SetDefaultTimezone("Asia/Jakarta")
	loc := timeutil.LoadDefaultLocation()

	var from, to time.Time
	switch {
	case *fromFlag != "" && *toFlag != "":
		from, err = time.ParseInLocation("2006-01-02", *fromFlag, loc)
		if err != nil {
			log.Fatalf("invalid --from: %v", err)
		}
		to, err = time.ParseInLocation("2006-01-02", *toFlag, loc)
		if err != nil {
			log.Fatalf("invalid --to: %v", err)
		}
	case *dateFlag != "":
		from, err = time.ParseInLocation("2006-01-02", *dateFlag, loc)
		if err != nil {
			log.Fatalf("invalid --date: %v", err)
		}
		to = from
	default:
		now := time.Now().In(loc)
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		to = from
	}

	dailyRepo := attendanceRepo.NewPostgresDailyAttendanceRepo(db)
	overrideRepo := scheduleRepo.NewPostgresScheduleOverrideRepo(db)
	scheduleResolver := attendanceAdapter.NewScheduleResolverAdapter(overrideRepo)
	processor := attendanceUc.NewDailyProcessor(dailyRepo, scheduleResolver)

	if from == to {
		log.Printf("processing date: %s", from.Format("2006-01-02"))
	} else {
		log.Printf("processing range: %s to %s", from.Format("2006-01-02"), to.Format("2006-01-02"))
	}

	start := time.Now()
	ctx := context.Background()

	if err := processor.ProcessRange(ctx, from, to); err != nil {
		log.Fatalf("process range failed: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("done in %s\n", elapsed)
}
