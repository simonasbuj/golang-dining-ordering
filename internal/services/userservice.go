package services

import (
	"context"
	db "golang-dining-ordering/internal/db/generated"
	"golang-dining-ordering/internal/repository"
)

type UserService interface {
	CreateUser(ctx context.Context) (*db.User, error)
}

type userService struct {
	repo repository.UsersRepository
}

func NewUserService(repo repository.UsersRepository) *userService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context) (*db.User, error) {
	return s.repo.CreateUser(ctx)
}
