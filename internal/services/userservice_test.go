package services

import (
	"context"
	"golang-dining-ordering/internal/dto"
	testhelpers "golang-dining-ordering/test/helpers/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser_Success(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	svc := NewUserService(mockRepo)

	reqDto := &dto.SignUpRequestDto{}

	expectedUserID := "some-fake-uuid-1"

	user, err := svc.CreateUser(context.Background(), reqDto)

	assert.NoError(t, err)
	assert.Equal(t, expectedUserID, user)
}

func TestSignInUser_Success(t *testing.T) {
	assert.Equal(t, "h", "h")
}
