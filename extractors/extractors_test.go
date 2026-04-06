// extractors/extractors_test.go
package extractors_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/selfshop-dev/lib-logger/extractors"
)

// field returns the first zap.Field extracted from ctx by ex, or a zero value.
func field(ctx context.Context, ex interface {
	Extract(context.Context) []zap.Field
},
) (zap.Field, bool) {
	fields := ex.Extract(ctx)
	if len(fields) == 0 {
		return zap.Field{}, false
	}
	return fields[0], true
}

func TestWithTraceID(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithTraceID(context.Background(), "trace-1")
	f, ok := field(ctx, extractors.TraceID())
	assert.True(t, ok)
	assert.Equal(t, "trace_id", f.Key)
	assert.Equal(t, "trace-1", f.String)
}

func TestWithSpanID(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithSpanID(context.Background(), "span-1")
	f, ok := field(ctx, extractors.SpanID())
	assert.True(t, ok)
	assert.Equal(t, "span_id", f.Key)
}

func TestWithRequestID(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithRequestID(context.Background(), "req-1")
	f, ok := field(ctx, extractors.RequestID())
	assert.True(t, ok)
	assert.Equal(t, "request_id", f.Key)
}

func TestWithCorrelationID(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithCorrelationID(context.Background(), "corr-1")
	f, ok := field(ctx, extractors.CorrelationID())
	assert.True(t, ok)
	assert.Equal(t, "correlation_id", f.Key)
}

func TestWithTenantID(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithTenantID(context.Background(), "tenant-1")
	f, ok := field(ctx, extractors.TenantID())
	assert.True(t, ok)
	assert.Equal(t, "tenant_id", f.Key)
}

func TestWithUserID(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithUserID(context.Background(), "user-1")
	f, ok := field(ctx, extractors.UserID())
	assert.True(t, ok)
	assert.Equal(t, "user_id", f.Key)
}

func TestExtractors_MissingKey_ReturnsNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ex   interface {
			Extract(context.Context) []zap.Field
		}
	}{
		{"TraceID", extractors.TraceID()},
		{"SpanID", extractors.SpanID()},
		{"RequestID", extractors.RequestID()},
		{"CorrelationID", extractors.CorrelationID()},
		{"TenantID", extractors.TenantID()},
		{"UserID", extractors.UserID()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.ex.Extract(context.Background())
			assert.Nil(t, got, "%s.Extract(empty ctx) = %v, want nil", tc.name, got)
		})
	}
}

func TestExtractors_EmptyValue_ReturnsNil(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithTraceID(context.Background(), "")
	got := extractors.TraceID().Extract(ctx)
	assert.Nil(t, got, "TraceID().Extract(ctx with empty value) = %v, want nil", got)
}

func TestTracing(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithTraceID(context.Background(), "trace-1")
	ctx = extractors.WithSpanID(ctx, "span-1")

	fields := extractors.Tracing().Extract(ctx)
	assert.Len(t, fields, 2)

	keys := map[string]bool{}
	for _, f := range fields {
		keys[f.Key] = true
	}
	assert.True(t, keys["trace_id"], "Tracing() missing trace_id")
	assert.True(t, keys["span_id"], "Tracing() missing span_id")
}

func TestFull(t *testing.T) {
	t.Parallel()

	ctx := extractors.WithTraceID(context.Background(), "t")
	ctx = extractors.WithSpanID(ctx, "s")
	ctx = extractors.WithCorrelationID(ctx, "c")
	ctx = extractors.WithRequestID(ctx, "r")
	ctx = extractors.WithUserID(ctx, "u")
	ctx = extractors.WithTenantID(ctx, "ten")

	fields := extractors.Full().Extract(ctx)
	assert.Len(t, fields, 6, "Full().Extract() = %d fields, want 6", len(fields))
}

func TestFull_PartialContext(t *testing.T) {
	t.Parallel()

	// Only two keys set — Full must return only the populated ones.
	ctx := extractors.WithTraceID(context.Background(), "t")
	ctx = extractors.WithUserID(ctx, "u")

	fields := extractors.Full().Extract(ctx)
	assert.Len(t, fields, 2, "Full().Extract(partial ctx) = %d fields, want 2", len(fields))
}
