package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	logger "github.com/selfshop-dev/lib-logger"
)

func TestLevelManager_Level(t *testing.T) {
	t.Parallel()

	lm := logger.NewLevelManager(zap.InfoLevel)
	assert.Equal(t, zapcore.InfoLevel, lm.Level())
}

func TestLevelManager_SetLevel(t *testing.T) {
	t.Parallel()

	lm := logger.NewLevelManager(zap.InfoLevel)
	lm.SetLevel(zap.DebugLevel)
	assert.Equal(t, zapcore.DebugLevel, lm.Level())
}

func TestLevelManager_Enabler(t *testing.T) {
	t.Parallel()

	lm := logger.NewLevelManager(zap.WarnLevel)
	assert.False(t, lm.Enabler().Enabled(zapcore.InfoLevel))
	assert.True(t, lm.Enabler().Enabled(zapcore.WarnLevel))
	assert.True(t, lm.Enabler().Enabled(zapcore.ErrorLevel))
}

func TestLevelHandler_GET(t *testing.T) {
	t.Parallel()

	lm := logger.NewLevelManager(zap.InfoLevel)
	l := zaptest.NewLogger(t)
	srv := httptest.NewServer(logger.LevelHandler(lm, l))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var body struct {
		Level zapcore.Level `json:"level"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, zapcore.InfoLevel, body.Level)
}

func TestLevelHandler_PUT(t *testing.T) {
	t.Parallel()

	lm := logger.NewLevelManager(zap.InfoLevel)
	l := zaptest.NewLogger(t)
	srv := httptest.NewServer(logger.LevelHandler(lm, l))
	t.Cleanup(srv.Close)

	body, _ := json.Marshal(map[string]string{"level": "debug"})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, srv.URL, bytes.NewReader(body))
	require.NoError(t, err)

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, zapcore.DebugLevel, lm.Level())
}

func TestLevelHandler_PUT_SameLevel(t *testing.T) {
	t.Parallel()

	lm := logger.NewLevelManager(zap.InfoLevel)
	l := zaptest.NewLogger(t)
	srv := httptest.NewServer(logger.LevelHandler(lm, l))
	t.Cleanup(srv.Close)

	body, _ := json.Marshal(map[string]string{"level": "info"})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, srv.URL, bytes.NewReader(body))
	require.NoError(t, err)

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, zapcore.InfoLevel, lm.Level())
}

func TestLevelHandler_PUT_InvalidBody(t *testing.T) {
	t.Parallel()

	lm := logger.NewLevelManager(zap.InfoLevel)
	l := zaptest.NewLogger(t)
	srv := httptest.NewServer(logger.LevelHandler(lm, l))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, srv.URL, bytes.NewBufferString("not json"))
	require.NoError(t, err)

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLevelHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	lm := logger.NewLevelManager(zap.InfoLevel)
	l := zaptest.NewLogger(t)
	srv := httptest.NewServer(logger.LevelHandler(lm, l))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, srv.URL, nil)
	require.NoError(t, err)

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
