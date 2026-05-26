package repository

import (
	"context"
	"errors"
	"med_book/internal/database"
	"med_book/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(user *model.User) error
	FindByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByFIO(fisrtName, lastName, patronymic string) (*model.User, error)
}

func (ur *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	result := ur.db.GetDB().Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("User not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

type UserRepository struct {
	db *database.Database
}

func NewUserRepository(db *database.Database) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) Create(user *model.User) error {
	result := ur.db.GetDB().Create(user)
	return result.Error
}

func (ur *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	result := ur.db.GetDB().First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("User not found")
	}
	return &user, nil
}

func (ur *UserRepository) FindByFIO(fisrtName, lastName, patronymic string) (*model.User, error) {
	var user model.User
	result := ur.db.GetDB().
		Where("first_name = ?", fisrtName).
		Where("first_name = ?", fisrtName).
		Where("first_name = ?", fisrtName).
		Find(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("User not found")
	}

	return &user, nil
}
