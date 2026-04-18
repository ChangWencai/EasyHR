package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/attendance"
	"github.com/wencai/easyhr/internal/audit"
	"github.com/wencai/easyhr/internal/city"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/database"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/internal/dashboard"
	"github.com/wencai/easyhr/internal/department"
	"github.com/wencai/easyhr/internal/employee"
	"github.com/wencai/easyhr/internal/finance"
	"github.com/wencai/easyhr/internal/salary"
	"github.com/wencai/easyhr/internal/socialinsurance"
	"github.com/wencai/easyhr/internal/tax"
	"github.com/wencai/easyhr/internal/todo"
	"github.com/wencai/easyhr/internal/upload"
	"github.com/wencai/easyhr/internal/user"
	"github.com/wencai/easyhr/internal/wxmp"
	"github.com/wencai/easyhr/pkg/sms"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	cfg *config.Config
	db  *gorm.DB
	rdb *redis.Client
)

func initApp() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	var err error
	cfg, err = config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.InitWithConfig(&cfg.Log)
	logger.Logger.Info("logger initialized",
		zap.String("level", cfg.Log.Level),
		zap.String("path", cfg.Log.Path),
		zap.String("filename", cfg.Log.Filename),
	)

	db = database.Init(&cfg.Database)

	if err := db.AutoMigrate(
		&model.Organization{},
		&model.User{},
		&audit.AuditLog{},
		&employee.Employee{},
		&employee.Invitation{},
		&employee.Offboarding{},
		&employee.Contract{},
		&employee.Registration{},
		&department.Department{},
		&socialinsurance.SocialInsurancePolicy{},
		&socialinsurance.SocialInsuranceRecord{},
		&socialinsurance.ChangeHistory{},
			&socialinsurance.SIMonthlyPayment{},
		&tax.TaxBracket{},
		&tax.SpecialDeduction{},
		&tax.TaxRecord{},
		&tax.TaxDeclaration{},
		&tax.TaxReminder{},
		&salary.SalaryTemplateItem{},
		&salary.SalaryItem{},
		&salary.PayrollRecord{},
		&salary.PayrollItem{},
		&salary.PayrollSlip{},

		// 薪资增强模型
		&salary.SalaryAdjustment{},
		&salary.PerformanceCoefficient{},
		&salary.SickLeavePolicy{},
		&salary.SalarySlipSendLog{},

		// 财务模块模型
		&finance.Account{}, &finance.Period{}, &finance.Voucher{}, &finance.JournalEntry{},
		&finance.Invoice{}, &finance.ExpenseReimbursement{}, &finance.ReportSnapshot{},

		// 待办中心模型
		&todo.TodoItem{},
		&todo.CarouselItem{},
		&todo.TodoInvite{},
	); err != nil {
		logger.Logger.Fatal("auto migrate failed", zap.Error(err))
	}

	// 考勤模块 AutoMigrate
	if err := attendance.AutoMigrateTables(db); err != nil {
		logger.Logger.Fatal("attendance auto migrate failed", zap.Error(err))
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Logger.Warn("redis not available, continuing without cache", zap.Error(err))
	}
}

