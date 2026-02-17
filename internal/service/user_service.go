package service

import (
	"context"
	"errors"
	"strings"

	"github.com/dowglassantana/golang-with-dynamodb/internal/entity"
	"github.com/dowglassantana/golang-with-dynamodb/internal/repository"
)

var (
	ErrUserNotFound    = errors.New("usuario nao encontrado")
	ErrInvalidInput    = errors.New("name e email sao obrigatorios")
)

// UserService define o contrato de regras de negocio de usuarios.
type UserService interface {
	Create(ctx context.Context, input entity.CreateUserInput) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetAll(ctx context.Context) ([]entity.User, error)
	Update(ctx context.Context, id string, input entity.UpdateUserInput) error
	Delete(ctx context.Context, id string) error
}

type userServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userServiceImpl{repo: repo}
}

func (s *userServiceImpl) Create(ctx context.Context, input entity.CreateUserInput) (*entity.User, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Email) == "" {
		return nil, ErrInvalidInput
	}

	user := entity.NewUser(input.Name, input.Email)

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userServiceImpl) GetByID(ctx context.Context, id string) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userServiceImpl) GetAll(ctx context.Context) ([]entity.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *userServiceImpl) Update(ctx context.Context, id string, input entity.UpdateUserInput) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Email) == "" {
		return ErrInvalidInput
	}
	return s.repo.Update(ctx, id, input)
}

func (s *userServiceImpl) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
