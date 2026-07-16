package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	authAdapter "hrms/internal/auth/adapter"
	authEntity "hrms/internal/auth/entity"
	authRepo "hrms/internal/auth/repository"
	empEntity "hrms/internal/employee/entity"
	empRepo "hrms/internal/employee/repository"
	"hrms/internal/pkg/config"
	"hrms/internal/pkg/database"

	"github.com/google/uuid"
)

type SeedRecord struct {
	Username               string  `json:"username"`
	Password               string  `json:"password"`
	Role                   string  `json:"role"`
	FullName               string  `json:"full_name"`
	EmployeeNumber         string  `json:"employee_number"`
	Phone                  string  `json:"phone"`
	PersonalEmail          *string `json:"personal_email"`
	EmergencyContactName   string  `json:"emergency_contact_name"`
	EmergencyContactPhone  string  `json:"emergency_contact_phone"`
	PlaceOfBirth           string  `json:"place_of_birth"`
	DateOfBirth            *string `json:"date_of_birth"`
	JoinDate               *string `json:"join_date"`
	Gender                 string  `json:"gender"`
	Education              string  `json:"education"`
	Status                 string  `json:"status"`
	Address                string  `json:"address"`
	DesignationID          *string `json:"designation_id"`
	NationalID             string  `json:"national_id"`
	Religion               string  `json:"religion"`
	ProfilePhotoID         *string `json:"profile_photo_id"`
	BankHolder             string  `json:"bank_holder"`
	BankName               string  `json:"bank_name"`
	BankNumber             string  `json:"bank_number"`
	IsActive               bool    `json:"is_active"`
	EmployeeID             *string `json:"employee_id"`
	UserID                 *string `json:"user_id"`
}

func main() {
	cfgPath := flag.String("config", "config/config.yaml", "path to configuration file")
	filePath := flag.String("file", "", "path to JSON file or directory containing JSON files")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("usage: seed_account_employee -file <path> [-config config/config.yaml]\n  <path> can be a single .json file or a directory (reads all *.json files)")
	}

	var jsonFiles []string
	info, err := os.Stat(*filePath)
	if err != nil {
		log.Fatalf("access path: %v", err)
	}

	if info.IsDir() {
		matches, err := filepath.Glob(filepath.Join(*filePath, "*.json"))
		if err != nil {
			log.Fatalf("glob directory: %v", err)
		}
		sort.Strings(matches)
		jsonFiles = matches
		if len(jsonFiles) == 0 {
			log.Fatalf("no .json files found in %s", *filePath)
		}
		fmt.Printf("found %d JSON file(s) in %s\n", len(jsonFiles), *filePath)
	} else {
		jsonFiles = []string{*filePath}
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

	hasher := authAdapter.NewBcryptHasherWithCost(12)
	userRepo := authRepo.NewPostgresUserRepo(db)
	employeeRepo := empRepo.NewPostgresEmployeeRepo(db)
	ctx := context.Background()
	now := time.Now()

	totalCreated := 0
	totalRecords := 0
	for _, f := range jsonFiles {
		records := loadJSON(f)
		totalRecords += len(records)
		created := seedRecords(ctx, hasher, userRepo, employeeRepo, records, now)
		totalCreated += created
		fmt.Printf("  %s: %d/%d created\n", filepath.Base(f), created, len(records))
	}

	fmt.Printf("\ndone. %d/%d total records created\n", totalCreated, totalRecords)
}

func loadJSON(path string) []SeedRecord {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s: %v", path, err)
	}

	var records []SeedRecord
	if err := json.Unmarshal(data, &records); err != nil {
		log.Fatalf("parse %s: %v", path, err)
	}
	return records
}

func seedRecords(
	ctx context.Context,
	hasher *authAdapter.BcryptHasher,
	userRepo authRepo.UserRepository,
	employeeRepo empRepo.EmployeeRepository,
	records []SeedRecord,
	now time.Time,
) int {
	created := 0
	for i, r := range records {
		if r.FullName == "" {
			r.FullName = r.Username
		}

		userID := uuid.New().String()
		if r.UserID != nil && *r.UserID != "" {
			userID = *r.UserID
		}

		hash, err := hasher.Hash(r.Password)
		if err != nil {
			log.Printf("  [%d] hash password: %v", i, err)
			continue
		}

		role := authEntity.RoleUser
		if r.Role != "" {
			role = authEntity.Role(r.Role)
		}

		user := &authEntity.User{
			ID:           userID,
			Username:     r.Username,
			PasswordHash: hash,
			FullName:     r.FullName,
			IsActive:     true,
			Role:         role,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := userRepo.Create(ctx, user); err != nil {
			log.Printf("  [%d] create user %q: %v", i, r.Username, err)
			continue
		}

		empID := uuid.New().String()
		if r.EmployeeID != nil && *r.EmployeeID != "" {
			empID = *r.EmployeeID
		}

		var dateOfBirth *empEntity.Date
		if r.DateOfBirth != nil && *r.DateOfBirth != "" {
			d, err := empEntity.ParseDate(*r.DateOfBirth)
			if err != nil {
				log.Printf("  [%d] parse date_of_birth %q: %v", i, *r.DateOfBirth, err)
			} else {
				dateOfBirth = &d
			}
		}

		var joinDate *empEntity.Date
		if r.JoinDate != nil && *r.JoinDate != "" {
			d, err := empEntity.ParseDate(*r.JoinDate)
			if err != nil {
				log.Printf("  [%d] parse join_date %q: %v", i, *r.JoinDate, err)
			} else {
				joinDate = &d
			}
		}

		empStatus := empEntity.StatusActive
		if r.Status != "" {
			empStatus = empEntity.Status(r.Status)
		}

		emp := &empEntity.Employee{
			ID:                    empID,
			UserID:                &userID,
			FullName:              r.FullName,
			EmployeeNumber:        empEntity.FromString(r.EmployeeNumber),
			Phone:                 empEntity.PhoneFromDB(r.Phone),
			PersonalEmail:         coalesceStr(r.PersonalEmail),
			EmergencyContactName:  r.EmergencyContactName,
			EmergencyContactPhone: empEntity.PhoneFromDB(r.EmergencyContactPhone),
			PlaceOfBirth:          r.PlaceOfBirth,
			DateOfBirth:           dateOfBirth,
			JoinDate:              joinDate,
			Gender:                empEntity.Gender(r.Gender),
			Education:             r.Education,
			Status:                empStatus,
			Address:               r.Address,
			DesignationID:         r.DesignationID,
			NationalID:            r.NationalID,
			Religion:              empEntity.Religion(r.Religion),
			ProfilePhotoID:        r.ProfilePhotoID,
			BankAccount:           empEntity.BankAccountFromDB(r.BankHolder, r.BankName, r.BankNumber),
			IsActive:              r.IsActive,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		if err := employeeRepo.Create(ctx, emp); err != nil {
			log.Printf("  [%d] create employee %q: %v", i, r.Username, err)
			continue
		}
		created++
	}
	return created
}

func coalesceStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
