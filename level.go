package logger

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LevelManager wraps [zap.AtomicLevel] and exposes it for runtime changes
// via [LevelHandler]. Pass it to [LevelHandler] to register a GET/PUT
// endpoint in your router.
type LevelManager struct {
	lvl zap.AtomicLevel
}

// NewLevelManager creates a LevelManager initialised to level l.
func NewLevelManager(l zapcore.Level) *LevelManager {
	return &LevelManager{
		lvl: zap.NewAtomicLevelAt(l),
	}
}

// Enabler returns the underlying [zapcore.LevelEnabler] for use in core construction.
func (lm *LevelManager) Enabler() zapcore.LevelEnabler { return lm.lvl }

// Level returns the current minimum log level.
func (lm *LevelManager) Level() zapcore.Level { return lm.lvl.Level() }

// SetLevel changes the minimum log level atomically.
// Safe for concurrent use.
func (lm *LevelManager) SetLevel(l zapcore.Level) { lm.lvl.SetLevel(l) }

// levelResponse is the JSON shape for GET and PUT /log/level.
type levelResponse struct {
	Level zapcore.Level `json:"level"`
}

// LevelHandler returns an [http.Handler] that exposes the log level over HTTP.
//
// GET /log/level → {"level":"info"}
// PUT /log/level → body: {"level":"debug"} → {"level":"debug"}
//
// Register in your router:
//
//	r.Handle(logger.LevelPath, logger.LevelHandler(l.Level, l.Unsampled()))
//
// The unsampled logger is used to emit a warning when the level changes —
// this message must never be sampled away.
func LevelHandler(lm *LevelManager, l *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetLevel(w, lm)
		case http.MethodPut:
			handlePutLevel(w, r, lm, l)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}

func handleGetLevel(w http.ResponseWriter, lm *LevelManager) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(levelResponse{lm.Level()}) //nolint:errcheck // write error after headers are sent is unrecoverable
}

func handlePutLevel(w http.ResponseWriter, r *http.Request, lm *LevelManager, l *zap.Logger) {
	var b levelResponse
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	prev := lm.Level()
	if b.Level == prev {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(levelResponse{prev}) //nolint:errcheck // write error after headers are sent is unrecoverable
		return
	}

	l.Warn("log level changed",
		zap.Stringer("from", prev),
		zap.Stringer("to", b.Level),
		zap.String("remote_addr", r.RemoteAddr),
	)

	lm.SetLevel(b.Level)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(levelResponse{b.Level}) //nolint:errcheck // write error after headers are sent is unrecoverable
}
