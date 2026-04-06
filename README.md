# lib-logger

[![CI](https://github.com/selfshop-dev/lib-logger/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-logger/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-logger/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-logger)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-logger)](https://goreportcard.com/report/github.com/selfshop-dev/lib-logger)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-logger)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-logger)](LICENSE)

Dual-core zap-логгер с runtime-управлением уровнем, инжекцией полей из контекста и встроенным сэмплингом. Проект организации [selfshop-dev](https://github.com/selfshop-dev).

### Installation

```bash
go get -u github.com/selfshop-dev/lib-logger
```

## Overview

`lib-logger` оборачивает два [go.uber.org/zap](https://github.com/uber-go/zap) core, подключённых через `zapcore.NewTee`. Сэмплированный core обрабатывает высокочастотные сообщения ниже порогового уровня; критический core никогда не сэмплируется и доступен напрямую через `Unsampled()` для аудит-логов. Все `With*`-методы возвращают новые экземпляры — ресивер никогда не мутируется.

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

### Быстрый старт

```go
import logger "github.com/selfshop-dev/lib-logger"

cfg := logger.DefaultConfig()
cfg.ServiceName = "payments"

l, _ := logger.New(cfg)
defer l.Sync()

l.Info("ready", zap.String("port", "8080"))
```

## Config

`DefaultConfig` возвращает безопасные production-дефолты: порог `InfoLevel`, сэмплер 100/100, sink — `os.Stdout`. Переопределяй только то, что отличается.

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

Глобальные поля `service`, `version`, `env` и поля из `InitialFields` добавляются к каждой записи автоматически.

## Полная от контекста

`WithContext` обогащает логгер полями, извлечёнными из контекста через один или несколько `FieldExtractor`. Вызывай в начале каждого обработчика запроса.

```go
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log := h.log.WithContext(r.Context(), extractors.Full())
    log.Info("handling request")
}
```

`MergeExtractors` объединяет несколько экстракторов в один:

```go
ex := logger.MergeExtractors(extractors.Tracing(), extractors.UserID())
```

Собственный экстрактор реализуется через `FieldExtractFunc`:

```go
ex := logger.FieldExtractFunc(func(ctx context.Context) []zap.Field {
    return []zap.Field{zap.String("request_id", httpx.RequestID(ctx))}
})
```

## extractors

Субпакет `extractors` предоставляет готовые экстракторы для стандартных полей запроса.

```go
import "github.com/selfshop-dev/lib-logger/extractors"

// Записать в контекст
ctx = extractors.WithTraceID(ctx, "trace-abc")
ctx = extractors.WithUserID(ctx, "usr-1")

// Извлечь при логировании
log := l.WithContext(ctx, extractors.Full())
```

Доступные экстракторы: `TraceID()`, `SpanID()`, `RequestID()`, `CorrelationID()`, `UserID()`, `TenantID()`. Комбинированные: `Tracing()` (trace + span), `Full()` (все шесть полей). Если ключ отсутствует или пуст в контексте — поле молча пропускается.

## Unsampled

`Unsampled()` возвращает underlying критический `*zap.Logger` напрямую. Используй для аудит-логов, событий безопасности и любых сообщений, которые никогда не должны быть отброшены сэмплером.

```go
l.Unsampled().Warn("user password changed",
    zap.String("user_id", uid),
    zap.String("remote_addr", addr),
)
```

`With`, `Named` и `WithContext` на derived-логгере propagate оба core, поэтому `Unsampled()` на дочернем логгере несёт те же поля:

```go
child := l.With(zap.String("request_id", id))
child.Unsampled().Warn("audit event") // содержит request_id
```

## Runtime level

`LevelManager` оборачивает `zap.AtomicLevel` и позволяет менять порог логирования без перезапуска. Зарегистрируй `LevelHandler` в роутере для GET/PUT эндпоинта:

```go
mux.Handle("/log/level", logger.LevelHandler(l.Level, l.Unsampled()))

// GET /log/level  → {"level":"info"}
// PUT /log/level  body: {"level":"debug"} → {"level":"debug"}
```

Unsampled-логгер передаётся в `LevelHandler`, чтобы предупреждение об изменении уровня никогда не попало под сэмплер.

## Лицензия

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)