package services

import (
	"context"
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
	SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.SignInResponseDto, error)
}

type authService struct {
	secret	string
	repo 	repository.UsersRepository
}

func NewAuthService(secret string, repo repository.UsersRepository) *authService {
	return &authService{
		secret: secret,
		repo: 	repo,
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

func (s *authService) SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.SignInResponseDto, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err

	}

	ok := s.verifyPassword(req.Password, user.PasswordHash)
	if !ok {
		return nil, ce.UnauthorizedError
	}

	token, err := s.generateToken(user.ID, user.Email, user.Role, 24 * 7)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user.ID, user.Email, user.Role, 24 * 14)
	if err != nil {
		return nil, err
	}

	res := &dto.SignInResponseDto{
		Token:        token,
		RefreshToken: refreshToken,
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
		"userID": 		userID,
		"email": 		email,
		"role":			role,
		"expiresAt":	time.Now().Add(time.Hour * time.Duration(validDurationHours)),
	})

	tokenStr, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
