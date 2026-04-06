package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/audit"
	"github.com/wencai/easyhr/internal/city"
	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/database"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/model"
	"github.com/wencai/easyhr/internal/user"
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

	logger.Init(cfg.Server.Mode)

	db = database.Init(&cfg.Database)

	if err := db.AutoMigrate(
		&model.Organization{},
		&model.User{},
		&audit.AuditLog{},
	); err != nil {
		logger.Logger.Fatal("auto migrate failed", zap.Error(err))
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := rdb.Ping(nil).Err(); err != nil {
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
	})

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, rdb, smsClient, cfg.JWT, cfg.Crypto)
	userHandler := user.NewHandler(userSvc)

	authMiddleware := middleware.Auth(cfg.JWT.Secret, rdb)

	v1 := r.Group("/api/v1")
	{
		userHandler.RegisterRoutes(v1, authMiddleware)
		city.NewHandler().RegisterRoutes(v1)
		audit.NewHandler(audit.NewRepository(db)).RegisterRoutes(v1)

		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Logger.Info("server starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Logger.Fatal("server failed to start", zap.Error(err))
	}
}
