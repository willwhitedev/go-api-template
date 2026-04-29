# Template Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add typed response models, request middleware, config, graceful shutdown, Makefile, Docker scaffolding, and an OpenAPI spec to the Go API template.

**Architecture:** New packages (`internal/models`, `internal/middleware`, `internal/config`) are added alongside the existing `internal/server` and `internal/repository` packages. Middleware is wired into the gorilla/mux router via `r.Use(...)`. `main.go` is updated to use config and graceful shutdown. Supporting files (`Makefile`, `Dockerfile`, `docs/openapi.yaml`) are added at the repo root.

**Tech Stack:** Go 1.22, `gorilla/mux`, `log/slog`, `crypto/rand`, standard library only (no new dependencies).

---

### Task 1: Response models

**Files:**
- Create: `internal/models/models.go`

- [ ] **Step 1: Create the models file**

```go
package models

type UserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/models/...`
Expected: no output, exit 0

- [ ] **Step 3: Commit**

```bash
git add internal/models/models.go
git commit -m "feat: add response models package"
```

---

### Task 2: Update handler to use response models

**Files:**
- Modify: `internal/server/router_test.go`
- Modify: `internal/server/user_handler.go`

- [ ] **Step 1: Write a failing test for error response body shape**

In `internal/server/router_test.go`, add this test after `TestGetUserByIDNotFound`:

```go
func TestGetUserByIDNotFoundBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/404", nil)
	rec := httptest.NewRecorder()

	NewRouterWithRepository(repository.NewInMemoryUserRepository()).ServeHTTP(rec, req)

	var got struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Error != "user not found" {
		t.Fatalf("error = %q, want %q", got.Error, "user not found")
	}
}
```

Also update `TestGetUserByID` to decode into `models.UserResponse` instead of `repository.User`. The import block becomes:

```go
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api-template/internal/models"
	"go-api-template/internal/repository"
)
```

And the decode line in `TestGetUserByID`:

```go
var got models.UserResponse
if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
    t.Fatalf("decode response: %v", err)
}
```

- [ ] **Step 2: Run tests to verify the new test fails**

Run: `go test ./internal/server/...`
Expected: FAIL — `TestGetUserByIDNotFoundBody` fails because the handler currently returns `{"status":"user not found"}` not `{"error":"user not found"}`

- [ ] **Step 3: Update user_handler.go to use models**

Replace the entire content of `internal/server/user_handler.go`:

```go
package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"go-api-template/internal/models"
	"go-api-template/internal/repository"
)

type userHandler struct {
	users repository.UserRepository
}

func newUserHandler(users repository.UserRepository) *userHandler {
	return &userHandler{
		users: users,
	}
}

func (h *userHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "missing user id"})
		return
	}

	user, found, err := h.users.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to load user"})
		return
	}

	if !found {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, models.UserResponse{ID: user.ID, Name: user.Name})
}
```

- [ ] **Step 4: Run tests to verify all pass**

Run: `go test ./internal/server/...`
Expected: PASS — all tests pass including the new body assertion

- [ ] **Step 5: Commit**

```bash
git add internal/server/router_test.go internal/server/user_handler.go
git commit -m "feat: use typed response models in user handler"
```

---

### Task 3: Config package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/config/config_test.go`:

```go
package config

import (
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	cfg := Load()
	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q, want :8080", cfg.Addr)
	}
}

func TestLoad_env(t *testing.T) {
	t.Setenv("ADDR", ":9090")
	cfg := Load()
	if cfg.Addr != ":9090" {
		t.Fatalf("Addr = %q, want :9090", cfg.Addr)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/config/...`
Expected: FAIL — compilation error, `config` package does not exist yet

- [ ] **Step 3: Implement config package**

Create `internal/config/config.go`:

```go
package config

import "os"

type Config struct {
	Addr string
}

func Load() Config {
	return Config{
		Addr: getEnv("ADDR", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: add config package"
```

---

### Task 4: Request ID middleware

