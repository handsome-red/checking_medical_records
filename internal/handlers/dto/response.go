package dto

import (
	"med_book/internal/service"

	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronomyc string `json:"patronymic"`
	Password   string `json:"password"`
}

type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Patronymic string    `json:"patronymic"`
}

func ToUserResponse(user *service.User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Patronymic: user.Patronymic,
	}
}
