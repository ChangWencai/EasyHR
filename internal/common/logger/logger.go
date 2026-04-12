package logger

import (
	"os"
	"path/filepath"

	"github.com/wencai/easyhr/internal/common/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger      *zap.Logger
	SugarLogger *zap.SugaredLogger
)

// Init 初始化日志（简化版本，使用默认值）
func Init(mode string) {
	InitWithConfig(&config.LogConfig{
		Level:      "debug",
		Path:       "./logs",
		Filename:   "server.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     7,
		Compress:   true,
	})
}

// InitWithConfig 初始化日志（支持配置文件）
func InitWithConfig(cfg *config.LogConfig) {
	// 确保日志目录存在
	if err := os.MkdirAll(cfg.Path, 0755); err != nil {
		_, _ = os.Stderr.WriteString("failed to create log directory: " + err.Error() + "\n")
		os.Exit(1)
	}

	// 解析日志级别
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.DebugLevel
	}

	// 日志文件路径
	logFile := filepath.Join(cfg.Path, cfg.Filename)

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 同时输出到控制台和文件
	var cores []zapcore.Core

	// 控制台输出（开发模式使用彩色输出）
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	cores = append(cores, zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		level,
	))

	// 文件输出（使用 JSON 格式便于日志收集）
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	writer := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}
	cores = append(cores, zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(writer),
		level,
	))

	Logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
	zap.ReplaceGlobals(Logger)
	SugarLogger = Logger.Sugar()
}
