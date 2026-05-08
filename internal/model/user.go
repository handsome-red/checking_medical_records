package model

import (
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FirstName    string    `json:"first_name" db:"first_name" gorm:"not null;size:50"`
	LastName     string    `json:"last_name" db:"last_name" gorm:"not null;size:50"`
	Patronomyc   string    `json:"patronomyc" db:"patronomyc" gorm:"not null;size:50"`
	PasswordHash string    `json:"-" db:"password_hash" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at,omitempty" db:"updated_at" gorm:"autoUpdateTime"`
}

func NewUser(firstName, lastName, patronomyc, password string) (*User, error) {
	if len(firstName) == 0 || len(lastName) == 0 || len(patronomyc) == 0 {
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
		Patronomyc:   patronomyc,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
