package service

import (
	"context"
	db "golang-dining-ordering/services/auth/db/generated"
	"golang-dining-ordering/services/auth/dto"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

//nolint:gochecknoglobals
var (
	TestUserID             = uuid.MustParse("67676767-6767-4676-8767-676767676767")
	TestEmail              = "user-123@email.com"
	TestPassword           = "password123"
	TestName               = "sim"
	TestLastname           = "sim"
	TestRole               = dto.Role(2)
	TestTokenVersion int64 = 2
)

// MockUsersRepository is a mock implementation of repository.UsersRepository.
type mockUsersRepository struct {
	sync.Mutex

	users []*db.AuthUser
}

// NewMockUserRepository creates a new mock implementation of UsersRepository for testing.
func NewMockUserRepository() *mockUsersRepository {
	return &mockUsersRepository{
		users: make([]*db.AuthUser, 0),
		Mutex: sync.Mutex{},
	}
}

// CreateUser returns a mock user for testing purposes.
func (r *mockUsersRepository) CreateUser(
	_ context.Context,
	req *dto.SignUpRequestDto,
) (uuid.UUID, error) {
	r.Lock()
	defer r.Unlock()

	user := &db.AuthUser{ //nolint:exhaustruct
		ID:           uuid.MustParse("67676767-6767-4676-8767-676767676767"),
		Email:        req.Email,
		PasswordHash: req.Password,
		Name:         req.Name,
		Lastname:     req.Lastname,
		Role:         int(req.Role),
	}

	r.users = append(r.users, user)

	return user.ID, nil
}

// GetUserByEmail returns a mock user for testing purposes.
func (r *mockUsersRepository) GetUserByEmail(
	_ context.Context,
	email string,
) (*db.AuthUser, error) {
	user := &db.AuthUser{ //nolint:exhaustruct
		ID:    uuid.MustParse("67676767-6767-4676-8767-676767676767"),
		Email: email,
		// hash for password123 with cost factor = 10
		PasswordHash: "$2a$10$00.4AZj71Ls5Riz43mlXUebnpdCuBWine0/v3KtSPpmM/Cb3IyURi",
		Role:         2,
	}

	return user, nil
}

// SaveRefreshToken mocks saving refresh token in db.
func (r *mockUsersRepository) SaveRefreshToken(
	_ context.Context,
	_ string,
	_ *dto.TokenClaimsDto,
) error {
	return nil
}

// GetRefreshToken mocks fetching refresh token from db.
func (r *mockUsersRepository) GetRefreshToken(
	_ context.Context,
	_ uuid.UUID,
	_ string,
) error {
	return nil
}

// DeleteRefreshToken mocks deleting refresh token from db.
func (r *mockUsersRepository) DeleteRefreshToken(
	_ context.Context,
	_ uuid.UUID,
	_ string,
) error {
	return nil
}

type AuthServiceTestSuite struct {
	suite.Suite

	svc      *service
	mockRepo *mockUsersRepository
}

func (suite *AuthServiceTestSuite) SetupSuite() {
	suite.mockRepo = NewMockUserRepository()
	cfg := &Config{
		Secret:                   "test-auth-secret",
		TokenValidSeconds:        604800,
		RefreshTokenValidSeconds: 1209600,
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

	expectedUserID := TestUserID

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
	tokenStr, err := suite.svc.generateToken(generateTokenParams{
		UserID:               TestUserID,
		Email:                TestEmail,
		TokenType:            tokenTypeAccess,
		Role:                 TestRole,
		ValidDurationSeconds: 1,
	})

	suite.Require().NoError(err)
	suite.NotEmpty(tokenStr)

	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		suite.Equal(jwt.SigningMethodHS256, token.Method)

		return []byte(suite.svc.cfg.Secret), nil
	})
	suite.Require().NoError(err)
	suite.True(parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	suite.Require().True(ok)

	suite.Equal(TestUserID, uuid.MustParse(claims["userID"].(string))) //nolint:forcetypeassert
	suite.Equal(TestEmail, claims["email"])
	suite.Equal(
		TestRole,
		dto.Role(int(claims["role"].(float64))), //nolint:forcetypeassert
	)

	expFloat, ok := claims["exp"].(float64)
	suite.True(ok, "exp claim should be a float64")

	exp := int64(expFloat)

	suite.Greater(exp, time.Now().Unix())
	suite.LessOrEqual(exp, time.Now().Add(time.Hour*time.Duration(1)).Unix())
}

func (suite *AuthServiceTestSuite) TestGenerateToken_InvalidInput() {
	testCases := []struct {
		desc         string
		userID       uuid.UUID
		email        string
		role         dto.Role
		tokenVersion int64
	}{
		{"missing userID", uuid.Nil, TestEmail, TestRole, TestTokenVersion},
		{"missing email", TestUserID, "", TestRole, TestTokenVersion},
		{"missing role", TestUserID, TestEmail, 0, TestTokenVersion},
		{"missing tokenVersion", TestUserID, TestEmail, 0, 0},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.desc, func(_ *testing.T) {
			tokenStr, err := suite.svc.generateToken(generateTokenParams{
				UserID:               tc.userID,
				Email:                tc.email,
				TokenType:            tokenTypeAccess,
				Role:                 tc.role,
				ValidDurationSeconds: 1,
			})

			suite.Require().Error(err, "expected error when %s", tc.desc)
			suite.Require().Empty(tokenStr, "token should be empty when %s", tc.desc)
		})
	}
}

