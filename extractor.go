package logger

import (
	"context"

	"go.uber.org/zap"
)

// FieldExtractFunc is a function type that implements [FieldExtractor].
type FieldExtractFunc func(ctx context.Context) []zap.Field

// FieldExtractor extracts zap fields from a context.
type FieldExtractor interface {
	Extract(ctx context.Context) []zap.Field
}

// Compile-time check: [FieldExtractFunc] implements [FieldExtractor].
var _ FieldExtractor = (FieldExtractFunc)(nil)

// Extract calls fn with ctx and returns the resulting fields.
func (fn FieldExtractFunc) Extract(ctx context.Context) []zap.Field { return fn(ctx) }

// MergeExtractors returns a [FieldExtractor] that runs each extractor in fes
// and concatenates the results into a single slice.
func MergeExtractors(fes ...FieldExtractor) FieldExtractor {
	return FieldExtractFunc(func(ctx context.Context) []zap.Field {
		var fs []zap.Field
		for _, ex := range fes {
			fs = append(fs, ex.Extract(ctx)...)
		}
		return fs
	})
}
