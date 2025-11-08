// Package service implements the core business logic of the application, such as authentication and user management.
package service

import (
	"context"
	"fmt"
	"golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/auth/repository"
	"time"

	ce "golang-dining-ordering/services/auth/customerrors"

	"github.com/golang-jwt/jwt/v4"

	"golang.org/x/crypto/bcrypt"
)

// Service defines authentication-related operations for users.
type Service interface {
	SignUpUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error)
	SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.TokenResponseDto, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponseDto, error)
}

// Config holds configuration values for authentication, such as secret keys
// and token expiration durations.
type Config struct {
	Secret                   string
	TokenValidSeconds        int
	RefreshTokenValidSeconds int
}

type service struct {
	cfg  *Config
	repo repository.Repository
}

// NewAuthService creates a new instance of authService.
//
//nolint:revive // intended unexported type return
func NewAuthService(cfg *Config, repo repository.Repository) *service {
	return &service{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *service) SignUpUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error) {
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return "", err
	}

	req.Password = hashedPassword

	userID, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to sign up user: %w", err)
	}

	return userID, nil
}

func (s *service) SignInUser(
	ctx context.Context,
	req *dto.SignInRequestDto,
) (*dto.TokenResponseDto, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from db: %w", err)
	}

	ok := s.verifyPassword(req.Password, user.PasswordHash)
	if !ok {
		return nil, ce.ErrUnauthorized
	}

	token, err := s.generateToken(user.ID, user.Email, user.Role, s.cfg.TokenValidSeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := s.generateToken(
		user.ID,
		user.Email,
		user.Role,
		s.cfg.RefreshTokenValidSeconds,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	res := &dto.TokenResponseDto{
		Token:        token,
		RefreshToken: refreshToken,
	}

	return res, nil
}

func (s *service) RefreshToken(
	_ context.Context,
	refreshToken string,
) (*dto.TokenResponseDto, error) {
	claims, err := s.verifyToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, userOk := claims["userID"].(string)
	email, emailOk := claims["email"].(string)
	role, roleOk := claims["role"].(float64)

	if !userOk || !emailOk || !roleOk {
		return nil, ce.ErrMissingClaims
	}

	newToken, err := s.generateToken(userID, email, int(role), s.cfg.TokenValidSeconds)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateToken(
		userID,
		email,
		int(role),
		s.cfg.RefreshTokenValidSeconds,
	)
	if err != nil {
		return nil, err
	}

	res := &dto.TokenResponseDto{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}

	return res, nil
}

func (s *service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hashed passoword: %w", err)
	}

	return string(hash), nil
}

func (s *service) verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))

	return err == nil
}

func (s *service) generateToken(
	userID, email string,
	role, validDurationSeconds int,
) (string, error) {
	if userID == "" || email == "" || role == 0 {
		return "", fmt.Errorf(
			"%w: userID=%s, email=%s, role=%d",
			ce.ErrInvalidTokenData,
			userID,
			email,
			role,
		)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"email":  email,
		"role":   role,
		"exp":    time.Now().Add(time.Second * time.Duration(validDurationSeconds)).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to create SignedString from JWT token: %w", err)
	}

	return tokenStr, nil
}

func (s *service) verifyToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
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
