package logger_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	logger "github.com/selfshop-dev/lib-logger"
)

func TestFieldExtractFunc_Extract(t *testing.T) {
	t.Parallel()

	field := zap.String("key", "value")
	fn := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
		return []zap.Field{field}
	})

	got := fn.Extract(context.Background())
	if diff := cmp.Diff([]zap.Field{field}, got); diff != "" {
		t.Errorf("FieldExtractFunc.Extract() mismatch (-want +got):\n%s", diff)
	}
}

func TestFieldExtractFunc_Extract_Nil(t *testing.T) {
	t.Parallel()

	fn := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
		return nil
	})

	got := fn.Extract(context.Background())
	assert.Nil(t, got)
}

func TestMergeExtractors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		extractors []logger.FieldExtractor
		want       []zap.Field
	}{
		{
			name:       "no extractors returns nil",
			extractors: nil,
			want:       nil,
		},
		{
			name: "single extractor",
			extractors: []logger.FieldExtractor{
				logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
					return []zap.Field{zap.String("a", "1")}
				}),
			},
			want: []zap.Field{zap.String("a", "1")},
		},
		{
			name: "multiple extractors concatenated in order",
			extractors: []logger.FieldExtractor{
				logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
					return []zap.Field{zap.String("a", "1")}
				}),
				logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
					return []zap.Field{zap.String("b", "2")}
				}),
			},
			want: []zap.Field{zap.String("a", "1"), zap.String("b", "2")},
		},
		{
			name: "extractor returning nil is skipped",
			extractors: []logger.FieldExtractor{
				logger.FieldExtractFunc(func(_ context.Context) []zap.Field { return nil }),
				logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
					return []zap.Field{zap.String("b", "2")}
				}),
			},
			want: []zap.Field{zap.String("b", "2")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := logger.MergeExtractors(tc.extractors...).Extract(context.Background())
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("MergeExtractors().Extract() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
