package logger_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	logger "github.com/selfshop-dev/lib-logger"
)

// newBufferedLogger builds a Logger that writes into buf so tests can inspect output.
func newBufferedLogger(t *testing.T, buf *bytes.Buffer, level zapcore.Level) *logger.Logger {
	t.Helper()
	cfg := logger.DefaultConfig()
	cfg.Level = logger.NewLevelManager(level)
	cfg.SampledLevel = logger.NewLevelManager(level)
	cfg.Sink = zapcore.AddSync(buf)
	l, err := logger.New(cfg)
	require.NoError(t, err)
	return l
}

func TestNew_DefaultConfig(t *testing.T) {
	t.Parallel()

	l, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, l)
}

func TestNew_NilSink_DefaultsToStdout(t *testing.T) {
	t.Parallel()

	cfg := logger.DefaultConfig()
	cfg.Sink = nil

	l, err := logger.New(cfg)
	require.NoError(t, err)
	require.NotNil(t, l)
}

func TestNew_GlobalFields_AppearsInOutput(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cfg := logger.DefaultConfig()
	cfg.Sink = zapcore.AddSync(&buf)
	cfg.ServiceName = "svc"
	cfg.Version = "v1.2.3"
	cfg.Environment = "prod"
	cfg.InitialFields = map[string]string{"commit": "abc123"}

	l, err := logger.New(cfg)
	require.NoError(t, err)
	l.Info("probe")

	out := buf.String()
	assert.Contains(t, out, `"service":"svc"`)
	assert.Contains(t, out, `"version":"v1.2.3"`)
	assert.Contains(t, out, `"env":"prod"`)
	assert.Contains(t, out, `"commit":"abc123"`)
}

func TestNew_DevelopmentMode(t *testing.T) {
	t.Parallel()

	cfg := logger.DefaultConfig()
	cfg.Development = true

	l, err := logger.New(cfg)
	require.NoError(t, err)
	require.NotNil(t, l)
}

func TestLogger_Named(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	l := newBufferedLogger(t, &buf, zapcore.InfoLevel)

	named := l.Named("subsystem")
	require.NotNil(t, named)
	assert.NotSame(t, l, named)

	named.Info("probe")
	assert.Contains(t, buf.String(), "subsystem")
}

func TestLogger_With(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	l := newBufferedLogger(t, &buf, zapcore.InfoLevel)

	enriched := l.With(zap.String("request_id", "req-1"))
	require.NotNil(t, enriched)
	assert.NotSame(t, l, enriched)

	enriched.Info("probe")
	assert.Contains(t, buf.String(), `"request_id":"req-1"`)
}

func TestLogger_WithContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	l := newBufferedLogger(t, &buf, zapcore.InfoLevel)

	ex := logger.FieldExtractFunc(func(_ context.Context) []zap.Field {
		return []zap.Field{zap.String("trace_id", "abc")}
	})

	enriched := l.WithContext(context.Background(), ex)
	require.NotNil(t, enriched)
	assert.NotSame(t, l, enriched)

	enriched.Info("probe")
	assert.Contains(t, buf.String(), `"trace_id":"abc"`)
}

func TestLogger_WithContext_NoExtractors(t *testing.T) {
	t.Parallel()

	l, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)

	enriched := l.WithContext(context.Background())
	assert.NotNil(t, enriched)
}

func TestLogger_Unsampled_NotNil(t *testing.T) {
	t.Parallel()

	l, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)

	assert.NotNil(t, l.Unsampled())
}

func TestLogger_Unsampled_WritesWhenSampledWouldDrop(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cfg := logger.DefaultConfig()
	cfg.Level = logger.NewLevelManager(zapcore.WarnLevel)
	cfg.SampledLevel = logger.NewLevelManager(zapcore.WarnLevel)
	cfg.SamplerFirst = 0
	cfg.SamplerThereafter = 0
	cfg.Sink = zapcore.AddSync(&buf)

	l, err := logger.New(cfg)
	require.NoError(t, err)

	l.Unsampled().Warn("must appear")
	assert.Contains(t, buf.String(), "must appear", "Unsampled().Warn() must appear in output")
}

func TestLogger_Sync(t *testing.T) {
	t.Parallel()

	l, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)

	// Sync on stdout may return ENOTTY on some platforms — must be swallowed.
	assert.NoError(t, l.Sync())
}

func TestLogger_Level_RuntimeChange(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	l := newBufferedLogger(t, &buf, zapcore.WarnLevel)

	l.Info("before") // below threshold — must not appear
	assert.Empty(t, buf.String(), "Info logged below threshold")

	l.Level.SetLevel(zapcore.InfoLevel)
	l.Info("after") // now at threshold — must appear
	assert.Contains(t, buf.String(), "after")
}

// failSyncer is a WriteSyncer whose Sync always returns err.
type failSyncer struct{ err error }

func (f *failSyncer) Write(p []byte) (int, error) { return len(p), nil }
func (f *failSyncer) Sync() error                 { return f.err }

func TestLogger_Sync_NonIgnorableError(t *testing.T) {
	t.Parallel()

	// errSyncer returns a non-ignorable error from Sync.
	errSyncer := &failSyncer{err: errors.New("disk full")}
	cfg := logger.DefaultConfig()
	cfg.Sink = errSyncer
	l, err := logger.New(cfg)
	require.NoError(t, err)

	assert.ErrorContains(t, l.Sync(), "disk full")
}

func TestLogger_Sync_PathErrorNotIgnorable(t *testing.T) {
	t.Parallel()

	errSyncer := &failSyncer{err: &os.PathError{
		Op:   "sync",
		Path: "/dev/sda",
		Err:  syscall.EIO,
	}}
	cfg := logger.DefaultConfig()
	cfg.Sink = errSyncer
	l, err := logger.New(cfg)
	require.NoError(t, err)
	assert.Error(t, l.Sync())
}
