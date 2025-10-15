// Package services implements the core business logic of the application, such as authentication and user management.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	ce "golang-dining-ordering/internal/customerrors"
	"golang-dining-ordering/internal/dto"
	"golang-dining-ordering/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines authentication-related operations for users.
type AuthService interface {
	SignUpUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error)
	SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.TokenResponseDto, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponseDto, error)
}

// AuthConfig holds configuration values for authentication, such as secret keys
// and token expiration durations.
type AuthConfig struct {
	Secret                 string
	TokenValidHours        int
	RefreshTokenValidHours int
}

type authService struct {
	cfg  *AuthConfig
	repo repository.UsersRepository
}

// NewAuthService creates a new instance of authService.
//
//nolint:revive
func NewAuthService(cfg *AuthConfig, repo repository.UsersRepository) *authService {
	return &authService{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *authService) SignUpUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error) {
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return "", err
	}

	req.Password = hashedPassword

	return s.repo.CreateUser(ctx, req)
}

func (s *authService) SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.TokenResponseDto, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to sign in user: %w", err)
	}

	ok := s.verifyPassword(req.Password, user.PasswordHash)
	if !ok {
		return nil, ce.ErrUnauthorized
	}

	token, err := s.generateToken(user.ID, user.Email, user.Role, s.cfg.TokenValidHours)
	if err != nil {
		return nil, fmt.Errorf("failed to sign in user: %w", err)
	}

	refreshToken, err := s.generateToken(user.ID, user.Email, user.Role, s.cfg.RefreshTokenValidHours)
	if err != nil {
		return nil, fmt.Errorf("failed to sign in user: %w", err)
	}

	res := &dto.TokenResponseDto{
		Token:        token,
		RefreshToken: refreshToken,
	}

	return res, nil
}

func (s *authService) RefreshToken(_ context.Context, refreshToken string) (*dto.TokenResponseDto, error) {
	claims, err := s.verifyToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, userOk := claims["userID"].(string)
	email, emailOk := claims["email"].(string)
	role, roleOk := claims["role"].(string)

	if !userOk || !emailOk || !roleOk {
		return nil, ce.ErrMissingClaims
	}

	newToken, err := s.generateToken(userID, email, role, s.cfg.TokenValidHours)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateToken(userID, email, role, s.cfg.RefreshTokenValidHours)
	if err != nil {
		return nil, err
	}

	res := &dto.TokenResponseDto{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}

	return res, nil
}

func (s *authService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hashed passoword: %w", err)
	}

	return string(hash), nil
}

func (s *authService) verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))

	return err == nil
}

func (s *authService) generateToken(userID, email, role string, validDurationHours int) (string, error) {
	if userID == "" || email == "" || role == "" {
		return "", fmt.Errorf("%w: userID=%s, email=%s, role=%s", ce.ErrInvalidTokenData, userID, email, role)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"email":  email,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * time.Duration(validDurationHours)).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to create SignedString from JWT token: %w", err)
	}

	return tokenStr, nil
}

func (s *authService) verifyToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ce.ErrUnexpectedSigninMethod
		}

		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	if !token.Valid {
		return nil, ce.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ce.ErrParseClaims
	}

	return claims, nil
}
