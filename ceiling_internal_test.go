package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestCeilingCore_Enabled_BelowCeiling(t *testing.T) {
	t.Parallel()

	core, _ := observer.New(zapcore.DebugLevel)
	ceil := zapcore.WarnLevel
	cc := &ceilingCore{Core: core, ceil: ceil}

	assert.True(t, cc.Enabled(zapcore.InfoLevel))
}

func TestCeilingCore_Enabled_AtCeiling(t *testing.T) {
	t.Parallel()

	core, _ := observer.New(zapcore.DebugLevel)
	ceil := zapcore.WarnLevel
	cc := &ceilingCore{Core: core, ceil: ceil}

	assert.False(t, cc.Enabled(zapcore.WarnLevel))
}

func TestCeilingCore_Check_PassesBelowCeiling(t *testing.T) {
	t.Parallel()

	core, logs := observer.New(zapcore.DebugLevel)
	cc := &ceilingCore{Core: core, ceil: zapcore.WarnLevel}

	entry := zapcore.Entry{Level: zapcore.InfoLevel, Message: "below"}
	ce := cc.Check(entry, nil)
	assert.NotNil(t, ce)
	_ = logs
}

func TestCeilingCore_Check_BlocksAtCeiling(t *testing.T) {
	t.Parallel()

	core, _ := observer.New(zapcore.DebugLevel)
	cc := &ceilingCore{Core: core, ceil: zapcore.WarnLevel}

	entry := zapcore.Entry{Level: zapcore.WarnLevel, Message: "at ceiling"}
	// ce passed in as nil — must be returned unchanged when Enabled is false.
	ce := cc.Check(entry, nil)
	assert.Nil(t, ce)
}

func TestCeilingCore_With(t *testing.T) {
	t.Parallel()

	core, _ := observer.New(zapcore.DebugLevel)
	cc := &ceilingCore{Core: core, ceil: zapcore.WarnLevel}

	child := cc.With([]zapcore.Field{{Key: "k", Type: zapcore.StringType, String: "v"}})
	_, ok := child.(*ceilingCore)
	assert.True(t, ok, "With() must return *ceilingCore")
}
