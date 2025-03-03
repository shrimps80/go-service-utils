package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config represents the configuration for the logger
type Config struct {
	Filename   string
	MaxSize    int  // megabytes
	MaxBackups int
	MaxAge     int  // days
	Compress   bool
	Level      string
}

// NewLogger creates a new zap logger with rotation support
func NewLogger(cfg *Config) (*zap.Logger, error) {
	// 设置日志级别
	var level zapcore.Level
	err := level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 配置lumberjack进行日志轮转
	writer := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// 创建zapcore
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(writer),
		level,
	)

	// 创建logger
	logger := zap.New(core, zap.AddCaller())
	return logger, nil
}