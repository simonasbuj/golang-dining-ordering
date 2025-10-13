package services

import (
	"context"
	"fmt"
	"golang-dining-ordering/internal/dto"
	testhelpers "golang-dining-ordering/test/helpers/repository"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceTestSuite struct {
	suite.Suite

	svc      *authService
	mockRepo *testhelpers.MockUsersRepository
}

func (suite *AuthServiceTestSuite) SetupSuite() {
	suite.mockRepo = testhelpers.NewMockUserRepository()
	suite.svc = NewAuthService("test-auth-secret", suite.mockRepo)
}

func (suite *AuthServiceTestSuite) TestSignUpUser_Success() {
	reqDto := &dto.SignUpRequestDto{}

	expectedUserID := "some-fake-id-1"

	user, err := suite.svc.SignUpUser(context.Background(), reqDto)

	suite.NoError(err)
	suite.Equal(expectedUserID, user)
}

func (suite *AuthServiceTestSuite) TestVerifyPassword() {
	password := "mypassword"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	suite.True(suite.svc.verifyPassword(password, string(hash)))
	suite.False(suite.svc.verifyPassword("wrong-password", string(hash)))
}

func (suite *AuthServiceTestSuite) TestHashPassword() {
	password := "mypassword"

	hashedPassword, err := suite.svc.hashPassword(password)

	suite.NoError(err)
	suite.NotEmpty(hashedPassword)

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	suite.NoError(err, "hash should match original password")

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("wrong-password"))
	suite.Error(err, "hash should not match different password")
}

func (suite *AuthServiceTestSuite) TestGenerateToken_Success() {
	userID := "user-123"
	email := "user-123@email.com"
	role := "waiter"
	duration := 1

	tokenStr, err := suite.svc.generateToken(userID, email, role, duration)

	suite.NoError(err)
	suite.NotEmpty(tokenStr)

	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		suite.Equal(jwt.SigningMethodHS256, token.Method)
		return []byte(suite.svc.secret), nil
	})
	suite.NoError(err)
	suite.True(parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	suite.True(ok)

	suite.Equal(userID, claims["userID"])
	suite.Equal(email, claims["email"])
	suite.Equal(role, claims["role"])

	exp := int64(claims["exp"].(float64))
	suite.Greater(exp, time.Now().Unix())
	suite.LessOrEqual(exp, time.Now().Add(time.Hour*time.Duration(1)).Unix())
}

func (suite *AuthServiceTestSuite) TestGenerateToken_InvalidInput() {
	testCases := []struct {
		name   string
		userID string
		email  string
		Role   string
	}{
		{"missing userID", "", "user@email.com", "waiter"},
		{"missing email", "user-123", "", "waiter"},
		{"missing role", "user=123", "user@email.com", ""},
	}

	for _, tc := range testCases {
		tokenStr, err := suite.svc.generateToken(tc.userID, tc.email, tc.Role, 1)

		suite.Error(err, fmt.Sprintf("expected error when %s", tc.name))
		suite.Empty(tokenStr, fmt.Sprintf("token should be empty when %s", tc.name))
	}

}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
