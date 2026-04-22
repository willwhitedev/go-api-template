package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
