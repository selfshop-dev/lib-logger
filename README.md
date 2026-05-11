# lib-logger

[![CI](https://github.com/selfshop-dev/lib-logger/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-logger/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-logger/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-logger)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-logger)](https://goreportcard.com/report/github.com/selfshop-dev/lib-logger)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-logger)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-logger)](LICENSE)

Dual-core zap logger with runtime level control, context field injection, and built-in sampling. A project by [selfshop-dev](https://github.com/selfshop-dev).

### Installation

```bash
go get -u github.com/selfshop-dev/lib-logger
```

## Overview

`lib-logger` wraps two [go.uber.org/zap](https://github.com/uber-go/zap) cores connected via `zapcore.NewTee`. The sampled core handles high-frequency messages below the threshold level; the critical core is never sampled and is accessible directly via `Unsampled()` for audit logs. All `With*` methods return new instances — the receiver is never mutated.

```go
cfg := logger.DefaultConfig()
cfg.ServiceName = "orders"
cfg.Version     = "v1.2.3"
cfg.Environment = "prod"

l, err := logger.New(cfg)
if err != nil {
    return err
}
defer l.Sync()

l.Info("service started")
```

### Quick Start

```go
import logger "github.com/selfshop-dev/lib-logger"

cfg := logger.DefaultConfig()
cfg.ServiceName = "payments"

l, _ := logger.New(cfg)
defer l.Sync()

l.Info("ready", zap.String("port", "8080"))
```

## Config

`DefaultConfig` returns safe production defaults: threshold `InfoLevel`, sampler 100/100, sink — `os.Stdout`. Override only what differs.

```go
cfg := logger.DefaultConfig()
cfg.ServiceName       = "orders"
cfg.Version           = "v1.2.3"
cfg.Environment       = "prod"
cfg.InitialFields     = map[string]string{"commit": "e3f1a2b"}
cfg.SamplerFirst      = 200
cfg.SamplerThereafter = 50
cfg.Development       = false
```

Global fields `service`, `version`, `env` and fields from `InitialFields` are added to every log entry automatically.

## Context Enrichment

`WithContext` enriches the logger with fields extracted from the context via one or more `FieldExtractor` instances. Call it at the beginning of each request handler.

```go
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log := h.log.WithContext(r.Context(), extractors.Full())
    log.Info("handling request")
}
```

`MergeExtractors` combines multiple extractors into one:

```go
ex := logger.MergeExtractors(extractors.Tracing(), extractors.UserID())
```

A custom extractor is implemented via `FieldExtractFunc`:

```go
ex := logger.FieldExtractFunc(func(ctx context.Context) []zap.Field {
    return []zap.Field{zap.String("request_id", httpx.RequestID(ctx))}
})
```

## extractors

The `extractors` subpackage provides ready-to-use extractors for standard request fields.

```go
import "github.com/selfshop-dev/lib-logger/extractors"

// Write to context
ctx = extractors.WithTraceID(ctx, "trace-abc")
ctx = extractors.WithUserID(ctx, "usr-1")

// Extract when logging
log := l.WithContext(ctx, extractors.Full())
```

Available extractors: `TraceID()`, `SpanID()`, `RequestID()`, `CorrelationID()`, `UserID()`, `TenantID()`. Combined: `Tracing()` (trace + span), `Full()` (all six fields). If a key is missing or empty in the context, the field is silently skipped.

## Unsampled

`Unsampled()` returns the underlying critical `*zap.Logger` directly. Use it for audit logs, security events, and any messages that must never be dropped by the sampler.

```go
l.Unsampled().Warn("user password changed",
    zap.String("user_id", uid),
    zap.String("remote_addr", addr),
)
```

`With`, `Named`, and `WithContext` on a derived logger propagate both cores, so `Unsampled()` on a child logger carries the same fields:

```go
child := l.With(zap.String("request_id", id))
child.Unsampled().Warn("audit event") // contains request_id
```

## Runtime Level

`LevelManager` wraps `zap.AtomicLevel` and allows changing the log threshold without a restart. Register `LevelHandler` in your router for a GET/PUT endpoint:

```go
mux.Handle("/log/level", logger.LevelHandler(l.Level, l.Unsampled()))

// GET /log/level  → {"level":"info"}
// PUT /log/level  body: {"level":"debug"} → {"level":"debug"}
```

The unsampled logger is passed to `LevelHandler` so that the level-change warning is never subject to sampling.

## License

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)