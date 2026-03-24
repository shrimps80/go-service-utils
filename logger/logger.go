package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config represents the configuration for the logger
type Config struct {
	Filename   string
	MaxSize    int // megabytes
	MaxBackups int
	MaxAge     int // days
	Compress   bool
	Level      string
}

// Logger is a wrapper around zap.Logger to encapsulate it
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new zap logger with rotation support
func NewLogger(cfg *Config) (*Logger, error) {
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
	zapLogger := zap.New(core, zap.AddCaller())
	return &Logger{Logger: zapLogger}, nil
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close closes the logger and syncs all logs
func (l *Logger) Close() error {
	return l.Logger.Sync()
}

// Debug logs a debug message with optional key-value pairs
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	if l.Logger == nil {
		return
	}
	l.Logger.Debug(msg, toZapFields(keysAndValues...)...)
}

// Info logs an info message with optional key-value pairs
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	if l.Logger == nil {
		return
	}
	l.Logger.Info(msg, toZapFields(keysAndValues...)...)
}

// Warn logs a warning message with optional key-value pairs
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	if l.Logger == nil {
		return
	}
	l.Logger.Warn(msg, toZapFields(keysAndValues...)...)
}

// Error logs an error message with optional key-value pairs
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	if l.Logger == nil {
		return
	}
	l.Logger.Error(msg, toZapFields(keysAndValues...)...)
}

// Fatal logs a fatal message with optional key-value pairs and then calls os.Exit(1)
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	if l.Logger == nil {
		return
	}
	l.Logger.Fatal(msg, toZapFields(keysAndValues...)...)
}

// With returns a new logger with the given key-value pairs added to all log messages
func (l *Logger) With(keysAndValues ...interface{}) *Logger {
	if l.Logger == nil {
		return l
	}
	return &Logger{Logger: l.Logger.With(toZapFields(keysAndValues...)...)}
}

// Named adds a sub-logger with a name
func (l *Logger) Named(name string) *Logger {
	if l.Logger == nil {
		return l
	}
	return &Logger{Logger: l.Logger.Named(name)}
}

// toZapFields converts key-value pairs to zap fields
func toZapFields(keysAndValues ...interface{}) []zap.Field {
	if len(keysAndValues)%2 != 0 {
		// If odd number of arguments, ignore the last one
		keysAndValues = keysAndValues[:len(keysAndValues)-1]
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key := keysAndValues[i]
		value := keysAndValues[i+1]

		// Convert key to string
		keyStr, ok := key.(string)
		if !ok {
			continue
		}

		// Add field based on value type
		switch v := value.(type) {
		case string:
			fields = append(fields, zap.String(keyStr, v))
		case int:
			fields = append(fields, zap.Int(keyStr, v))
		case int32:
			fields = append(fields, zap.Int32(keyStr, v))
		case int64:
			fields = append(fields, zap.Int64(keyStr, v))
		case float32:
			fields = append(fields, zap.Float32(keyStr, v))
		case float64:
			fields = append(fields, zap.Float64(keyStr, v))
		case bool:
			fields = append(fields, zap.Bool(keyStr, v))
		case error:
			fields = append(fields, zap.Error(v))
		default:
			fields = append(fields, zap.Any(keyStr, v))
		}
	}
	return fields
}
