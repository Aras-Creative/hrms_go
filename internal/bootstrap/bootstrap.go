package bootstrap

import (
	"context"
	"log"
	"time"

	attendanceAdapter "hrms/internal/attendance/adapter"
	attendanceDelivery "hrms/internal/attendance/delivery"
	attendanceRepo "hrms/internal/attendance/repository"
	attendanceUc "hrms/internal/attendance/usecase"
	auditRepo "hrms/internal/audit/repository"
	auditUc "hrms/internal/audit/usecase"
	auditDelivery "hrms/internal/audit/delivery"
	authAdapter "hrms/internal/auth/adapter"
	authdelivery "hrms/internal/auth/delivery"
	authrepo "hrms/internal/auth/repository"
	authuc "hrms/internal/auth/usecase"
	contractAdapter "hrms/internal/contract/adapter"
	contractDelivery "hrms/internal/contract/delivery"
	contractProcessor "hrms/internal/contract/processor"
	contractRepo "hrms/internal/contract/repository"
	contractUc "hrms/internal/contract/usecase"
	dashboardDelivery "hrms/internal/dashboard/delivery"
	dashboardRepo "hrms/internal/dashboard/repository"
	dashboardUc "hrms/internal/dashboard/usecase"
	desigAdapter "hrms/internal/designation/adapter"
	designationDelivery "hrms/internal/designation/delivery"
	designationRepo "hrms/internal/designation/repository"
	designationUc "hrms/internal/designation/usecase"
	emplAdapter "hrms/internal/employee/adapter"
	emplDelivery "hrms/internal/employee/delivery"
	"hrms/internal/employee/numbergen"
	emplRepo "hrms/internal/employee/repository"
	emplUC "hrms/internal/employee/usecase"
	eventsDelivery "hrms/internal/events/delivery"
	leaveAdapter "hrms/internal/leave/adapter"
	leaveDelivery "hrms/internal/leave/delivery"
	leaveRepo "hrms/internal/leave/repository"
	leaveUc "hrms/internal/leave/usecase"
	notificationAdapter "hrms/internal/notification/adapter"
	notificationDelivery "hrms/internal/notification/delivery"
	notificationRepo "hrms/internal/notification/repository"
	notificationUc "hrms/internal/notification/usecase"
	payrollAdapter "hrms/internal/payroll/adapter"
	payrollDelivery "hrms/internal/payroll/delivery"
	"hrms/internal/payroll/processor"
	payrollRepo "hrms/internal/payroll/repository"
	payrollUc "hrms/internal/payroll/usecase"
	"hrms/internal/pkg/config"
	"hrms/internal/pkg/database"
	hrmsJwt "hrms/internal/pkg/jwt"
	"hrms/internal/pkg/logger"
	"hrms/internal/pkg/middleware"
	"hrms/internal/pkg/sse"
	"hrms/internal/pkg/timeutil"
	scheduleAdapter "hrms/internal/schedule/adapter"
	scheduleDelivery "hrms/internal/schedule/delivery"
	scheduleRepo "hrms/internal/schedule/repository"
	scheduleUc "hrms/internal/schedule/usecase"
	"hrms/internal/server"
	settingAdapter "hrms/internal/setting/adapter"
	settingDelivery "hrms/internal/setting/delivery"
	settingRepo "hrms/internal/setting/repository"
	settingUc "hrms/internal/setting/usecase"
	storageAdapter "hrms/internal/storage/adapter"
	storageDelivery "hrms/internal/storage/delivery"
	storageRepo "hrms/internal/storage/repository"
	storageUc "hrms/internal/storage/usecase"
)