**Files:**
- Create: `internal/middleware/requestid.go`
- Create: `internal/middleware/requestid_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/middleware/requestid_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_setsHeader(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestRequestID_uniquePerRequest(t *testing.T) {
	var ids [2]string
	i := 0
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids[i] = w.Header().Get("X-Request-ID")
		i++
	}))

	for range ids {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	if ids[0] == ids[1] {
		t.Fatalf("expected unique request IDs, got %q twice", ids[0])
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/middleware/...`
Expected: FAIL — compilation error, `middleware` package does not exist

- [ ] **Step 3: Implement request ID middleware**

Create `internal/middleware/requestid.go`:

```go
package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

type contextKey string

var requestIDKey = contextKey("requestID")

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := generateRequestID()
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/middleware/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/requestid.go internal/middleware/requestid_test.go
git commit -m "feat: add request ID middleware"
```

---

### Task 5: Logger middleware

**Files:**
- Create: `internal/middleware/logger.go`
- Create: `internal/middleware/logger_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/middleware/logger_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger_passesThrough(t *testing.T) {
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestStatusResponseWriter_capturesStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	srw := &statusResponseWriter{ResponseWriter: rec, status: http.StatusOK}
	srw.WriteHeader(http.StatusNotFound)

	if srw.status != http.StatusNotFound {
		t.Fatalf("captured status = %d, want %d", srw.status, http.StatusNotFound)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("underlying status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/middleware/...`
Expected: FAIL — `Logger` and `statusResponseWriter` are undefined

- [ ] **Step 3: Implement logger middleware**

Create `internal/middleware/logger.go`:

```go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *statusResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(srw, r)

		requestID, _ := r.Context().Value(requestIDKey).(string)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", srw.status,
			"duration", time.Since(start),
			"request_id", requestID,
		)
	})
}
```

- [ ] **Step 4: Run all middleware tests to verify they pass**

Run: `go test ./internal/middleware/...`
Expected: PASS — all four middleware tests pass

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/logger.go internal/middleware/logger_test.go
git commit -m "feat: add logger middleware"
```

---

### Task 6: Wire middleware into router

**Files:**
- Modify: `internal/server/router.go`

- [ ] **Step 1: Run existing server tests to establish baseline**

Run: `go test ./internal/server/...`
Expected: PASS — all tests pass before changes

- [ ] **Step 2: Update router.go to wire middleware**

Replace the entire content of `internal/server/router.go`:

```go
package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go-api-template/internal/middleware"
	"go-api-template/internal/repository"
)

type response struct {
	Status string `json:"status"`
}

func NewRouter() http.Handler {
	return NewRouterWithRepository(repository.NewInMemoryUserRepository())
}

