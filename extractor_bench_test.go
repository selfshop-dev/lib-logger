package logger_test

import (
	"context"
	"testing"

	"go.uber.org/zap"

	logger "github.com/selfshop-dev/lib-logger"
)

func BenchmarkMergeExtractors_Extract(b *testing.B) {
	ex := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
		return []zap.Field{zap.String("k", "v")}
	})
	merged := logger.MergeExtractors(ex, ex, ex)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_ = merged.Extract(ctx)
	}
}
