package testhelpers

import (
	"context"
	db "golang-dining-ordering/internal/db/generated"
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

func (r *MockUsersRepository) CreateUser(ctx context.Context) (*db.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	user := &db.User{
		ID: "some-fake-uuid-1",
	}

	r.users = append(r.users, user)

	return user, nil
}
