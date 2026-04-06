package logger_test

import (
	"context"
	"testing"

	"go.uber.org/zap"

	logger "github.com/selfshop-dev/lib-logger"
)

func FuzzFieldExtractFunc(f *testing.F) {
	f.Add("key", "value")
	f.Add("", "")
	f.Add("trace_id", "abc-123")

	f.Fuzz(func(t *testing.T, key, val string) {
		fn := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
			return []zap.Field{zap.String(key, val)}
		})
		got := fn.Extract(context.Background())
		if len(got) != 1 {
			t.Errorf("FieldExtractFunc.Extract() = %d fields, want 1", len(got))
		}
	})
}
