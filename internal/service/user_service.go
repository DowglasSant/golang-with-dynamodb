package service

import (
	"context"
	"errors"

	"github.com/dowglassantana/golang-with-dynamodb/internal/entity"
	"github.com/dowglassantana/golang-with-dynamodb/internal/repository"
	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("usuario nao encontrado")

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, input entity.CreateUserInput) (*entity.User, error) {
	user := entity.NewUser(input.Name, input.Email)
	user.ID = uuid.New().String()

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetAll(ctx context.Context) ([]entity.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) Update(ctx context.Context, id string, input entity.UpdateUserInput) error {
	return s.repo.Update(ctx, id, input)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
