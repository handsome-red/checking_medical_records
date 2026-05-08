package service

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	FirstName  string
	LastName   string
	Patronymic string
	Password   string // добавили поле
}

func NewUser(firstName, lastName, patronymic, password string) (*User, error) {
	// Базовая валидация
	if firstName == "" || lastName == "" {
		return nil, errors.New("first name and last name are required")
	}
	if len(password) < 6 {
		return nil, errors.New("password too short")
	}

	return &User{
		ID:         uuid.New(),
		FirstName:  firstName,
		LastName:   lastName,
		Patronymic: patronymic,
		Password:   password,
	}, nil
}
