package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	ce "golang-dining-ordering/internal/customerrors"
	"golang-dining-ordering/internal/dto"
	"golang-dining-ordering/internal/repository"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error)
	SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.TokenResponseDto, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponseDto, error)
}

type authService struct {
	secret string
	repo   repository.UsersRepository
}

func NewAuthService(secret string, repo repository.UsersRepository) *authService {
	return &authService{
		secret: secret,
		repo:   repo,
	}
}

func (s *authService) CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error) {
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
		return nil, err

	}

	ok := s.verifyPassword(req.Password, user.PasswordHash)
	if !ok {
		return nil, ce.UnauthorizedError
	}

	token, err := s.generateToken(user.ID, user.Email, user.Role, 24*7)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user.ID, user.Email, user.Role, 24*14)
	if err != nil {
		return nil, err
	}

	res := &dto.TokenResponseDto{
		Token:        token,
		RefreshToken: refreshToken,
	}
	return res, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponseDto, error) {

	claims, err := s.verifyToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID := fmt.Sprintf("%v", claims["userID"])
	email := fmt.Sprintf("%v", claims["email"])
	role := fmt.Sprintf("%v", claims["role"])

	newToken, err := s.generateToken(userID, email, role, 24*7)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateToken(userID, email, role, 24*14)
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
		return "", err
	}
	return string(hash), nil
}

func (s *authService) verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (s *authService) generateToken(userID, email, role string, validDurationHours int) (string, error) {
	if userID == "" || email == "" || role == "" {
		return "", fmt.Errorf("generating token failed, because one of these is invalid - userID: %s, email: %s, role %s", userID, email, role)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"email":  email,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * time.Duration(validDurationHours)).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s *authService) verifyToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse claims")
	}

	return claims, nil
}
