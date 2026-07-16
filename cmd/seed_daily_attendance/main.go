package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	attendanceEntity "hrms/internal/attendance/entity"
	attendanceRepo "hrms/internal/attendance/repository"
	"hrms/internal/pkg/config"
	"hrms/internal/pkg/database"
	"hrms/internal/pkg/timeutil"

	"github.com/google/uuid"
)

type LegacyRecord struct {
	Date              string  `json:"date"`
	Status            string  `json:"status"`
	IsLate            bool    `json:"is_late"`
	IsEarlyLeave      bool    `json:"is_early_leave"`
	ExpectedStartTime *string `json:"expected_start_time"`
	ExpectedEndTime   *string `json:"expected_end_time"`
	Source            string  `json:"source"`
	FirstPunchIn      *string `json:"first_punch_in"`
	LastPunchOut      *string `json:"last_punch_out"`
	TotalWorkSeconds  *int    `json:"total_work_seconds"`
}

func main() {
	cfgPath := flag.String("config", "config/config.yaml", "path to configuration file")
	employeeID := flag.String("employee", "", "employee UUID")
	filePath := flag.String("file", "", "path to JSON file with legacy attendance records")
	flag.Parse()

	if *employeeID == "" || *filePath == "" {
		log.Fatal("usage: seed_daily_attendance -employee <uuid> -file legacy_data.json")
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("read file: %v", err)
	}

	var records []LegacyRecord
	if err := json.Unmarshal(data, &records); err != nil {
		log.Fatalf("parse json: %v", err)
	}

	if len(records) == 0 {
		log.Fatal("no records found in JSON")
	}

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
	now := time.Now()

	dailyRepo := attendanceRepo.NewPostgresDailyAttendanceRepo(db)
	ctx := context.Background()

	inserted := 0
	for _, r := range records {
		date, err := time.ParseInLocation("2006-01-02", r.Date, loc)
		if err != nil {
			log.Printf("skip invalid date %q: %v", r.Date, err)
			continue
		}

		if r.Source == "" {
			r.Source = "legacy"
		}

		var firstPunchIn, lastPunchOut *time.Time
		if r.FirstPunchIn != nil && *r.FirstPunchIn != "" {
			t, err := time.ParseInLocation("2006-01-02T15:04:05", *r.FirstPunchIn, loc)
			if err != nil {
				t, err = time.ParseInLocation("2006-01-02 15:04:05", *r.FirstPunchIn, loc)
			}
			if err == nil {
				firstPunchIn = &t
			}
		}
		if r.LastPunchOut != nil && *r.LastPunchOut != "" {
			t, err := time.ParseInLocation("2006-01-02T15:04:05", *r.LastPunchOut, loc)
			if err != nil {
				t, err = time.ParseInLocation("2006-01-02 15:04:05", *r.LastPunchOut, loc)
			}
			if err == nil {
				lastPunchOut = &t
			}
		}

		da := &attendanceEntity.DailyAttendance{
			ID:                 uuid.New().String(),
			EmployeeID:         *employeeID,
			Date:               date,
			Status:             attendanceEntity.AttendanceStatus(r.Status),
			IsLate:             r.IsLate,
			IsEarlyLeave:       r.IsEarlyLeave,
			ExpectedStartTime:  r.ExpectedStartTime,
			ExpectedEndTime:    r.ExpectedEndTime,
			Source:             r.Source,
			FirstPunchIn:       firstPunchIn,
			LastPunchOut:       lastPunchOut,
			TotalWorkSeconds:   r.TotalWorkSeconds,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		if err := dailyRepo.Upsert(ctx, da); err != nil {
			log.Printf("  ERROR upsert %s: %v", r.Date, err)
			continue
		}
		inserted++
	}

	fmt.Printf("done. %d/%d records upserted for employee %s\n", inserted, len(records), *employeeID)
}
