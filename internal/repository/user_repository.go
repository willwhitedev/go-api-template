package repository

import "context"

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserRepository interface {
	FindByID(ctx context.Context, id string) (User, bool, error)
}

type InMemoryUserRepository struct {
	users map[string]User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: map[string]User{
			"1": {
				ID:   "1",
				Name: "Ada Lovelace",
			},
			"2": {
				ID:   "2",
				Name: "Grace Hopper",
			},
		},
	}
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (User, bool, error) {
	if err := ctx.Err(); err != nil {
		return User{}, false, err
	}

	user, found := r.users[id]
	return user, found, nil
}
