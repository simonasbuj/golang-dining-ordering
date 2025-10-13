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
		PasswordHash: "hashed-pw",
	}

	return user, nil
}
