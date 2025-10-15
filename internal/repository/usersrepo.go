// Package repository provides data access implementations for interacting with databases
package repository

import (
	"context"
	"golang-dining-ordering/internal/dto"
	"strings"

	ce "golang-dining-ordering/internal/customerrors"
	db "golang-dining-ordering/internal/db/generated"

	"github.com/google/uuid"
)

// UsersRepository defines methods for accessing and managing user data.
type UsersRepository interface {
	CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
}

type userRepository struct {
	q *db.Queries
}

// NewUserRepository creates a new userRepository with the given database connection.
//
//nolint:revive
func NewUserRepository(q *db.Queries) *userRepository {
	return &userRepository{
		q: q,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error) {
	userRow, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: req.Password,
		Name:         req.Name,
		Lastname:     req.Lastname,
		Role:         req.Role,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return "", &ce.UniqueConstraintError{
				CustomError: ce.CustomError{Message: err.Error()},
			}
		}

		return "", err
	}

	return userRow.ID, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
