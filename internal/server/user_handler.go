package server

import (
	"net/http"

	"github.com/gorilla/mux"
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
		writeJSON(w, http.StatusBadRequest, response{Status: "missing user id"})
		return
	}

	user, found, err := h.users.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Status: "failed to load user"})
		return
	}

	if !found {
		writeJSON(w, http.StatusNotFound, response{Status: "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}
