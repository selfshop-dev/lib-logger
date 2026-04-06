package extractors_test

import (
	"context"
	"testing"

	"github.com/selfshop-dev/lib-logger/extractors"
)

func BenchmarkFull_Extract(b *testing.B) {
	ctx := extractors.WithTraceID(context.Background(), "t")
	ctx = extractors.WithSpanID(ctx, "s")
	ctx = extractors.WithCorrelationID(ctx, "c")
	ctx = extractors.WithRequestID(ctx, "r")
	ctx = extractors.WithUserID(ctx, "u")
	ctx = extractors.WithTenantID(ctx, "ten")

	ex := extractors.Full()
	b.ResetTimer()
	for b.Loop() {
		_ = ex.Extract(ctx)
	}
}

func BenchmarkTracing_Extract(b *testing.B) {
	ctx := extractors.WithTraceID(context.Background(), "trace-1")
	ctx = extractors.WithSpanID(ctx, "span-1")

	ex := extractors.Tracing()
	b.ResetTimer()
	for b.Loop() {
		_ = ex.Extract(ctx)
	}
}
