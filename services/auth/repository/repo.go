// Package repository provides data access implementations for interacting with databases
package repository

import (
	"context"
	"fmt"
	"golang-dining-ordering/services/auth/dto"
	"strings"
	"time"

	ce "golang-dining-ordering/services/auth/customerrors"
	db "golang-dining-ordering/services/auth/db/generated"

	"github.com/google/uuid"
)

// Repository defines methods for accessing and managing user data.
type Repository interface {
	CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*db.AuthUser, error)
	SaveRefreshToken(ctx context.Context, token string, claims *dto.TokenClaimsDto) error
	GetRefreshToken(ctx context.Context, userID uuid.UUID, token string) error
	DeleteRefreshToken(ctx context.Context, userID uuid.UUID, token string) error
}

type repository struct {
	q *db.Queries
}

// NewRepository creates a new userRepository with the given database connection.
//
//nolint:revive // intended unexported type return
func NewRepository(q *db.Queries) *repository {
	return &repository{
		q: q,
	}
}

func (r *repository) CreateUser(
	ctx context.Context,
	req *dto.SignUpRequestDto,
) (uuid.UUID, error) {
	userRow, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: req.Password,
		Name:         req.Name,
		Lastname:     req.Lastname,
		Role:         int(req.Role),
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return uuid.Nil, &ce.UniqueConstraintError{
				CustomError: ce.CustomError{Message: err.Error()},
			}
		}

		return uuid.Nil, fmt.Errorf("inserting user to db: %w", err)
	}

	return userRow.ID, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*db.AuthUser, error) {
	user, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("fetching user from db: %w", err)
	}

	return &user, nil
}

func (r *repository) SaveRefreshToken(
	ctx context.Context,
	token string,
	claims *dto.TokenClaimsDto,
) error {
	_, err := r.q.SaveRefreshToken(ctx, db.SaveRefreshTokenParams{
		ID:        token,
		UserID:    claims.UserID,
		ExpiresAt: time.Unix(claims.Exp, 0).UTC(),
	})
	if err != nil {
		return fmt.Errorf("saving refresh token: %w", err)
	}

	return nil
}

func (r *repository) GetRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	_, err := r.q.GetRefreshToken(ctx, db.GetRefreshTokenParams{
		UserID: userID,
		ID:     token,
	})
	if err != nil {
		return fmt.Errorf("fetching refresh token: %w", err)
	}

	return nil
}

func (r *repository) DeleteRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	err := r.q.DeleteRefreshToken(ctx, db.DeleteRefreshTokenParams{
		UserID: userID,
		ID:     token,
	})
	if err != nil {
		return fmt.Errorf("deleting refresh token from db: %w", err)
	}

	return nil
}
