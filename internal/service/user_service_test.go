package service

import (
	"context"
	"errors"
	"testing"

	"go-api-template/internal/repository"
)

func TestUserServiceGetByIDReturnsUser(t *testing.T) {
	users := NewUserService(&stubUserRepository{
		user:  repository.User{ID: "1", Name: "Ada Lovelace"},
		found: true,
	})

	got, found, err := users.GetByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if !found {
		t.Fatal("found = false, want true")
	}

	if got.ID != "1" || got.Name != "Ada Lovelace" {
		t.Fatalf("user = %+v, want Ada Lovelace with ID 1", got)
	}
}

func TestUserServiceGetByIDRejectsMissingID(t *testing.T) {
	users := NewUserService(&stubUserRepository{})

	_, _, err := users.GetByID(context.Background(), "")
	if !errors.Is(err, ErrMissingUserID) {
		t.Fatalf("error = %v, want %v", err, ErrMissingUserID)
	}
}

type stubUserRepository struct {
	user  repository.User
	found bool
	err   error
}

func (r *stubUserRepository) FindByID(ctx context.Context, id string) (repository.User, bool, error) {
	return r.user, r.found, r.err
}
