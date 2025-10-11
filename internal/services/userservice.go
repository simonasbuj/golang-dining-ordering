package services

import (
	"context"
	db "golang-dining-ordering/internal/db/generated"
	"golang-dining-ordering/internal/dto"
	"golang-dining-ordering/internal/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (*db.User, error)
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

func (s *userService) CreateUser(ctx context.Context, req *dto.SignUpRequestDto) (*db.User, error) {
	return s.repo.CreateUser(ctx, req)
}

func (s *userService) SignInUser(ctx context.Context, req *dto.SignInRequestDto) (*dto.SignInResponseDto, error) {
	res := &dto.SignInResponseDto{
		Token:        "some-fake-token",
		RefreshToken: "some-fake-refresh-token",
	}
	return res, nil
}
