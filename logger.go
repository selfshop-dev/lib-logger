package logger

import (
	"context"
	"errors"
	"os"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps [zap.Logger] with dual-core routing: a sampled path for high-volume
// traffic and a critical path for important events.
//
// Structure:
//  1. Sampled Core: Applies token-bucket sampling (SamplerFirst/SamplerThereafter)
//     to messages below Config.Level.
//  2. Critical Core: Never sampled. Handles messages at or above Config.Level.
//
// Logger is safe for concurrent use. Unlike your domain Error type,
// Logger's With* methods return new instances (copy-on-write) —
// they never mutate the receiver.
type Logger struct {
	*zap.Logger
	unsampled *zap.Logger

	// Level is the dynamic level controller for the unsampled critical core.
	// Use it to register [LevelHandler] in your router for runtime level changes.
	Level *LevelManager
}

// New constructs a [Logger] from c.
//
// The core is a tee of two paths:
//  1. Sampled core — receives messages below c.Level; applies token-bucket
//     sampling (SamplerFirst per second, then SamplerThereafter).
//  2. Critical core — receives messages at or above c.Level; never sampled.
//
// A [ceilingCore] wrapper on the sampled path prevents messages from
// appearing twice when both cores are enabled at the same level.
// If c.Sink is nil it defaults to locked os.Stdout.
func New(c Config) (*Logger, error) {
	enc := newEncoder(c.Development)

	if c.Sink == nil {
		c.Sink = zapcore.Lock(os.Stdout)
	}

	comm := zapcore.NewCore(enc, c.Sink, c.SampledLevel.Enabler())
	samp := zapcore.NewSamplerWithOptions(comm,
		time.Second,
		c.SamplerFirst, c.SamplerThereafter,
	)

	crit := zapcore.NewCore(enc, c.Sink, c.Level.Enabler())

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(applyGlobalFields(c)...),
	}

	merg := zap.New(
		zapcore.NewTee(&ceilingCore{
			Core: samp,
			ceil: c.Level.Enabler(),
		}, crit),
		opts...,
	)
	if c.Development {
		merg = merg.WithOptions(zap.Development())
	}

	unsamp := zap.New(crit, opts...)

	return &Logger{
		Logger:    merg,
		Level:     c.Level,
		unsampled: unsamp,
	}, nil
}

// Unsampled returns the underlying unsampled *[zap.Logger].
// Use for audit logs, security events, or any message that must never
// be dropped regardless of throughput.
func (l *Logger) Unsampled() *zap.Logger { return l.unsampled }

// Named returns a new Logger with the given name appended to the logger's name.
func (l *Logger) Named(s string) *Logger {
	return &Logger{
		Logger:    l.Logger.Named(s),
		Level:     l.Level,
		unsampled: l.unsampled.Named(s),
	}
}

// With returns a new Logger with the given fields added to every log entry.
func (l *Logger) With(fs ...zap.Field) *Logger {
	return &Logger{
		Logger:    l.Logger.With(fs...),
		Level:     l.Level,
		unsampled: l.unsampled.With(fs...),
	}
}

// WithContext returns a new Logger enriched with fields extracted from ctx
// by the provided extractors. Use at the start of a request handler to
// attach trace IDs, user IDs, and other request-scoped fields.
func (l *Logger) WithContext(ctx context.Context, fes ...FieldExtractor) *Logger {
	var fs []zap.Field
	for _, ex := range fes {
		fs = append(fs, ex.Extract(ctx)...)
	}
	return &Logger{
		Logger:    l.Logger.With(fs...),
		Level:     l.Level,
		unsampled: l.unsampled.With(fs...),
	}
}

// Sync flushes any buffered log entries. Call on shutdown.
// Errors from syncing stdout/stderr are ignored — they are expected on
// some platforms (EINVAL, ENOTTY).
func (l *Logger) Sync() error {
	if err := l.Logger.Sync(); !ignorableSyncErr(err) {
		return err
	}
	return nil
}

func newEncoder(dev bool) zapcore.Encoder {
	if dev {
		enc := zap.NewDevelopmentEncoderConfig()
		enc.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enc.EncodeTime = func(t time.Time, e zapcore.PrimitiveArrayEncoder) { e.AppendString(t.Format("15:04:05.000")) }
		return zapcore.NewConsoleEncoder(enc)
	}
	enc := zap.NewProductionEncoderConfig()
	enc.EncodeDuration = zapcore.MillisDurationEncoder
	enc.EncodeTime = zapcore.RFC3339TimeEncoder
	return zapcore.NewJSONEncoder(enc)
}

func applyGlobalFields(c Config) []zap.Field {
	var fs []zap.Field
	if c.ServiceName != "" {
		fs = append(fs, zap.String("service", c.ServiceName))
	}
	if c.Version != "" {
		fs = append(fs, zap.String("version", c.Version))
	}
	if c.Environment != "" {
		fs = append(fs, zap.String("env", c.Environment))
	}
	for k, v := range c.InitialFields {
		fs = append(fs, zap.String(k, v))
	}
	return fs
}

func ignorableSyncErr(err error) bool {
	pe, ok := errors.AsType[*os.PathError](err)
	if !ok {
		return false
	}
	msg := strings.ToLower(pe.Err.Error())
	return errors.Is(pe.Err, syscall.EINVAL) ||
		errors.Is(pe.Err, syscall.ENOTTY) ||
		strings.Contains(msg, "invalid argument") ||
		strings.Contains(msg, "inappropriate ioctl")
}
