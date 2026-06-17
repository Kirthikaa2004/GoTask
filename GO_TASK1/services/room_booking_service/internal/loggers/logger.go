package loggers

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.Logger
}

type LogConfig struct {
	Level       string
	LogDir      string
	FileName    string
	ServiceName string
}


func NewLogger(cfg LogConfig) *Logger {
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create log directory %s: %v", cfg.LogDir, err))
	}

	filePath := filepath.Join(cfg.LogDir, cfg.FileName)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file %s: %v", filePath, err))
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		NameKey:      "service",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	level := parseLevel(cfg.Level)
	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(file), level),
	)

	return &Logger{zap: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Named(cfg.ServiceName)}
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)

}
func (l *Logger) Error(msg string, fields ...zap.Field) {

	l.zap.Error(msg, fields...)

}
func (l *Logger) Warn(msg string, fields ...zap.Field) {

	l.zap.Warn(msg, fields...)

}
func (l *Logger) Fatal(msg string, fields ...zap.Field) {

	l.zap.Fatal(msg, fields...)
}
func (l *Logger) Sync() {
	_ = l.zap.Sync()
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
