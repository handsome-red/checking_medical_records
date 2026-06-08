package service

import (
	"context"
	"errors"
	"med_book/internal/model"
	"med_book/internal/repository"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Register(ctx context.Context, email, firstName, lastName, patronymic, password string) (*model.User, error) {
	// Проверяем, существует ли пользователь
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, errors.New("user already exists")
	}

	// Создаём пользователя
	user, err := model.NewUser(firstName, lastName, patronymic, email, password)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if ok := user.CheckPassword(password); !ok {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (s *UserService) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]model.User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *UserService) PromoteToAdmin(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	return s.userRepo.SetRole(ctx, user.ID, model.RoleAdmin)
}
