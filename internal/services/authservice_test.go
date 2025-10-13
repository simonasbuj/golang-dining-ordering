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
		desc   string
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

		suite.Error(err, fmt.Sprintf("expected error when %s", tc.desc))
		suite.Empty(tokenStr, fmt.Sprintf("token should be empty when %s", tc.desc))
	}

}

func (suite *AuthServiceTestSuite) TestVerifyToken_Success() {
	userID := "user-123"
	email := "user-123@email.com"
	role := "waiter"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"email":  email,
		"role":   role,
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(suite.svc.secret))
	suite.Require().NoError(err)

	claims, err := suite.svc.verifyToken(tokenStr)

	suite.NoError(err)
	suite.NotNil(claims)
	suite.Equal(userID, claims["userID"])
	suite.Equal(email, claims["email"])
	suite.Equal(role, claims["role"])
}

func (suite *AuthServiceTestSuite) TestVerifyToken_InvalidSecret() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": "user-123",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte("different-secret"))
	suite.Require().NoError(err)

	claims, err := suite.svc.verifyToken(tokenStr)
	suite.Error(err)
	suite.Nil(claims)
}

func (suite *AuthServiceTestSuite) TestVerifyToken_ExpiredToken() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": "user-123",
		"exp":    time.Now().Add(-time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(suite.svc.secret))
	suite.Require().NoError(err)

	claims, err := suite.svc.verifyToken(tokenStr)
	suite.Error(err)
	suite.Nil(claims)
}

func (suite *AuthServiceTestSuite) TestVerifyToken_MalformedToken() {
	claims, err := suite.svc.verifyToken("this-is-not-a-jwt-not-even-close")
	suite.Error(err)
	suite.Nil(claims)
}

// hmm using other functions from svc like generateToken and verifyToken as helpers in test to make it easier to test?
// but this seems wrong,,, like test should only call one func.
func (suite *AuthServiceTestSuite) TestRefreshToken_Success1() {
	userID := "user-123"
	email := "user-123@email.com"
	role := "waiter"

	validRefreshToken, err := suite.svc.generateToken(userID, email, role, 24)
	suite.Require().NoError(err)

	res, err := suite.svc.RefreshToken(context.Background(), validRefreshToken)

	suite.NoError(err)
	suite.NotNil(res)
	suite.NotEmpty(res.Token)
	suite.NotEmpty(res.RefreshToken)

	for _, tokenStr := range []string{res.Token, res.RefreshToken} {
		claims, err := suite.svc.verifyToken(tokenStr)
		suite.NoError(err)
		suite.Equal(userID, claims["userID"])
		suite.Equal(email, claims["email"])
		suite.Equal(role, claims["role"])
	}
}

func (suite *AuthServiceTestSuite) TestRefreshToken_InvalidToken() {
	invalidToken := "this-is-definitely-not-a-valid-jwt-token"

	res, err := suite.svc.RefreshToken(context.Background(), invalidToken)

	suite.Error(err)
	suite.Nil(res)
}

func (suite *AuthServiceTestSuite) TestRefreshToken_TokenMissingClaims() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(suite.svc.secret))
	suite.Require().NoError(err)

	res, err := suite.svc.RefreshToken(context.Background(), tokenStr)
	suite.Error(err)
	suite.Nil(res)
}

func (suite *AuthServiceTestSuite) TestSignInUser_Success() {
	req := &dto.SignInRequestDto{
		Email:    "user@email.com",
		Password: "password123",
	}

	res, err := suite.svc.SignInUser(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(res)
	suite.NotEmpty(res.Token)
	suite.NotEmpty(res.RefreshToken)
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
