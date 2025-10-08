package services

import (
	"context"
	db "golang-dining-ordering/internal/db/generated"
	testhelpers "golang-dining-ordering/test/helpers/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser_Success(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	svc := NewUserService(mockRepo)

	expectedUser := &db.User{
		ID: "some-fake-uuid-1",
	}

	user, err := svc.CreateUser(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}
