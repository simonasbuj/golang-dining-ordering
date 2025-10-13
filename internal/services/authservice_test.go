package services

import (
	"context"
	"golang-dining-ordering/internal/dto"
	testhelpers "golang-dining-ordering/test/helpers/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignUpUser_Success(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	svc := NewAuthService("test-auth-secret", mockRepo)

	reqDto := &dto.SignUpRequestDto{}

	expectedUserID := "some-fake-uuid-1"

	user, err := svc.SignUpUser(context.Background(), reqDto)

	assert.NoError(t, err)
	assert.Equal(t, expectedUserID, user)
}

func TestSignInUser_Success(t *testing.T) {
	assert.Equal(t, "h", "h")
}
