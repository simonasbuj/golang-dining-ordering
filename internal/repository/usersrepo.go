package repository

import (
	"context"
	db "golang-dining-ordering/internal/db/generated"

	"github.com/google/uuid"
)

type UsersRepository interface {
	CreateUser(ctx context.Context) (*db.User, error)
}

type userRepository struct {
	q *db.Queries
}

func NewUserRepository(q *db.Queries) *userRepository {
	return &userRepository{
		q: q,
	}
}

func (r *userRepository) CreateUser(ctx context.Context) (*db.User, error) {
	user, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID: uuid.New().String(),
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}
