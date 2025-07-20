package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	Key       contextKey = "logger"
	RequestID contextKey = "request_id"
)

type Logger struct {
	l *zap.Logger
}

func New(env string, level string) (*Logger, error) {
	var config zap.Config

	switch env {
	case "development", "dev":
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case "testing", "test":
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	default:
		config = zap.NewProductionConfig()
	}

	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{l: logger}, nil
}

func (l *Logger) log(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	if ctx != nil && ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}

	switch level {
	case zapcore.DebugLevel:
		l.l.Debug(msg, fields...)
	case zapcore.InfoLevel:
		l.l.Info(msg, fields...)
	case zapcore.WarnLevel:
		l.l.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		l.l.Error(msg, fields...)
	case zapcore.FatalLevel:
		l.l.Fatal(msg, fields...)
	}
}

func GetLoggerFromCtx(ctx context.Context) *Logger {
	return ctx.Value(Key).(*Logger)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.DebugLevel, msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.InfoLevel, msg, fields...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.WarnLevel, msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.ErrorLevel, msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.FatalLevel, msg, fields...)
}

func (l *Logger) Sync() error {
	return l.l.Sync()
}
