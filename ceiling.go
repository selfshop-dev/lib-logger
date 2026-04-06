package logger

import "go.uber.org/zap/zapcore"

// ceilingCore is a [zapcore.Core] that passes only messages below the ceiling
// level — messages at or above the ceiling are handled by the critical core.
// This creates a clean separation between the sampled and unsampled streams.
type ceilingCore struct {
	zapcore.Core
	ceil zapcore.LevelEnabler
}

func (c *ceilingCore) Enabled(l zapcore.Level) bool {
	return !c.ceil.Enabled(l) && c.Core.Enabled(l)
}

func (c *ceilingCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(e.Level) {
		return c.Core.Check(e, ce)
	}
	return ce
}

func (c *ceilingCore) With(fs []zapcore.Field) zapcore.Core {
	return &ceilingCore{
		Core: c.Core.With(fs),
		ceil: c.ceil,
	}
}
