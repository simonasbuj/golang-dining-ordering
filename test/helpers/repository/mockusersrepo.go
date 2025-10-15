// Package testhelpers provides mock implementations of data repositories for testing.
package testhelpers

import (
	"context"
	"sync"

	db "golang-dining-ordering/internal/db/generated"
	"golang-dining-ordering/internal/dto"
)

// MockUsersRepository is a mock implementation of repository.UsersRepository.
type MockUsersRepository struct {
	sync.Mutex

	users []*db.User
}

// NewMockUserRepository creates a new mock implementation of UsersRepository for testing.
func NewMockUserRepository() *MockUsersRepository {
	return &MockUsersRepository{
		users: make([]*db.User, 0),
		Mutex: sync.Mutex{},
	}
}

// CreateUser returns a mock user for testing purposes.
func (r *MockUsersRepository) CreateUser(_ context.Context, req *dto.SignUpRequestDto) (string, error) {
	r.Lock()
	defer r.Unlock()

	user := &db.User{ //nolint:exhaustruct
		ID:           "some-fake-id-1",
		Email:        req.Email,
		PasswordHash: req.Password,
		Name:         req.Name,
		Lastname:     req.Lastname,
		Role:         req.Role,
	}

	r.users = append(r.users, user)

	return user.ID, nil
}

// GetUserByEmail returns a mock user for testing purposes.
func (r *MockUsersRepository) GetUserByEmail(_ context.Context, email string) (*db.User, error) {
	user := &db.User{ //nolint:exhaustruct
		ID:    "user-123",
		Email: email,
		// hash for password123 with cost factor = 10
		PasswordHash: "$2a$10$00.4AZj71Ls5Riz43mlXUebnpdCuBWine0/v3KtSPpmM/Cb3IyURi",
		Role:         "waiter",
	}

	return user, nil
}
