package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api-template/internal/models"
	"go-api-template/internal/repository"
)

func TestPing(t *testing.T) {
	assertEndpoint(t, "/ping", http.StatusOK, "pong")
}

func TestHealth(t *testing.T) {
	assertEndpoint(t, "/health", http.StatusOK, "ok")
}

func TestGetUserByID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()

	NewRouterWithRepository(repository.NewInMemoryUserRepository()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var got models.UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != "1" {
		t.Fatalf("user id = %q, want 1", got.ID)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/404", nil)
	rec := httptest.NewRecorder()

	NewRouterWithRepository(repository.NewInMemoryUserRepository()).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

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

func assertEndpoint(t *testing.T, path string, wantStatus int, wantBodyStatus string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != wantStatus {
		t.Fatalf("status code = %d, want %d", rec.Code, wantStatus)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}

	var got response
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Status != wantBodyStatus {
		t.Fatalf("response status = %q, want %q", got.Status, wantBodyStatus)
	}
}