func main() {
	initApp()

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())

	smsClient, _ := sms.NewClient(sms.Config{
		AccessKeyID:     cfg.SMS.AccessKeyID,
		AccessKeySecret: cfg.SMS.AccessKeySecret,
		SignName:        cfg.SMS.SignName,
		TemplateCode:    cfg.SMS.TemplateCode,
		TestMode:        cfg.SMS.TestMode,
	})

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, rdb, smsClient, cfg.JWT, cfg.Crypto)
	userHandler := user.NewHandler(userSvc)

	// 员工模块依赖注入
	empRepo := employee.NewRepository(db)
	empSvc := employee.NewService(empRepo, cfg.Crypto)
	empHandler := employee.NewHandler(empSvc)

	// 邀请模块依赖注入
	invRepo := employee.NewInvitationRepository(db)
	invSvc := employee.NewInvitationService(invRepo, empRepo, cfg.Crypto)
	invHandler := employee.NewInvitationHandler(invSvc)

	// 员工信息登记模块依赖注入
	regRepo := employee.NewRegistrationRepository(db)
	regSvc := employee.NewRegistrationService(regRepo, empRepo, cfg.Crypto)
	regHandler := employee.NewRegistrationHandler(regSvc)

	// 部门模块依赖注入
	deptRepo := department.NewRepository(db)
	deptSvc := department.NewService(deptRepo, empRepo)
	deptHandler := department.NewDepartmentHandler(deptSvc)

	// 社保模块依赖注入（前置，供离职模块使用）
	siRepo := socialinsurance.NewRepository(db)
	siReminderRepo := socialinsurance.NewReminderRepository(db)
	empAdapter := socialinsurance.NewEmployeeAdapter(empRepo)
	siSvc := socialinsurance.NewService(siRepo, empAdapter, siReminderRepo)
	siPaymentRepo := socialinsurance.NewMonthlyPaymentRepository(db)
	siDashboardSvc := socialinsurance.NewSIDashboardService(db, siPaymentRepo, siRepo)
	siHandler := socialinsurance.NewHandler(siSvc, siDashboardSvc, siPaymentRepo)

	// 合同管理模块依赖注入（前置，供个税模块使用）
	contractRepo := employee.NewContractRepository(db)
	contractSvc := employee.NewContractService(contractRepo, empRepo, db, cfg.Crypto)
	contractHandler := employee.NewContractHandler(contractSvc)

	// 个税模块依赖注入
	taxRepo := tax.NewRepository(db)
	taxEmpAdapter := tax.NewEmployeeAdapter(contractRepo, empRepo)
	taxSIAdapter := tax.NewSocialInsuranceAdapter(siSvc)
	taxSvc := tax.NewService(taxRepo, taxEmpAdapter, taxSIAdapter)
	taxHandler := tax.NewHandler(taxSvc)

	// 离职管理模块依赖注入（集成社保停缴回调）
	obRepo := employee.NewOffboardingRepository(db)
	obSvc := employee.NewOffboardingService(obRepo, empRepo, siSvc)
	obHandler := employee.NewOffboardingHandler(obSvc)

	// 工资模块依赖注入
	salaryRepo := salary.NewRepository(db)
	salaryTemplateRepo := salary.NewSalaryTemplateRepository(db)
	salaryTaxAdapter := salary.NewTaxAdapter(taxSvc)
	salarySIAdapter := salary.NewSIAdapter(siSvc)
	salaryEmpAdapter := salary.NewEmployeeAdapter(empRepo, contractRepo)
	salaryAttendanceProvider := attendance.NewAttendanceProvider(db)
	salarySickLeavePolicySvc := salary.NewSickLeavePolicyService(db)
	salarySvc := salary.NewService(salaryRepo, salaryTemplateRepo, salaryTaxAdapter, salarySIAdapter, salaryEmpAdapter, salarySIAdapter, salaryAttendanceProvider, salarySickLeavePolicySvc, nil, cfg.Crypto)
	salaryDashboardSvc := salary.NewDashboardService(db)
	salaryHandler := salary.NewHandler(salarySvc, salaryDashboardSvc)
	taxUploadHandler := salary.NewTaxUploadHandler(salarySvc)

	// 工资条发送模块依赖注入（asynq）
	salarySlipSendSvc := salary.NewSlipSendService(salarySvc, fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port))
	salarySlipSendHandler := salary.NewSlipSendHandler(salarySlipSendSvc)
	salary.RegisterGlobalSlipSendService(salarySlipSendSvc)

	// 调薪模块依赖注入
	salaryAdjustmentRepo := salary.NewAdjustmentRepository(db)
	salaryAdjustmentSvc := salary.NewAdjustmentService(salaryAdjustmentRepo)
	salaryAdjustmentHandler := salary.NewAdjustmentHandler(salaryAdjustmentSvc)

	// 绩效系数模块依赖注入
	salaryPerformanceRepo := salary.NewPerformanceRepository(db)
	salaryPerformanceSvc := salary.NewPerformanceService(salaryPerformanceRepo)
	salaryPerformanceHandler := salary.NewPerformanceHandler(salaryPerformanceSvc)

	// 薪资列表模块依赖注入
	salaryListHandler := salary.NewSalaryListHandler(salarySvc)

	// 解锁模块依赖注入
	salaryUnlockSvc := salary.NewUnlockService(salaryRepo, fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port), db)
	salaryUnlockHandler := salary.NewUnlockHandler(salaryUnlockSvc, fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port))

	// 考勤模块依赖注入
	attendanceRepo := attendance.NewAttendanceRepository(db)
	attendanceSvc := attendance.NewAttendanceService(attendanceRepo)
	attendanceHandler := attendance.NewHandler(attendanceSvc)
	approvalSvc := attendance.NewApprovalService(attendanceRepo)
	attendanceHandler.SetApprovalService(approvalSvc)

	// 财务模块依赖注入
	accountRepo := finance.NewAccountRepository(db)
	periodRepo := finance.NewPeriodRepository(db)
	voucherRepo := finance.NewVoucherRepository(db)
	invoiceRepo := finance.NewInvoiceRepository(db)
	expenseRepo := finance.NewExpenseRepository(db)
	journalRepo := finance.NewJournalEntryRepository(db)
	snapshotRepo := finance.NewSnapshotRepository(db)

	accountSvc := finance.NewAccountServiceWithPeriod(accountRepo, periodRepo)
	voucherSvc := finance.NewVoucherService(voucherRepo, periodRepo, accountRepo)
	invoiceSvc := finance.NewInvoiceService(invoiceRepo, voucherRepo)
	expenseSvc := finance.NewExpenseService(expenseRepo, accountRepo, voucherSvc)

	reportSvc := finance.NewReportService(db, snapshotRepo, journalRepo, invoiceRepo, periodRepo)
	bookSvc := finance.NewBookService(db, accountRepo, journalRepo, periodRepo)
	periodSvc := finance.NewPeriodService(periodRepo, voucherRepo, journalRepo, reportSvc)

	accountHandler := finance.NewAccountHandler(accountSvc)
	voucherHandler := finance.NewVoucherHandler(voucherSvc)
	invoiceHandler := finance.NewInvoiceHandler(invoiceSvc, voucherSvc)
	expenseHandler := finance.NewExpenseHandler(expenseSvc, voucherSvc)
	bookHandler := finance.NewBookHandler(bookSvc)
	reportHandler := finance.NewReportHandler(reportSvc, periodSvc)
	financeHandler := finance.NewFinanceHandler(accountHandler, voucherHandler, invoiceHandler, expenseHandler, bookHandler, reportHandler)

	authMiddleware := middleware.Auth(cfg.JWT.Secret, rdb)

	v1 := r.Group("/api/v1")
	{
		userHandler.RegisterRoutes(v1, authMiddleware)
		empHandler.RegisterRoutes(v1, authMiddleware)
		deptHandler.RegisterRoutes(v1, authMiddleware)
		invHandler.RegisterRoutes(v1, authMiddleware)
		regHandler.RegisterRoutes(v1, authMiddleware)
		obHandler.RegisterRoutes(v1, authMiddleware)
		contractHandler.RegisterRoutes(v1, authMiddleware)
		siHandler.RegisterRoutes(v1, authMiddleware)
		taxHandler.RegisterRoutes(v1, authMiddleware)
		salaryHandler.RegisterRoutes(v1, authMiddleware)
		salarySlipSendHandler.RegisterRoutes(v1, authMiddleware)
		salaryAdjustmentHandler.RegisterRoutes(v1, authMiddleware)
		salaryPerformanceHandler.RegisterRoutes(v1, authMiddleware)
		salaryListHandler.RegisterRoutes(v1, authMiddleware)
		salaryUnlockHandler.RegisterRoutes(v1, authMiddleware)
		taxUploadHandler.RegisterRoutes(v1, authMiddleware)
		attendanceHandler.RegisterRoutes(v1, authMiddleware)
		financeHandler.RegisterRoutes(v1.Group(""), authMiddleware)
		city.NewHandler().RegisterRoutes(v1)
		audit.NewHandler(audit.NewRepository(db)).RegisterRoutes(v1)
		dashboard.RegisterRouter(v1.Group("/dashboard"), authMiddleware, db)
		todo.RegisterRouter(v1.Group(""), authMiddleware, db)
		upload.RegisterRouter(v1.Group(""), "./uploads", "")
		wxmp.RegisterWXMPRouter(v1, db, cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, rdb, cfg.Crypto.AESKey, userSvc)

		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Logger.Info("server starting", zap.String("addr", addr))

	// 社保缴费提醒定时任务
	siScheduler, err := socialinsurance.StartScheduler(rdb, siSvc)
	if err != nil {
		logger.Logger.Warn("social insurance scheduler start failed", zap.Error(err))
	}
	defer func() {
		if siScheduler != nil {
			siScheduler.Shutdown()
		}
	}()

	// 个税申报提醒定时任务
	taxScheduler, err := tax.StartScheduler(rdb, taxSvc)
	if err != nil {
		logger.Logger.Warn("tax scheduler start failed", zap.Error(err))
	}
	defer func() {
		if taxScheduler != nil {
			taxScheduler.Shutdown()
		}
	}()

		// 待办中心定时任务（urgency scan + carousel activation + 任务生成）
		todoRepo := todo.NewRepository(db)
		todoSched, todoSchedErr := todo.NewScheduler(todoRepo, rdb, nil).Start()
		if todoSchedErr != nil {
			logger.Logger.Warn("todo scheduler start failed", zap.Error(todoSchedErr))
		}
		defer func() {
			if todoSched != nil {
				todoSched.Shutdown()
			}
		}()

	// 初始化个税税率表种子数据（如果不存在）
	if err := taxSvc.SeedDefaultBrackets(time.Now().Year()); err != nil {
		logger.Logger.Warn("tax bracket seed failed", zap.Error(err))
	}

	// 初始化薪资模板种子数据（如果不存在）
	if err := salarySvc.SeedTemplateItems(); err != nil {
		logger.Logger.Warn("salary template seed failed", zap.Error(err))
	}

	// asynq Worker 启动（工资条批量发送队列）
	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)},
		asynq.Config{Concurrency: 10},
	)
	asynqMux := asynq.NewServeMux()
	asynqMux.HandleFunc(salary.TypeSlipSend, salary.HandleSlipSendTask)
	go func() {
		if err := asynqServer.Run(asynqMux); err != nil {
			logger.Logger.Error("asynq worker failed", zap.Error(err))
		}
	}()
	logger.Logger.Info("asynq worker started")

	// 初始化病假系数策略种子数据（北上广深）
	if err := salarySickLeavePolicySvc.SeedInitialPolicies(); err != nil {
		logger.Logger.Warn("sick leave policy seed failed", zap.Error(err))
	}

	if err := r.Run(addr); err != nil {
		logger.Logger.Fatal("server failed to start", zap.Error(err))
	}
}