func NewRouterWithRepository(users repository.UserRepository) http.Handler {
	r := mux.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	userHandler := newUserHandler(users)

	r.HandleFunc("/ping", pingHandler).Methods(http.MethodGet)
	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/users/{id}", userHandler.getByID).Methods(http.MethodGet)

	return r
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{Status: "pong"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{Status: "ok"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
```

- [ ] **Step 3: Run all tests to verify nothing broke**

Run: `go test ./...`
Expected: PASS — all tests pass with middleware wired in

- [ ] **Step 4: Commit**

```bash
git add internal/server/router.go
git commit -m "feat: wire request ID and logger middleware into router"
```

---

### Task 7: Graceful shutdown and config in main

**Files:**
- Modify: `cmd/api/main.go`

- [ ] **Step 1: Update main.go**

Replace the entire content of `cmd/api/main.go`:

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api-template/internal/config"
	"go-api-template/internal/server"
)

func main() {
	cfg := config.Load()

	api := &http.Server{
		Addr:              cfg.Addr,
		Handler:           server.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("starting api server", "addr", cfg.Addr)
		if err := api.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("api server stopped", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := api.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify it builds**

Run: `go build ./cmd/api/...`
Expected: no output, exit 0

- [ ] **Step 3: Run all tests to confirm nothing broke**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: add graceful shutdown and config-driven server startup"
```

---

### Task 8: Makefile and .gitignore

**Files:**
- Create: `Makefile`
- Create: `.gitignore`

- [ ] **Step 1: Create Makefile**

```makefile
.PHONY: run test build lint docker-build

run:
	go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/api ./cmd/api

lint:
	go vet ./...

docker-build:
	docker build -t go-api-template .
```

> **Note:** Each recipe line MUST be indented with a real tab character, not spaces. If your editor converts tabs to spaces, use `cat -A Makefile` to verify — recipe lines should show `^I` at the start.

- [ ] **Step 2: Create .gitignore**

```
bin/
```

- [ ] **Step 3: Verify make targets**

Run: `make test`
Expected: PASS — runs `go test ./...`

Run: `make build`
Expected: produces `bin/api` executable, exit 0

Run: `make lint`
Expected: no output, exit 0

- [ ] **Step 4: Commit**

```bash
git add Makefile .gitignore
git commit -m "feat: add Makefile and .gitignore"
```

---

### Task 9: Docker scaffolding

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`

- [ ] **Step 1: Create Dockerfile**

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/api ./cmd/api

FROM scratch
COPY --from=builder /app/bin/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
```

- [ ] **Step 2: Create .dockerignore**

```
.git
bin/
docs/
*.md
```

- [ ] **Step 3: Verify the image builds (requires Docker)**

Run: `make docker-build`
Expected: image tagged `go-api-template` is built successfully

If Docker is not available in the current environment, skip this step and note it for manual verification.

- [ ] **Step 4: Commit**

```bash
git add Dockerfile .dockerignore
git commit -m "feat: add multi-stage Dockerfile"
```

---

### Task 10: OpenAPI spec and README

**Files:**
- Create: `docs/openapi.yaml`
- Modify: `README.md`

- [ ] **Step 1: Create docs/openapi.yaml**

```yaml
openapi: "3.0.3"
info:
  title: Go API Template
  version: "1.0.0"
paths:
  /ping:
    get:
      summary: Ping
      responses:
        "200":
          description: Successful response
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/StatusResponse"
  /health:
    get:
      summary: Health check
      responses:
        "200":
          description: Successful response
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/StatusResponse"
  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: User found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/UserResponse"
        "400":
          description: Bad request
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ErrorResponse"
        "404":
          description: User not found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ErrorResponse"
        "500":
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ErrorResponse"
components:
  schemas:
    StatusResponse:
      type: object
      properties:
        status:
          type: string
          example: "ok"
    UserResponse:
      type: object
      properties:
        id:
          type: string
          example: "1"
        name:
          type: string
          example: "Ada Lovelace"
    ErrorResponse:
      type: object
      properties:
        error:
          type: string
          example: "user not found"
```

- [ ] **Step 2: Update README.md**

Replace the entire content of `README.md`:

```markdown
# go-api-template

A Go HTTP API template using `gorilla/mux` with structured logging, request ID middleware, graceful shutdown, and OpenAPI documentation.

## Run

```sh
make run
```

Or directly:

```sh
go run ./cmd/api
```

The API listens on `:8080` by default. Override with the `ADDR` environment variable:

```sh
ADDR=:3000 make run
```

## Endpoints

- `GET /ping` → `{"status":"pong"}`
- `GET /health` → `{"status":"ok"}`
- `GET /users/{id}` → `{"id":"1","name":"Ada Lovelace"}` or `{"error":"user not found"}`

All responses include an `X-Request-ID` header.

## Make targets

| Target | Description |
|--------|-------------|
| `make run` | Run the API server |
| `make test` | Run all tests |
| `make build` | Build binary to `bin/api` |
| `make lint` | Run `go vet` |
| `make docker-build` | Build Docker image tagged `go-api-template` |

## Docker

```sh
make docker-build
docker run -p 8080:8080 go-api-template
```

## API spec

See [`docs/openapi.yaml`](docs/openapi.yaml) for the full OpenAPI 3.0 spec.

## Test

```sh
make test
```
```

- [ ] **Step 3: Run all tests one final time**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add docs/openapi.yaml README.md
git commit -m "feat: add OpenAPI spec and update README"
```
