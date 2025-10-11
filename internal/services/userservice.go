package services

import (
	"context"

	"golang-dining-ordering/internal/dto"
	"golang-dining-ordering/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error)
	SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.SignInResponseDto, error)
}

type userService struct {
	repo repository.UsersRepository
}

func NewUserService(repo repository.UsersRepository) *userService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (string, error) {
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return "", err
	}

	req.Password = hashedPassword

	return s.repo.CreateUser(ctx, req)
}

func (s *userService) SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.SignInResponseDto, error) {
	res := &dto.SignInResponseDto{
		Token:        "some-fake-token",
		RefreshToken: "some-fake-refresh-token",
	}
	return res, nil
}

func (s *userService) hashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

func (s *userService) verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
