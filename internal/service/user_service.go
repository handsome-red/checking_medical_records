package service

import (
	"errors"
	"github.com/google/uuid"
	"strings"
)

type UserService struct {
	users map[string]User // email или имя как ключ
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]User),
	}
}

func (s *UserService) RegisterUser(firstName, lastName, patronymic, password string) (*User, error) {
	// Валидация
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if firstName == "" || lastName == "" {
		return nil, errors.New("invalid name")
	}

	if len(password) < 6 {
		return nil, errors.New("too short password")
	}

	// Проверка на существование (простая - по имени+фамилии)
	key := firstName + "_" + lastName
	if _, exists := s.users[key]; exists {
		return nil, errors.New("user with this name already exists")
	}

	// Создаём пользователя
	user := &User{
		ID:         uuid.New(),
		FirstName:  firstName,
		LastName:   lastName,
		Patronymic: patronymic,
		Password:   password, // В продакшене хешируйте!
	}

	// Сохраняем
	s.users[key] = *user

	return user, nil
}

func (s *UserService) GetUser(id uuid.UUID) (*User, error) {
	for _, user := range s.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}