func Run(cfgPath string) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log, err := logger.New(&cfg.Log)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}

	db, err := database.NewPostgres(&cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	app := server.New(&cfg.Server, log)

	api := app.Group("/api/v1")

	accessTTL, err := time.ParseDuration(cfg.Auth.AccessTokenTTL)
	if err != nil {
		log.Fatalf("invalid access_token_ttl %q: %v", cfg.Auth.AccessTokenTTL, err)
	}
	refreshTTL, err := time.ParseDuration(cfg.Auth.RefreshTokenTTL)
	if err != nil {
		log.Fatalf("invalid refresh_token_ttl %q: %v", cfg.Auth.RefreshTokenTTL, err)
	}
	challengeTTL, err := time.ParseDuration(cfg.Auth.ChallengeTTL)
	if err != nil {
		log.Fatalf("invalid challenge_ttl %q: %v", cfg.Auth.ChallengeTTL, err)
	}
	log.Printf("access_token_ttl=%s refresh_token_ttl=%s challenge_ttl=%s", accessTTL, refreshTTL, challengeTTL)

	jwtCfg := hrmsJwt.Config{
		Secret:          cfg.Auth.JWTSecret,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	}
	jwtSvc := hrmsJwt.New(jwtCfg)

	authmw := middleware.NewAuthMiddleware(jwtSvc)

	passwordhasher := authAdapter.NewBcryptHasherWithCost(12)
	challengehasher := authAdapter.NewSHA256ChallengeHasher()
	tokengen := authAdapter.NewJWTAdapter(jwtSvc)

	userRepo := authrepo.NewPostgresUserRepo(db)
	sessionRepo := authrepo.NewPostgresSessionRepo(db)
	challengeRepo := authrepo.NewPostgresChallengeRepo(db)
	deviceRepo := authrepo.NewPostgresDeviceRepo(db)
	userUc := authuc.NewUserUsecase(userRepo)

	// Audit (early for all handlers)
	auditRepo := auditRepo.NewPostgresAuditRepo(db)
	auditUC := auditUc.NewAuditUsecase(auditRepo)
	auditHandler := auditDelivery.NewAuditHandler(auditUC)
	auditHandler.RegisterRoutes(api, authmw)

	// Module-specific audit loggers
	attendanceAuditLog := attendanceAdapter.NewAuditLogger(auditUC)
	employeeAuditLog := emplAdapter.NewAuditLogger(auditUC)
	designationAuditLog := desigAdapter.NewAuditLogger(auditUC)
	scheduleAuditLog := scheduleAdapter.NewAuditLogger(auditUC)
	leaveAuditLog := leaveAdapter.NewAuditLogger(auditUC)
	contractAuditLog := contractAdapter.NewAuditLogger(auditUC)
	payrollAuditLog := payrollAdapter.NewAuditLogger(auditUC)
	authAuditLog := authAdapter.NewAuditLogger(auditUC)
	storageAuditLog := storageAdapter.NewAuditLogger(auditUC)

	punchHub := sse.NewHub()

	// Notifications
	notificationRepo := notificationRepo.NewPostgresNotificationRepo(db)
	notificationUC := notificationUc.NewNotificationUsecase(notificationRepo, punchHub)
	notifAdapter := notificationAdapter.NewNotifierAdapter(notificationUC)

	authuc := authuc.NewAuthUsecase(userRepo, sessionRepo, challengeRepo, deviceRepo, passwordhasher, challengehasher, tokengen, accessTTL, refreshTTL, challengeTTL)
	authhandler := authdelivery.NewAuthHandler(authuc, userUc, authmw, accessTTL, refreshTTL, cfg.Auth.SecureCookies, authAuditLog, notifAdapter)

	authhandler.RegisterRoutes(api)

	// Employees (early for adapters)
	emplRepo := emplRepo.NewPostgresEmployeeRepo(db)
	leaveSubmissionRepo := leaveRepo.NewPostgresLeaveSubmissionRepo(db)

	// Designations
	designationRepo := designationRepo.NewPostgresDesignationRepo(db)
	designationUc := designationUc.NewDesignationUsecase(designationRepo)

	// Schedule
	workPatternRepo := scheduleRepo.NewPostgresWorkPatternRepo(db)
	workPatternUc := scheduleUc.NewWorkPatternUsecase(workPatternRepo)
	overrideRepo := scheduleRepo.NewPostgresScheduleOverrideRepo(db)
	overrideUc := scheduleUc.NewScheduleOverrideUsecase(overrideRepo)
	overviewUc := scheduleUc.NewScheduleOverviewUsecase(overrideRepo)
	ewpRepo := scheduleRepo.NewPostgresEmployeeWorkPatternRepo(db)
	scheduleEmpChecker := scheduleAdapter.NewEmployeeExistsAdapter(emplRepo)
	scheduleEmpFetcher := scheduleAdapter.NewEmployeeFetcherAdapter(emplRepo)
	ewpUc := scheduleUc.NewEmployeePatternUsecase(ewpRepo, workPatternRepo, scheduleEmpChecker)
	scheduleHandler := scheduleDelivery.NewScheduleHandler(workPatternUc, ewpUc, overrideUc, overviewUc, scheduleEmpFetcher, scheduleAuditLog, notifAdapter)
	scheduleHandler.RegisterRoutes(api, authmw)

	adminmw := middleware.NewAdminMiddleware()

	// Storage infra (early for settings logo resolver and below)
	docRepo := storageRepo.NewDocumentRepo(db)
	r2Storage := storageRepo.NewR2Storage(&cfg.Storage.S3)
	urlResolver := storageRepo.NewCDNURLResolver(&cfg.Storage.S3)

	// Settings — seed cache and apply timezone before any scheduler starts
	settingRepo := settingRepo.NewPostgresSettingRepo(db)
	settingUC := settingUc.NewSettingUsecase(settingRepo, settingAdapter.NewLogoResolverAdapter(docRepo, urlResolver))
	{
		s, err := settingUC.Get(context.Background())
		if err == nil && s != nil && s.Timezone != "" {
			timeutil.SetDefaultTimezone(s.Timezone)
			log.Printf("timezone set from settings: %s", s.Timezone)
		} else {
			timeutil.SetDefaultTimezone("Asia/Jakarta")
			log.Printf("timezone set to default: Asia/Jakarta (settings err=%v s=%v)", err, s)
		}
	}
	settingHandler := settingDelivery.NewSettingHandler(settingUC)
	settingHandler.RegisterRoutes(api, authmw, adminmw)

	loc, err := time.LoadLocation(timeutil.DefaultTimezone)
	if err != nil {
		loc = time.UTC
		log.Printf("WARN: LoadLocation(%s) failed: %v — falling back to UTC", timeutil.DefaultTimezone, err)
	} else {
		log.Printf("scheduler location loaded: %s", loc.String())
	}

	// Attendance
	punchRepo := attendanceRepo.NewPostgresPunchRepo(db)
	dailyRepo := attendanceRepo.NewPostgresDailyAttendanceRepo(db)
	scheduleResolver := attendanceAdapter.NewScheduleResolverAdapter(overrideRepo)
	dailyProcessor := attendanceUc.NewDailyProcessor(dailyRepo, scheduleResolver)
	scheduler := attendanceUc.NewScheduler(db, dailyProcessor, loc)
	scheduler.Start(context.Background())
	defer scheduler.Stop()
	attendanceLeaveFetcher := attendanceAdapter.NewLeaveFetcherAdapter(leaveSubmissionRepo)

	punchUc := attendanceUc.NewPunchUsecase(punchRepo, dailyProcessor, attendanceLeaveFetcher, punchHub)
	attendanceEmpFetcher := attendanceAdapter.NewEmployeeFetcherAdapter(emplRepo)
	correctionRepo := attendanceRepo.NewPostgresCorrectionRepo(db)
	dailyUc := attendanceUc.NewDailyAttendanceUsecase(dailyRepo, correctionRepo, punchRepo, dailyProcessor)
	correctionUc := attendanceUc.NewCorrectionUsecase(db, correctionRepo, dailyRepo, attendanceEmpFetcher, scheduleResolver)
	correctionAuditFetcher := attendanceAdapter.NewCorrectionAuditFetcherAdapter(auditUC)
	attendanceMeUc := attendanceUc.NewMeUsecase(attendanceEmpFetcher, dailyProcessor, dailyRepo, correctionAuditFetcher)

	notificationHandler := notificationDelivery.NewNotificationHandler(notificationUC)
	sseAuthMw := middleware.NewSSEAuthMiddleware(jwtSvc)
	notificationHandler.RegisterRoutes(api, authmw)

	// Global SSE stream
	eventsHandler := eventsDelivery.NewEventHandler(punchHub)
	eventsHandler.RegisterRoutes(api, sseAuthMw)

	// Employees
	designationFetcher := emplAdapter.NewDesignationFetcherAdapter(designationRepo)
	accountCreator := emplAdapter.NewAccountCreatorAdapter(authuc)
	emplNumberGen := numbergen.New(emplRepo, "1001")

	// Storage — photo resolver (docRepo, r2Storage, urlResolver already initialized above)
	photoResolver := emplAdapter.NewProfilePhotoResolverAdapter(docRepo, urlResolver)

	designationHandler := designationDelivery.NewDesignationHandler(designationUc, photoResolver, designationAuditLog)
	designationHandler.RegisterRoutes(api, authmw)

	attendanceHandler := attendanceDelivery.NewAttendanceHandler(punchUc, dailyUc, correctionUc, attendanceMeUc, attendanceEmpFetcher, punchHub, auditUC, photoResolver, attendanceAuditLog, notifAdapter)
	attendanceHandler.RegisterRoutes(api, authmw, adminmw)

	// Leave
	leaveTypeRepo := leaveRepo.NewPostgresLeaveTypeRepo(db)
	leaveBalanceRepo := leaveRepo.NewPostgresLeaveBalanceRepo(db)
	employeeFetcher := leaveAdapter.NewEmployeeFetcherAdapter(emplRepo)
	attendanceProcessor := leaveAdapter.NewAttendanceProcessorAdapter(dailyProcessor, punchRepo, emplRepo)
	leaveUc := leaveUc.NewLeaveUsecase(db, leaveTypeRepo, leaveBalanceRepo, leaveSubmissionRepo, employeeFetcher, leaveAdapter.NewUserNameAdapter(userRepo), attendanceProcessor, attendanceProcessor, log)
	attachmentResolver := leaveAdapter.NewStorageAttachmentResolver(docRepo, urlResolver)
	leaveHandler := leaveDelivery.NewLeaveHandler(leaveUc, attachmentResolver, photoResolver, leaveAuditLog, notifAdapter)
	leaveHandler.RegisterRoutes(api, authmw, adminmw)

	balanceAssigner := emplAdapter.NewBalanceAssignerAdapter(leaveUc)
	emplUC := emplUC.NewEmployeeUsecase(emplRepo, designationFetcher, emplNumberGen, accountCreator, balanceAssigner, photoResolver, nil)
	emplHandler := emplDelivery.NewEmployeeHandler(emplUC, employeeAuditLog)
	emplHandler.RegisterRoutes(api, authmw)

	// Payroll
	payrollEmployeeFetcher := payrollAdapter.NewEmployeeFetcherAdapter(emplRepo, designationRepo)
	payrollSalaryRepo := payrollRepo.NewPostgresEmployeeBaseSalaryRepo(db)
	payrollCompItemRepo := payrollRepo.NewPostgresCompensationItemRepo(db)
	payrollEmpCompRepo := payrollRepo.NewPostgresEmployeeCompensationRepo(db)
	payrollBenefitTypeRepo := payrollRepo.NewPostgresBenefitTypeRepo(db)
	payrollEmpBenefitRepo := payrollRepo.NewPostgresEmployeeBenefitRepo(db)
	payrollDeductionTypeRepo := payrollRepo.NewPostgresDeductionTypeRepo(db)
	payrollEmpDeductionRepo := payrollRepo.NewPostgresEmployeeDeductionRepo(db)
	payrollPeriodRepo := payrollRepo.NewPostgresPayrollPeriodRepo(db)
	payrollPaySlipRepo := payrollRepo.NewPostgresPaySlipRepo(db)
	payrollOverviewRepo := payrollRepo.NewPostgresOverviewRepo(db)
	payrollCalcRepo := payrollRepo.NewPostgresCalculationRepo(db)
	payrollProc := processor.New(payrollCalcRepo, payrollPaySlipRepo)

	salaryUc := payrollUc.NewSalaryUsecase(payrollSalaryRepo, payrollEmployeeFetcher)
	compUc := payrollUc.NewCompensationUsecase(payrollCompItemRepo, payrollEmpCompRepo, payrollEmployeeFetcher)
	benefitUc := payrollUc.NewBenefitUsecase(payrollBenefitTypeRepo, payrollEmpBenefitRepo, payrollEmployeeFetcher)
	deductionUc := payrollUc.NewDeductionUsecase(payrollDeductionTypeRepo, payrollEmpDeductionRepo, payrollEmployeeFetcher)
	periodUc := payrollUc.NewPeriodUsecase(payrollPeriodRepo, payrollPaySlipRepo, payrollEmployeeFetcher)
	procUc := payrollUc.NewProcessorUsecase(payrollPeriodRepo, payrollProc)
	payrollOverviewUc := payrollUc.NewOverviewUsecase(payrollPeriodRepo, payrollSalaryRepo, payrollOverviewRepo, payrollDeductionTypeRepo, payrollCalcRepo, photoResolver)
	setupUc := payrollUc.NewSetupUsecase(db, payrollEmployeeFetcher)
	manualPayslipUc := payrollUc.NewManualPaySlipUsecase(payrollPeriodRepo, payrollPaySlipRepo, payrollCompItemRepo, payrollDeductionTypeRepo)
	payslipEmpFetcher := payrollAdapter.NewPayslipEmployeeFetcherAdapter(emplRepo, designationRepo)
	pdfRenderer := payrollAdapter.NewChromedpRenderer()
	companySettings := payrollAdapter.NewCompanySettingsProviderAdapter(settingUC)
	renderUc := payrollUc.NewRenderUsecase(payrollPeriodRepo, payrollPaySlipRepo, payslipEmpFetcher, pdfRenderer, companySettings)

	payrollHandler := payrollDelivery.NewPayrollHandler(
		salaryUc, compUc, benefitUc, deductionUc,
		periodUc, procUc, payrollOverviewUc, setupUc,
		manualPayslipUc, renderUc, payrollAuditLog, notifAdapter,
		payrollEmployeeFetcher, photoResolver,
	)
	payrollHandler.RegisterRoutes(api, authmw, adminmw)

	// Contract (after payroll for salary repo)
	contractDbRepo := contractRepo.NewPostgresContractRepo(db)
	contractSigningRepo := contractRepo.NewPostgresSigningRepo(db)
	contractDocumentRepo := contractRepo.NewPostgresDocumentRepo(db)
	contractNumGen := numbergen.New(contractDbRepo, "2001")
	contractEmpFetcher := contractAdapter.NewEmployeeFetcherAdapter(emplRepo, photoResolver)
	contractDesFetcher := contractAdapter.NewDesignationFetcherAdapter(designationRepo)
	contractSalFetcher := contractAdapter.NewSalaryFetcherAdapter(payrollSalaryRepo)
	contractTmplRepo := contractRepo.NewPostgresTemplateRepo(db)
	contractUC := contractUc.NewContractUsecase(contractTmplRepo, contractDbRepo, contractSigningRepo, contractNumGen, contractEmpFetcher, contractDesFetcher, contractSalFetcher)
	shiftTimeFetcher := contractAdapter.NewShiftTimeFetcherAdapter(workPatternRepo, ewpRepo)
	contractRenderUC := contractUc.NewRenderUsecase(contractDbRepo, contractSigningRepo, contractEmpFetcher, shiftTimeFetcher, contractAdapter.NewChromedpRenderer())
	contractDocAdapter := contractAdapter.NewDocumentMetadataAdapter(docRepo)
	contractObjStore := contractAdapter.NewObjectStorageAdapter(r2Storage)
	contractDocUC := contractUc.NewDocumentUsecase(contractRenderUC, contractObjStore, contractDocAdapter, contractDocumentRepo)
	contractEmpActivator := contractAdapter.NewEmployeeActivatorAdapter(emplRepo)
	contractWPAssigner := contractAdapter.NewWorkPatternAssignerAdapter(workPatternRepo, ewpRepo)
	contractUserActivator := contractAdapter.NewUserActivatorAdapter(userRepo)
	contractEmpTerminator := contractAdapter.NewTerminateEmployeeAdapter(emplRepo, userRepo)
	contractSignUC := contractUc.NewSigningUsecase(db, contractDbRepo, contractSigningRepo, contractDocUC, contractEmpFetcher, contractEmpActivator, contractWPAssigner, contractUserActivator, contractEmpTerminator)
	contractTerminateUC := contractUc.NewTerminationUsecase(contractDbRepo, contractEmpTerminator, userRepo, sessionRepo, deviceRepo, ewpRepo, overrideRepo)
	emplUC.SetContractFetcher(emplAdapter.NewCurrentContractFetcherAdapter(contractDbRepo))

	// Compile-time checks: repos satisfy contract interfaces
	var _ contractUc.ContractFinder = contractDbRepo
	var _ contractUc.EmployeeTerminator = contractEmpTerminator
	var _ contractUc.UserDeactivator = userRepo
	var _ contractUc.SessionRevoker = sessionRepo
	var _ contractUc.DeviceRevoker = deviceRepo
	var _ contractUc.WorkPatternDeactivator = ewpRepo
	var _ contractUc.UserActivator = contractUserActivator
	var _ contractUc.EmployeeUserIDFinder = contractEmpTerminator
	var _ contractUc.ScheduleOverrideDeleter = overrideRepo
	contractHandler := contractDelivery.NewContractHandler(contractUC, contractSignUC, contractRenderUC, contractDocUC, contractTerminateUC, contractAuditLog, notifAdapter)
	contractHandler.RegisterRoutes(api, authmw, adminmw)

	// Contract expiry scheduler
	contractExpiryProcessor := contractProcessor.NewExpiryProcessor(contractDbRepo, emplRepo)
	contractExpiryScheduler := contractProcessor.NewScheduler(db, contractExpiryProcessor, loc)
	contractExpiryScheduler.Start(context.Background())
	defer contractExpiryScheduler.Stop()

	// Storage (handlers only — infra already initialized above)
	storageUC := storageUc.NewStorageUsecase(docRepo, r2Storage, log, cfg.Storage.S3.MaxUploadSize)
	storageHandler := storageDelivery.New(storageUC, urlResolver, storageAuditLog)
	storageHandler.RegisterRoutes(api, authmw)

	// Dashboard
	dashboardRepo := dashboardRepo.NewPostgresDashboardRepo(db)
	dashboardUC := dashboardUc.NewDashboardUsecase(dashboardRepo)
	dashboardHandler := dashboardDelivery.NewDashboardHandler(dashboardUC)
	dashboardHandler.RegisterRoutes(api, authmw)

	if err := server.Listen(app, &cfg.Server, log); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
