package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/tectix/hpcs/internal/config"
)

func New(cfg config.LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	switch cfg.Format {
	case "json":
		zapConfig = zap.NewProductionConfig()
	case "console":
		zapConfig = zap.NewDevelopmentConfig()
	default:
		return nil, fmt.Errorf("unsupported log format: %s", cfg.Format)
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	switch cfg.Output {
	case "stdout":
		zapConfig.OutputPaths = []string{"stdout"}
	case "stderr":
		zapConfig.OutputPaths = []string{"stderr"}
	case "file":
		if cfg.File == "" {
			return nil, fmt.Errorf("log file path not specified")
		}
		zapConfig.OutputPaths = []string{cfg.File}
	default:
		return nil, fmt.Errorf("unsupported log output: %s", cfg.Output)
	}

	zapConfig.ErrorOutputPaths = []string{"stderr"}

	if cfg.Format == "console" {
		zapConfig.Development = true
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

func NewDefault() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}
