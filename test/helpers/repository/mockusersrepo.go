package testhelpers

import (
	"context"
	db "golang-dining-ordering/internal/db/generated"
	"golang-dining-ordering/internal/dto"
	"sync"
)

// MockUsersRepository is a mock implementation of repository.UsersRepository
type MockUsersRepository struct {
	users []*db.User

	mutex sync.Mutex
}

func NewMockUserRepository() *MockUsersRepository {
	return &MockUsersRepository{
		users: make([]*db.User, 0),
	}
}

func (r *MockUsersRepository) CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	user := &db.User{
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

func (r *MockUsersRepository) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	user := &db.User{
		ID:           "user-123",
		Email:        "user@email.com",
		PasswordHash: "$2a$10$00.4AZj71Ls5Riz43mlXUebnpdCuBWine0/v3KtSPpmM/Cb3IyURi", //hash for password123 with cost factor = 10
		Role:         "waiter",
	}

	return user, nil
}
