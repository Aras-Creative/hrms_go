package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	attendanceEntity "hrms/internal/attendance/entity"
	attendanceRepo "hrms/internal/attendance/repository"
	emplModels "hrms/internal/employee/models"
	emplRepo "hrms/internal/employee/repository"
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

func findEmployeeByName(ctx context.Context, empRepo emplRepo.EmployeeRepository, nameSlug string) (string, string, error) {
	// Search by first name — DB uses ILIKE '%searchName%'
	result, total, err := empRepo.FindAllWithDetails(ctx, emplModels.ListEmployeeInput{
		SearchName: nameSlug,
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		return "", "", fmt.Errorf("find employee: %w", err)
	}

	if total == 0 {
		return "", "", fmt.Errorf("no employee found matching %q", nameSlug)
	}

	// Try exact first-name match first
	for _, e := range result {
		first := strings.ToLower(strings.Split(e.FullName, " ")[0])
		if first == nameSlug {
			return e.ID, e.FullName, nil
		}
	}

	// Fall back to the first result
	e := result[0]
	log.Printf("  warning: no exact match for %q, using %q (id=%s)", nameSlug, e.FullName, e.ID)
	return e.ID, e.FullName, nil
}

func main() {
	cfgPath := flag.String("config", "config/config.yaml", "path to configuration file")
	dataDir := flag.String("dir", "data", "directory containing JSON seed files")
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
	now := time.Now()

	dailyRepo := attendanceRepo.NewPostgresDailyAttendanceRepo(db)
	empRepo := emplRepo.NewPostgresEmployeeRepo(db)
	ctx := context.Background()

	entries, err := os.ReadDir(*dataDir)
	if err != nil {
		log.Fatalf("read dir %s: %v", *dataDir, err)
	}

	type job struct {
		file    string
		empID   string
		empName string
		records []LegacyRecord
	}
	var jobs []job

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(*dataDir, entry.Name())
		nameSlug := strings.TrimSuffix(entry.Name(), ".json")

		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("skip %s: read error: %v", entry.Name(), err)
			continue
		}

		var records []LegacyRecord
		if err := json.Unmarshal(data, &records); err != nil {
			log.Printf("skip %s: parse error: %v", entry.Name(), err)
			continue
		}
		if len(records) == 0 {
			log.Printf("skip %s: empty records", entry.Name())
			continue
		}

		empID, empName, err := findEmployeeByName(ctx, empRepo, nameSlug)
		if err != nil {
			log.Printf("skip %s: %v", entry.Name(), err)
			continue
		}

		jobs = append(jobs, job{
			file:    entry.Name(),
			empID:   empID,
			empName: empName,
			records: records,
		})
	}

	if len(jobs) == 0 {
		log.Fatal("no valid seed files found")
	}

	for _, j := range jobs {
		inserted := 0
		for _, r := range j.records {
			date, err := time.ParseInLocation("2006-01-02", r.Date, loc)
			if err != nil {
				log.Printf("  [%s] skip invalid date %q: %v", j.file, r.Date, err)
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
				EmployeeID:         j.empID,
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
				log.Printf("  [%s] ERROR upsert %s: %v", j.file, r.Date, err)
				continue
			}
			inserted++
		}

		fmt.Printf("%s (%s): %d/%d records upserted\n", j.file, j.empName, inserted, len(j.records))
	}
}
