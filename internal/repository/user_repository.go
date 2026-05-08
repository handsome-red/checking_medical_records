package repository

import (
	"med_book/internal/model"

	"github.com/google/uuid"
)

type UserRepositoryInterface interface {
	Create(user *model.User) error
	FindByID(id uuid.UUID) (*model.User, error)
	FindByFIO(firstName, lastName, patronomyc string) (*model.User, error)
}
