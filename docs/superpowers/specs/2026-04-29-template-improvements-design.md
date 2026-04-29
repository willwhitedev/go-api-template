---
title: Go API Template Improvements
date: 2026-04-29
status: approved
---

# Go API Template Improvements

## Goal

Improve the template with patterns every real Go API needs, without overcomplicating its structure. The result should be a clear, readable starting point developers can confidently extend.

## Scope

Eight focused additions:

1. `internal/models` — typed response structs
2. Request logging middleware
3. Request ID middleware + context propagation
4. Config struct loaded from env vars
5. Graceful shutdown
6. `Makefile`
7. `docs/openapi.yaml`
8. Docker scaffolding

---

## 1. Response Models (`internal/models/models.go`)

Add a new package with two structs:

```go
type UserResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}
```

`UserResponse` decouples the API response shape from `repository.User` so they can diverge independently. `ErrorResponse` gives error payloads a distinct JSON shape (`"error"` key) that clients can reliably distinguish from success responses.

**Handler updates:** `user_handler.go` maps `repository.User` → `models.UserResponse` on success and uses `models.ErrorResponse` for all error cases (400, 404, 500). The ping/health handlers keep the existing `response{Status}` struct — they are simple status responses and don't need the same separation.

**Test updates:** `TestGetUserByID` decodes into `models.UserResponse` instead of `repository.User`.

---

## 2. Middleware (`internal/middleware/`)

Two middleware functions in a new `internal/middleware` package.

### Request ID (`requestid.go`)

Generates a UUID v4 per request using `crypto/rand` (no new dependency), sets it on the response as `X-Request-ID`, and stores it in the request context under an unexported typed key:

```go
type contextKey string
var requestIDKey = contextKey("requestID")
```

The key is unexported because only code within the `middleware` package reads it (the logger). A typed key (not a plain string) is still used to prevent context value collisions — the pattern is demonstrated without unnecessary export.

### Logger (`logger.go`)

Wraps the next handler and logs: method, path, status code, duration, and request ID. Uses `log/slog` (already in use in `main.go`). Log format:

```
method=GET path=/users/1 status=200 duration=1.2ms request_id=<uuid>
```

Capturing the status code requires a `responseWriter` wrapper that implements `http.ResponseWriter` and records the written status code — a standard Go middleware pattern worth showing explicitly in the template.

### Wiring

Both middleware are applied in `router.go` via `r.Use(...)` — request ID first, then logger, so the logger can read the ID from context.

---

## 3. Context Propagation

The request ID stored in context by the request ID middleware flows through to the repository layer. `user_handler.go` passes `r.Context()` to `h.users.FindByID(...)` — this is already the case. No handler changes needed beyond what's in section 1.

The repository already accepts `ctx context.Context` and checks `ctx.Err()`. No repository changes needed — the context is already threaded correctly.

---

## 4. Config (`internal/config/config.go`)

A `Config` struct loaded once at startup from environment variables:

```go
type Config struct {
    Addr string // default: ":8080"
}

func Load() Config {
    return Config{
        Addr: getEnv("ADDR", ":8080"),
    }
}
```

`main.go` calls `config.Load()` and passes `cfg.Addr` to the server. The `getEnv` helper moves from `main.go` into `config.go`. This gives future developers a single place to add new env-backed settings.

---

## 5. Graceful Shutdown (`cmd/api/main.go`)

On `SIGTERM` or `SIGINT`, call `server.Shutdown(ctx)` with a 10-second timeout to drain in-flight requests before exiting. Pattern:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
server.Shutdown(ctx)
```

The server starts in a goroutine so `main` can block on the signal channel.

---

## 6. Makefile

A `Makefile` at the repo root with four targets:

| Target | Command |
|--------|---------|
| `make run` | `go run ./cmd/api` |
| `make test` | `go test ./...` |
| `make build` | `go build -o bin/api ./cmd/api` |
| `make lint` | `go vet ./...` |

`.PHONY` declared for all targets. `bin/` added to `.gitignore`.

---

## 7. OpenAPI Spec (`docs/openapi.yaml`)

A hand-written OpenAPI 3.0 spec documenting all three endpoints. No code generation — the spec is the source of truth and is maintained manually alongside handler changes.

Endpoints documented:

- `GET /ping` → `200` `{"status": "pong"}`
- `GET /health` → `200` `{"status": "ok"}`
- `GET /users/{id}` → `200` `UserResponse`, `404` `ErrorResponse`, `500` `ErrorResponse`

Components section defines `UserResponse` and `ErrorResponse` schemas, matching the Go structs in `internal/models`.

---

## 8. Docker Scaffolding

A multi-stage `Dockerfile` at the repo root:

- **Build stage:** `golang:1.22-alpine` — compiles the binary
- **Runtime stage:** `scratch` (or `alpine`) — copies only the compiled binary, keeping the image minimal

A `.dockerignore` excludes `.git`, `bin/`, and other non-essential files from the build context.

No `docker-compose.yml` — the template has no external service dependencies (in-memory repository), so compose adds no value here.

**Makefile addition:** `make docker-build` target: `docker build -t go-api-template .`

---

## File Changes Summary

| File | Action |
|------|--------|
| `internal/models/models.go` | Create |
| `internal/middleware/requestid.go` | Create |
| `internal/middleware/logger.go` | Create |
| `internal/config/config.go` | Create |
| `docs/openapi.yaml` | Create |
| `Makefile` | Create |
| `Dockerfile` | Create |
| `.dockerignore` | Create |
| `cmd/api/main.go` | Update — graceful shutdown, use config |
| `internal/server/router.go` | Update — wire middleware |
| `internal/server/user_handler.go` | Update — use models |
| `internal/server/router_test.go` | Update — decode into models.UserResponse |
| `README.md` | Update — document Makefile targets, config vars, Docker usage |
| `.gitignore` | Create or update — add `bin/` |

---

## What's Explicitly Out of Scope

- Database integration (keep in-memory repository)
- Authentication / JWT
- Dependency injection frameworks (manual constructor injection via `NewRouterWithRepository` is sufficient and idiomatic)
- Code-generated API docs (swaggo/swag)
- Docker Compose (no external service dependencies)
