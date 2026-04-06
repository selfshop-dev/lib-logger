package logger_test

import (
	"context"
	"io"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	logger "github.com/selfshop-dev/lib-logger"
)

func BenchmarkLogger_Info(b *testing.B) {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, err := logger.New(cfg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		l.Info("benchmark message", zap.String("key", "value"))
	}
}

func BenchmarkLogger_Info_Disabled(b *testing.B) {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	cfg.Level = logger.NewLevelManager(zapcore.ErrorLevel)
	l, _ := logger.New(cfg)
	b.ResetTimer()
	for b.Loop() {
		l.Info("benchmark", zap.String("key", "value"))
	}
}

func BenchmarkLogger_With(b *testing.B) {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, err := logger.New(cfg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		_ = l.With(zap.String("request_id", "req-1"), zap.String("trace_id", "trace-1"))
	}
}

func BenchmarkLogger_WithContext(b *testing.B) {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, err := logger.New(cfg)
	if err != nil {
		b.Fatal(err)
	}

	ex := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
		return []zap.Field{zap.String("trace_id", "t"), zap.String("user_id", "u")}
	})
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_ = l.WithContext(ctx, ex)
	}
}

func BenchmarkLogger_Unsampled_Warn(b *testing.B) {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, _ := logger.New(cfg)
	b.ResetTimer()
	for b.Loop() {
		l.Unsampled().Warn("benchmark", zap.String("key", "value"))
	}
}
