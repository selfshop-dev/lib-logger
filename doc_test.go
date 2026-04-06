package logger_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	logger "github.com/selfshop-dev/lib-logger"
)

func ExampleNew() {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	cfg.ServiceName = "orders"
	cfg.Version = "v1.0.0"

	l, err := logger.New(cfg)
	if err != nil {
		panic(err)
	}
	defer l.Sync()

	l.Info("service started")
	// Output:
}

func ExampleLogger_With() {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, _ := logger.New(cfg)

	req := l.With(
		zap.String("request_id", "req-abc"),
		zap.String("user_id", "usr-1"),
	)
	req.Info("handling request")
	req.Info("request complete")
	// Output:
}

func ExampleLogger_WithContext() {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, _ := logger.New(cfg)

	ex := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
		return []zap.Field{zap.String("trace_id", "trace-xyz")}
	})

	ctx := context.Background()
	req := l.WithContext(ctx, ex)
	req.Info("handling request")
	// Output:
}

func ExampleLogger_Named() {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, _ := logger.New(cfg)

	db := l.Named("db")
	db.Info("connection established")
	// Output:
}

func ExampleLogger_Unsampled() {
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, _ := logger.New(cfg)

	l.Unsampled().Warn("user password changed",
		zap.String("user_id", "usr-1"),
		zap.String("remote_addr", "1.2.3.4"),
	)
	// Output:
}

func ExampleLevelHandler() {
	lm := logger.NewLevelManager(zapcore.InfoLevel)
	cfg := logger.DefaultConfig()
	cfg.Level = lm
	cfg.Sink = zapcore.AddSync(io.Discard)
	l, _ := logger.New(cfg)

	mux := http.NewServeMux()
	mux.Handle("/log/level", logger.LevelHandler(lm, l.Unsampled()))
	// Output:
}

func ExampleMergeExtractors() {
	traceEx := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
		return []zap.Field{zap.String("trace_id", "t")}
	})
	userEx := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
		return []zap.Field{zap.String("user_id", "u")}
	})

	merged := logger.MergeExtractors(traceEx, userEx)
	fields := merged.Extract(context.Background())
	fmt.Println(len(fields))
	// Output:
	// 2
}

func ExampleDefaultConfig() {
	cfg := logger.DefaultConfig()
	cfg.ServiceName = "payments"
	cfg.Environment = "staging"

	var buf bytes.Buffer
	cfg.Sink = zapcore.AddSync(&buf)

	l, _ := logger.New(cfg)
	l.Info("ready")

	fmt.Println(buf.Len() > 0)
	// Output:
	// true
}
