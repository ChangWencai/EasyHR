package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger      *zap.Logger
	SugarLogger *zap.SugaredLogger
)

func Init(mode string) {
	var config zap.Config
	if mode == "release" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	Logger, err = config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to init logger: " + err.Error() + "\n")
		os.Exit(1)
	}
	SugarLogger = Logger.Sugar()
}
