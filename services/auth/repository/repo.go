// Package repository provides data access implementations for interacting with databases
package repository

import (
	"context"
	"fmt"
	"golang-dining-ordering/services/auth/dto"
	"strings"

	ce "golang-dining-ordering/services/auth/customerrors"
	db "golang-dining-ordering/services/auth/db/generated"

	"github.com/google/uuid"
)

// Repository defines methods for accessing and managing user data.
type Repository interface {
	CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*db.User, error)
}

type repository struct {
	q *db.Queries
}

// NewUserRepository creates a new userRepository with the given database connection.
//
//nolint:revive // intended unexported type return
func NewUserRepository(q *db.Queries) *repository {
	return &repository{
		q: q,
	}
}

func (r *repository) CreateUser(
	ctx context.Context,
	req *dto.SignUpRequestDto,
) (string, error) {
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

		return "", fmt.Errorf("failed to insert user to db: %w", err)
	}

	return userRow.ID, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from db: %w", err)
	}

	return &user, nil
}
