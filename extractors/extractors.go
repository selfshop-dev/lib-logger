package extractors

import (
	"context"

	"go.uber.org/zap"

	logger "github.com/selfshop-dev/lib-logger"
)

type ctxKey string

// Context keys for log field values. Stored and retrieved by the With* helpers;
// passed to [logger.Logger.WithContext] via the extractor functions below.
const (
	TraceIDKey       ctxKey = "trace_id"
	SpanIDKey        ctxKey = "span_id"
	CorrelationIDKey ctxKey = "correlation_id"
	RequestIDKey     ctxKey = "request_id"
	UserIDKey        ctxKey = "user_id"
	TenantIDKey      ctxKey = "tenant_id"
)

// WithTraceID returns a copy of ctx with TraceIDKey set to v.
func WithTraceID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, TraceIDKey, v)
}

// WithSpanID returns a copy of ctx with SpanIDKey set to v.
func WithSpanID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, SpanIDKey, v)
}

// WithRequestID returns a copy of ctx with RequestIDKey set to v.
func WithRequestID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, RequestIDKey, v)
}

// WithCorrelationID returns a copy of ctx with CorrelationIDKey set to v.
func WithCorrelationID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, v)
}

// WithTenantID returns a copy of ctx with TenantIDKey set to v.
func WithTenantID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, TenantIDKey, v)
}

// WithUserID returns a copy of ctx with UserIDKey set to v.
func WithUserID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, UserIDKey, v)
}

// TraceID returns a [logger.FieldExtractor] that reads TraceIDKey from ctx.
func TraceID() logger.FieldExtractor { return extractOne(TraceIDKey) }

// SpanID returns a [logger.FieldExtractor] that reads SpanIDKey from ctx.
func SpanID() logger.FieldExtractor { return extractOne(SpanIDKey) }

// UserID returns a [logger.FieldExtractor] that reads UserIDKey from ctx.
func UserID() logger.FieldExtractor { return extractOne(UserIDKey) }

// CorrelationID returns a [logger.FieldExtractor] that reads CorrelationIDKey from ctx.
func CorrelationID() logger.FieldExtractor { return extractOne(CorrelationIDKey) }

// TenantID returns a [logger.FieldExtractor] that reads TenantIDKey from ctx.
func TenantID() logger.FieldExtractor { return extractOne(TenantIDKey) }

// RequestID returns a [logger.FieldExtractor] that reads RequestIDKey from ctx.
func RequestID() logger.FieldExtractor { return extractOne(RequestIDKey) }

func extractOne(k ctxKey) logger.FieldExtractFunc {
	return func(ctx context.Context) []zap.Field {
		if v, ok := ctx.Value(k).(string); ok &&
			v != "" {
			return []zap.Field{zap.String(string(k), v)}
		}
		return nil
	}
}

// Tracing returns a [logger.FieldExtractor] that reads TraceIDKey and SpanIDKey from ctx.
func Tracing() logger.FieldExtractor {
	return logger.MergeExtractors(TraceID(), SpanID())
}

// Full returns a [FieldExtractor] that reads all six standard keys from ctx.
//
// It merges Trace, Span, Correlation, Request, User, and Tenant IDs.
// If a key is missing or empty in the context, it is silently skipped.
func Full() logger.FieldExtractor {
	return logger.MergeExtractors(Tracing(), CorrelationID(), RequestID(), UserID(), TenantID())
}
