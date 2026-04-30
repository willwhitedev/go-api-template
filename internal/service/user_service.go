package service

import (
	"context"
	"errors"

	"go-api-template/internal/repository"
)

var ErrMissingUserID = errors.New("missing user id")

type UserService interface {
	GetByID(ctx context.Context, id string) (repository.User, bool, error)
}

type RepositoryUserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *RepositoryUserService {
	return &RepositoryUserService{
		users: users,
	}
}

func (s *RepositoryUserService) GetByID(ctx context.Context, id string) (repository.User, bool, error) {
	if id == "" {
		return repository.User{}, false, ErrMissingUserID
	}

	return s.users.FindByID(ctx, id)
}
