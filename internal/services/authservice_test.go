package services

import (
	"context"
	"testing"
	"time"

	"golang-dining-ordering/internal/dto"
	testhelpers "golang-dining-ordering/test/helpers/repository"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

const (
	TestUserID   = "user-123"
	TestEmail    = "user-123@email.com"
	TestPassword = "password123"
	TestName     = "sim"
	TestLastname = "sim"
	TestRole     = "waiter"
)

type AuthServiceTestSuite struct {
	suite.Suite

	svc      *authService
	mockRepo *testhelpers.MockUsersRepository
}

func (suite *AuthServiceTestSuite) SetupSuite() {
	suite.mockRepo = testhelpers.NewMockUserRepository()
	cfg := &AuthConfig{
		Secret:                 "test-auth-secret",
		TokenValidHours:        168,
		RefreshTokenValidHours: 336,
	}
	suite.svc = NewAuthService(cfg, suite.mockRepo)
}

func (suite *AuthServiceTestSuite) TestSignUpUser_Success() {
	reqDto := &dto.SignUpRequestDto{
		Email:    TestEmail,
		Password: TestPassword,
		Name:     TestName,
		Lastname: TestLastname,
		Role:     TestRole,
	}

	expectedUserID := "some-fake-id-1"

	user, err := suite.svc.SignUpUser(context.Background(), reqDto)

	suite.Require().NoError(err)
	suite.Equal(expectedUserID, user)
}

func (suite *AuthServiceTestSuite) TestVerifyPassword() {
	password := TestPassword
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	suite.True(suite.svc.verifyPassword(password, string(hash)))
	suite.False(suite.svc.verifyPassword("wrong-password", string(hash)))
}

func (suite *AuthServiceTestSuite) TestHashPassword() {
	password := TestPassword

	hashedPassword, err := suite.svc.hashPassword(password)

	suite.Require().NoError(err)
	suite.NotEmpty(hashedPassword)

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	suite.Require().NoError(err, "hash should match original password")

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("wrong-password"))
	suite.Require().Error(err, "hash should not match different password")
}

func (suite *AuthServiceTestSuite) TestGenerateToken_Success() {
	tokenStr, err := suite.svc.generateToken(TestUserID, TestEmail, TestRole, 1)

	suite.Require().NoError(err)
	suite.NotEmpty(tokenStr)

	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		suite.Equal(jwt.SigningMethodHS256, token.Method)

		return []byte(suite.svc.cfg.Secret), nil
	})
	suite.Require().NoError(err)
	suite.True(parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	suite.True(ok)

	suite.Equal(TestUserID, claims["userID"])
	suite.Equal(TestEmail, claims["email"])
	suite.Equal(TestRole, claims["role"])

	expFloat, ok := claims["exp"].(float64)
	suite.True(ok, "exp claim should be a float64")

	exp := int64(expFloat)

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
		{"missing userID", "", TestEmail, TestRole},
		{"missing email", TestUserID, "", TestRole},
		{"missing role", TestUserID, TestEmail, ""},
	}

	for _, tc := range testCases {
		tokenStr, err := suite.svc.generateToken(tc.userID, tc.email, tc.Role, 1)

		suite.Require().Error(err, "expected error when %s", tc.desc)
		suite.Require().Empty(tokenStr, "token should be empty when %s", tc.desc)
	}
}

func (suite *AuthServiceTestSuite) TestVerifyToken_Success() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": TestUserID,
		"email":  TestEmail,
		"role":   TestRole,
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(suite.svc.cfg.Secret))
	suite.Require().NoError(err)

	claims, err := suite.svc.verifyToken(tokenStr)

	suite.Require().NoError(err)
	suite.NotNil(claims)
	suite.Equal(TestUserID, claims["userID"])
	suite.Equal(TestEmail, claims["email"])
	suite.Equal(TestRole, claims["role"])
}

func (suite *AuthServiceTestSuite) TestVerifyToken_InvalidSecret() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": TestUserID,
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte("different-secret"))
	suite.Require().NoError(err)

	claims, err := suite.svc.verifyToken(tokenStr)
	suite.Require().Error(err)
	suite.Nil(claims)
}

func (suite *AuthServiceTestSuite) TestVerifyToken_ExpiredToken() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": TestUserID,
		"exp":    time.Now().Add(-time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(suite.svc.cfg.Secret))
	suite.Require().NoError(err)

	claims, err := suite.svc.verifyToken(tokenStr)
	suite.Require().Error(err)
	suite.Nil(claims)
}

func (suite *AuthServiceTestSuite) TestVerifyToken_MalformedToken() {
	claims, err := suite.svc.verifyToken("this-is-not-a-jwt-not-even-close")
	suite.Require().Error(err)
	suite.Nil(claims)
}

// hmm using other functions from svc like generateToken and verifyToken as helpers in test to make it easier to test?
// but this seems wrong,,, like test should only call one func.
func (suite *AuthServiceTestSuite) TestRefreshToken_Success1() {
	validRefreshToken, err := suite.svc.generateToken(TestUserID, TestEmail, TestRole, 24)
	suite.Require().NoError(err)

	res, err := suite.svc.RefreshToken(context.Background(), validRefreshToken)

	suite.Require().NoError(err)
	suite.NotNil(res)
	suite.NotEmpty(res.Token)
	suite.NotEmpty(res.RefreshToken)

	for _, tokenStr := range []string{res.Token, res.RefreshToken} {
		claims, err := suite.svc.verifyToken(tokenStr)
		suite.Require().NoError(err)
		suite.Equal(TestUserID, claims["userID"])
		suite.Equal(TestEmail, claims["email"])
		suite.Equal(TestRole, claims["role"])
	}
}

func (suite *AuthServiceTestSuite) TestRefreshToken_InvalidToken() {
	invalidToken := "this-is-definitely-not-a-valid-jwt-token"

	res, err := suite.svc.RefreshToken(context.Background(), invalidToken)

	suite.Require().Error(err)
	suite.Nil(res)
}

func (suite *AuthServiceTestSuite) TestRefreshToken_TokenMissingClaims() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(suite.svc.cfg.Secret))
	suite.Require().NoError(err)

	res, err := suite.svc.RefreshToken(context.Background(), tokenStr)
	suite.Require().Error(err)
	suite.Nil(res)
}

func (suite *AuthServiceTestSuite) TestSignInUser_Success() {
	req := &dto.SignInRequestDto{
		Email:    TestEmail,
		Password: TestPassword,
	}

	res, err := suite.svc.SignInUser(context.Background(), req)
	suite.Require().NoError(err)
	suite.NotNil(res)
	suite.NotEmpty(res.Token)
	suite.NotEmpty(res.RefreshToken)
}

func TestAuthServiceTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(AuthServiceTestSuite))
}