func (suite *AuthServiceTestSuite) TestVerifyToken_Success() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    TestUserID,
		"email":     TestEmail,
		"tokenType": tokenTypeAccess,
		"role":      TestRole,
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(suite.svc.cfg.Secret))
	suite.Require().NoError(err)

	claims, err := suite.svc.verifyToken(context.Background(), tokenStr, tokenTypeAccess)

	suite.Require().NoError(err)
	suite.NotNil(claims)
	suite.Equal(TestUserID, claims.UserID)
	suite.Equal(TestEmail, claims.Email)
	suite.Equal(TestRole, claims.Role)
}

func (suite *AuthServiceTestSuite) TestVerifyToken_InvalidSecret() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": TestUserID,
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte("different-secret"))
	suite.Require().NoError(err)

	claims, err := suite.svc.verifyToken(context.Background(), tokenStr, tokenTypeAccess)
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

	claims, err := suite.svc.verifyToken(context.Background(), tokenStr, tokenTypeAccess)
	suite.Require().Error(err)
	suite.Nil(claims)
}

func (suite *AuthServiceTestSuite) TestVerifyToken_MalformedToken() {
	claims, err := suite.svc.verifyToken(
		context.Background(),
		"this-is-not-a-jwt-not-even-close",
		tokenTypeAccess,
	)
	suite.Require().Error(err)
	suite.Nil(claims)
}

// hmm using other functions from svc like generateToken and verifyToken as helpers in test to make it easier to test?
// but this seems wrong,,, like test should only call one func.
func (suite *AuthServiceTestSuite) TestRefreshToken_Success1() {
	validRefreshToken, err := suite.svc.generateToken(generateTokenParams{
		UserID:               TestUserID,
		Email:                TestEmail,
		TokenType:            tokenTypeRefresh,
		Role:                 TestRole,
		ValidDurationSeconds: 24,
	})
	suite.Require().NoError(err)

	res, err := suite.svc.RefreshToken(context.Background(), validRefreshToken)

	suite.Require().NoError(err)
	suite.NotNil(res)
	suite.NotEmpty(res.Token)
	suite.NotEmpty(res.RefreshToken)

	tokens := []struct {
		tokenType string
		value     string
	}{
		{tokenTypeAccess, res.Token},
		{tokenTypeRefresh, res.RefreshToken},
	}

	for _, tk := range tokens {
		suite.T().Run(tk.tokenType, func(_ *testing.T) {
			claims, err := suite.svc.verifyToken(context.Background(), tk.value, tk.tokenType)
			suite.Require().NoError(err)
			suite.Equal(TestUserID, claims.UserID)
			suite.Equal(TestEmail, claims.Email)
			suite.Equal(TestRole, claims.Role)
		})
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
