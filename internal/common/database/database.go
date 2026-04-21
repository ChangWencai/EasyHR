package database

import (
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(cfg *config.DatabaseConfig) *gorm.DB {
	dsn := "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai"
	dsn = fmt.Sprintf(dsn, cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	gormLogLevel := logger.Warn
	if cfg.SQLLog {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			zap.NewStdLog(zap.L().Named("gorm")),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormLogLevel,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		zap.L().Fatal("failed to connect database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("failed to get sql.DB", zap.Error(err))
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

// InitAdmin 连接 postgres 系统数据库（用于执行 DROP DATABASE 等管理操作）
func InitAdmin(cfg *config.DatabaseConfig) *gorm.DB {
	dsn := "host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s TimeZone=Asia/Shanghai"
	dsn = fmt.Sprintf(dsn, cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode)

	gormLogLevel := logger.Warn
	if cfg.SQLLog {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			zap.NewStdLog(zap.L().Named("gorm")),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormLogLevel,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if err != nil {
		zap.L().Fatal("failed to connect admin database", zap.Error(err))
	}

	return db
}
