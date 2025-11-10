// Package service implements the core business logic of the application, such as authentication and user management.
package service

import (
	"context"
	"encoding/json"
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
	LogoutUser(ctx context.Context, token string) error
	Authorize(ctx context.Context, token string) (*dto.TokenClaimsDto, error)
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

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type generateTokenParams struct {
	UserID               string
	Email                string
	TokenType            string
	TokenVersion         int64
	Role                 int
	ValidDurationSeconds int
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

	newTokenVersion, err := s.repo.IncrementTokenVersionForUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to incerement token version: %w", err)
	}

	token, err := s.generateToken(generateTokenParams{
		UserID:               user.ID,
		Email:                user.Email,
		TokenType:            tokenTypeAccess,
		TokenVersion:         newTokenVersion,
		Role:                 user.Role,
		ValidDurationSeconds: s.cfg.TokenValidSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := s.generateToken(generateTokenParams{
		UserID:               user.ID,
		Email:                user.Email,
		TokenType:            tokenTypeRefresh,
		TokenVersion:         newTokenVersion,
		Role:                 user.Role,
		ValidDurationSeconds: s.cfg.RefreshTokenValidSeconds,
	})
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
	ctx context.Context,
	refreshToken string,
) (*dto.TokenResponseDto, error) {
	claimsDto, err := s.verifyToken(ctx, refreshToken, tokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	if claimsDto.UserID == "" || claimsDto.Email == "" || claimsDto.Role == 0 {
		return nil, ce.ErrMissingClaims
	}

	newTokenVersion, err := s.repo.IncrementTokenVersionForUser(ctx, claimsDto.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment token version: %w", err)
	}

	newToken, err := s.generateToken(generateTokenParams{
		UserID:               claimsDto.UserID,
		Email:                claimsDto.Email,
		TokenType:            tokenTypeAccess,
		TokenVersion:         newTokenVersion,
		Role:                 claimsDto.Role,
		ValidDurationSeconds: s.cfg.TokenValidSeconds,
	})
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateToken(generateTokenParams{
		UserID:               claimsDto.UserID,
		Email:                claimsDto.Email,
		TokenType:            tokenTypeRefresh,
		TokenVersion:         newTokenVersion,
		Role:                 claimsDto.Role,
		ValidDurationSeconds: s.cfg.TokenValidSeconds,
	})
	if err != nil {
		return nil, err
	}

	res := &dto.TokenResponseDto{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}

	return res, nil
}

func (s *service) LogoutUser(ctx context.Context, tokenStr string) error {
	claims, err := s.verifyToken(ctx, tokenStr, tokenTypeAccess)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}

	_, err = s.repo.IncrementTokenVersionForUser(ctx, claims.UserID)
	if err != nil {
		return fmt.Errorf("failed to increment token version: %w", err)
	}

	return nil
}

func (s *service) Authorize(ctx context.Context, tokenStr string) (*dto.TokenClaimsDto, error) {
	claimsDto, err := s.verifyToken(ctx, tokenStr, tokenTypeAccess)
	if err != nil {
		return nil, err
	}

	return claimsDto, nil
}

func (s *service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hashed password: %w", err)
	}

	return string(hash), nil
}

func (s *service) verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))

	return err == nil
}

func (s *service) generateToken(p generateTokenParams) (string, error) {
	if p.UserID == "" || p.Email == "" || p.Role == 0 {
		return "", fmt.Errorf(
			"%w: userID=%s, email=%s, role=%d",
			ce.ErrInvalidTokenData,
			p.UserID,
			p.Email,
			p.Role,
		)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":       p.UserID,
		"email":        p.Email,
		"role":         p.Role,
		"tokenVersion": p.TokenVersion,
		"tokenType":    p.TokenType,
		"exp":          time.Now().Add(time.Second * time.Duration(p.ValidDurationSeconds)).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to create SignedString from JWT token: %w", err)
	}

	return tokenStr, nil
}

func (s *service) verifyToken(
	ctx context.Context,
	tokenStr string,
	expectedType string,
) (*dto.TokenClaimsDto, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ce.ErrUnexpectedSigninMethod
		}

		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, ce.ErrParseToken
	}

	if !token.Valid {
		return nil, ce.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ce.ErrParseClaims
	}

	claimsDto, err := s.mapClaimsToDTO(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims into dto: %w", err)
	}

	if claimsDto.TokenType != expectedType {
		return nil, ce.ErrInvalidTokenType
	}

	dbTokenVersion, err := s.repo.GetTokenVersionByUserID(ctx, claimsDto.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's token version from db: %w", err)
	}

	if dbTokenVersion != claimsDto.TokenVersion {
		return nil, ce.ErrInvalidTokenVersion
	}

	return claimsDto, nil
}

func (s *service) mapClaimsToDTO(claims jwt.MapClaims) (*dto.TokenClaimsDto, error) {
	data, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal claims: %w", err)
	}

	var dto dto.TokenClaimsDto

	err = json.Unmarshal(data, &dto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return &dto, nil
}
