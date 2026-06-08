package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id" gorm:"type:text;primaryKey"`
	FirstName    string    `json:"first_name" db:"first_name" gorm:"not null;size:50"`
	LastName     string    `json:"last_name" db:"last_name" gorm:"not null;size:50"`
	Patronymic   string    `json:"patronymic" db:"patronymic" gorm:"not null;size:50"`
	Email        string    `json:"email" db:"email" gorm:"not null;size:50;unique"`
	PasswordHash string    `json:"-" db:"password_hash" gorm:"not null"`
	Role         UserRole  `gorm:"type:varchar(20);default:'user'"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at,omitempty" db:"updated_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

func NewUser(firstName, lastName, patronymic, email, password string) (*User, error) {
	if len(firstName) == 0 || len(lastName) == 0 || len(patronymic) == 0 || len(email) == 0 {
		return nil, errors.New("invalid name")
	}

	if len(password) < 6 {
		return nil, errors.New("too short password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New(),
		FirstName:    firstName,
		LastName:     lastName,
		Patronymic:   patronymic,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	ensureUUID(&u.ID)
	return nil
}
