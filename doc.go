// Package logger provides a dual-core zap-based logger with runtime level
// control, request-scoped field injection, and built-in sampling.
//
// # Architecture
//
// Every [Logger] wraps two [go.uber.org/zap] cores wired together with
// [zapcore.NewTee]:
//
//  1. Sampled core — receives messages below [Config.Level]; applies a
//     token-bucket sampler (SamplerFirst messages per second pass through,
//     then SamplerThereafter). Backed by [ceilingCore] to prevent
//     duplicates when both cores share a threshold.
//  2. Critical core — receives messages at or above [Config.Level]; never
//     sampled. Also exposed via [Logger.Unsampled] for audit events.
//
// # Construction
//
// Use [DefaultConfig] as a starting point and override only what differs:
//
//	cfg := logger.DefaultConfig()
//	cfg.ServiceName   = "orders"
//	cfg.Version       = "v1.2.3"
//	cfg.Environment   = "prod"
//	cfg.InitialFields = map[string]string{"commit": "e3f1a2b"}
//
//	l, err := logger.New(cfg)
//	if err != nil {
//	    return err
//	}
//	defer l.Sync()
//
// # Request-scoped fields
//
// Attach fields from a context at the start of each request handler using
// [Logger.WithContext] and one or more [FieldExtractor] implementations.
// The [extractors] sub-package provides ready-made extractors for trace,
// span, request, correlation, user, and tenant IDs:
//
//	import "github.com/selfshop-dev/lib-logger/extractors"
//
//	func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	    log := h.log.WithContext(r.Context(), extractors.Full())
//	    log.Info("handling request")
//	}
//
// [MergeExtractors] combines multiple [FieldExtractor] values into one:
//
//	ex := logger.MergeExtractors(extractors.Tracing(), extractors.UserID())
//
// # Runtime level changes
//
// [Config.Level] is a [LevelManager] — a thin wrapper around
// [go.uber.org/zap.AtomicLevel]. Register [LevelHandler] in your router to
// expose a GET/PUT endpoint for live level changes without a restart:
//
//	mux.Handle("/log/level", logger.LevelHandler(l.Level, l.Unsampled()))
//
//	// GET /log/level → {"level":"info"}
//	// PUT /log/level body: {"level":"debug"} → {"level":"debug"}
//
// The unsampled logger is passed to [LevelHandler] so the level-change
// warning is never dropped by the sampler.
//
// # Unsampled logger
//
// [Logger.Unsampled] returns the underlying critical-core logger directly.
// Use it for audit logs, security events, or any message that must never
// be dropped regardless of throughput:
//
//	l.Unsampled().Warn("user password changed",
//	    zap.String("user_id", uid),
//	    zap.String("remote_addr", addr),
//	)
//
// # Copy-on-write
//
// [Logger.With], [Logger.WithContext], and [Logger.Named] all return new
// [Logger] instances — they never mutate the receiver. Both the sampled and
// unsampled cores are propagated to the child, so [Logger.Unsampled] on a
// derived logger carries the same fields:
//
//	child := l.With(zap.String("request_id", id))
//	child.Info("ok")                // sampled
//	child.Unsampled().Warn("audit") // never sampled, carries request_id
//
// # Sync
//
// Call [Logger.Sync] on shutdown to flush buffered entries. Sync swallows
// EINVAL and ENOTTY errors that some platforms return when syncing stdout:
//
//	defer l.Sync()
//
// # Concurrency
//
// [Logger] is safe for concurrent use after construction. [LevelManager.SetLevel]
// is atomic. The [MetaExtractor] passed to [LevelHandler] is called once per
// request and must itself be safe for concurrent use.
package logger
