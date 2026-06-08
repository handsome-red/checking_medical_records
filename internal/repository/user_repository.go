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
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	SetRole(ctx context.Context, userID uuid.UUID, role model.UserRole) error
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
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

func (ur *UserRepository) Create(ctx context.Context, user *model.User) error {
	result := ur.db.GetDB().Create(user)
	return result.Error
}

func (ur *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	result := ur.db.GetDB().WithContext(ctx).First(&user, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("User not found")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (ur *UserRepository) FindAll(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := ur.db.GetDB().WithContext(ctx).Order("last_name, first_name").Find(&users).Error
	return users, err
}

func (ur *UserRepository) SetRole(ctx context.Context, userID uuid.UUID, role model.UserRole) error {
	return ur.db.GetDB().WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("role", role).Error
}
