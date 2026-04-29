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
