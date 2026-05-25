package service

import (
	"errors"
	"med_book/internal/model"
	"med_book/internal/repository"
	"strings"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) RegisterUser(email, firstName, lastName, patronymic, password string) (*model.User, error) {
	email = strings.TrimSpace(email)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if firstName == "" || lastName == "" || email == "" {
		return nil, errors.New("invalid name")
	}

	if len(password) < 6 {
		return nil, errors.New("too short password")
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		Patronymic:   patronymic,
		PasswordHash: password,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(id uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) FindByFIO(fisrtName, lastName, patronymic string) (*model.User, error) {
	return s.userRepo.FindByFIO(fisrtName, lastName, patronymic)
}

func (s *UserService) FindByEmail(email string) (*model.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *UserService) Authenticate(email, password string) (*model.User, error) {
	user, err := s.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if ok := user.CheckPassword(password); !ok {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}
