package logger

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestIgnorableSyncErr_EINVAL(t *testing.T) {
	t.Parallel()

	err := &os.PathError{Op: "sync", Path: "/dev/stdout", Err: syscall.EINVAL}
	assert.True(t, ignorableSyncErr(err))
}

func TestIgnorableSyncErr_ENOTTY(t *testing.T) {
	t.Parallel()

	err := &os.PathError{Op: "sync", Path: "/dev/stdout", Err: syscall.ENOTTY}
	assert.True(t, ignorableSyncErr(err))
}

func TestIgnorableSyncErr_InvalidArgumentString(t *testing.T) {
	t.Parallel()

	err := &os.PathError{Op: "sync", Path: "/dev/stdout", Err: errors.New("invalid argument")}
	assert.True(t, ignorableSyncErr(err))
}

func TestIgnorableSyncErr_InappropriateIoctl(t *testing.T) {
	t.Parallel()

	err := &os.PathError{Op: "sync", Path: "/dev/stdout", Err: errors.New("inappropriate ioctl for device")}
	assert.True(t, ignorableSyncErr(err))
}

func TestIgnorableSyncErr_NotPathError(t *testing.T) {
	t.Parallel()

	assert.False(t, ignorableSyncErr(errors.New("disk full")))
}

func TestApplyGlobalFields_Empty(t *testing.T) {
	t.Parallel()

	fs := applyGlobalFields(Config{})
	assert.Empty(t, fs)
}

func TestApplyGlobalFields_AllFields(t *testing.T) {
	t.Parallel()

	c := Config{
		ServiceName:   "svc",
		Version:       "v1.0.0",
		Environment:   "prod",
		InitialFields: map[string]string{"commit": "abc123"},
	}
	fs := applyGlobalFields(c)

	keys := make(map[string]string, len(fs))
	for _, f := range fs {
		keys[f.Key] = f.String
	}

	assert.Equal(t, "svc", keys["service"])
	assert.Equal(t, "v1.0.0", keys["version"])
	assert.Equal(t, "prod", keys["env"])
	assert.Equal(t, "abc123", keys["commit"])
}

func TestApplyGlobalFields_InitialFieldsOnly(t *testing.T) {
	t.Parallel()

	c := Config{
		InitialFields: map[string]string{"build": "42"},
	}
	fs := applyGlobalFields(c)
	assert.Len(t, fs, 1)
	assert.Equal(t, zap.String("build", "42"), fs[0])
}

func TestNewEncoder_Production(t *testing.T) {
	t.Parallel()

	enc := newEncoder(false)
	assert.NotNil(t, enc)

	buf, err := enc.EncodeEntry(zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "hello",
	}, nil)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"msg":"hello"`)
}

func TestNewEncoder_Development(t *testing.T) {
	t.Parallel()

	enc := newEncoder(true)
	assert.NotNil(t, enc)

	buf, err := enc.EncodeEntry(zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "hello",
	}, nil)
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), `{"`)
}

func TestNewEncoder_Development_TimeFormat(t *testing.T) {
	t.Parallel()

	enc := newEncoder(true)

	buf, err := enc.EncodeEntry(zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "hello",
		Time:    time.Date(2024, 1, 2, 15, 4, 5, 123_000_000, time.UTC),
	}, nil)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "15:04:05.123")
}
