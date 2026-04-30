package server

import (
	"errors"
	"net/http"

	"go-api-template/internal/models"
	"go-api-template/internal/service"

	"github.com/gorilla/mux"
)

type userHandler struct {
	users service.UserService
}

func newUserHandler(users service.UserService) *userHandler {
	return &userHandler{
		users: users,
	}
}

func (h *userHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user, found, err := h.users.GetByID(r.Context(), id)
	if errors.Is(err, service.ErrMissingUserID) {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "missing user id"})
		return
	}

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
