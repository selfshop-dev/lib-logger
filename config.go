package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds all parameters for constructing a [Logger].
// Use [DefaultConfig] as a starting point and override only what you need.
//
// Fields are validated during [New] call. Sink must not be nil —
// if not provided, [New] defaults to a locked os.Stdout.
//
// Level and SampledLevel are dynamic: changing their state via [LevelManager]
// immediately affects all loggers derived from this config.
type Config struct {
	// Sink is the write destination. Defaults to locked os.Stdout.
	// Must not be nil — use [DefaultConfig] to get a safe default.
	Sink zapcore.WriteSyncer

	// Level is the minimum severity for the unsampled critical core.
	// Messages at this level and above are always written.
	Level *LevelManager

	// SampledLevel is the minimum severity for the sampled core.
	// Typically set lower than Level (e.g. DebugLevel) to capture
	// a representative sample of high-frequency messages.
	SampledLevel *LevelManager

	// InitialFields are added as typed string fields to every log entry.
	// Use for build metadata: commit hash, build time, etc.
	InitialFields map[string]string

	// Version is added as a "version" field to every log entry.
	Version string

	// ServiceName is added as a "service" field to every log entry.
	ServiceName string

	// Environment is added as an "env" field to every log entry.
	Environment string

	// SamplerFirst is the number of messages per second passed through
	// before sampling kicks in.
	SamplerFirst int

	// SamplerThereafter is the number of messages kept per second
	// after the SamplerFirst threshold is exceeded.
	SamplerThereafter int

	// Development enables zap development mode: DPanic panics, stack
	// traces on warnings, and console-friendly output.
	Development bool
}

// DefaultConfig returns a Config with safe production-ready defaults:
// InfoLevel threshold, 100/100 sampler, stdout sink.
func DefaultConfig() Config {
	return Config{
		Level:             NewLevelManager(zap.InfoLevel),
		SampledLevel:      NewLevelManager(zap.DebugLevel),
		SamplerFirst:      100,
		SamplerThereafter: 100,
		Sink:              zapcore.Lock(os.Stdout),
	}
}
