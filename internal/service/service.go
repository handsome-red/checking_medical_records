package service

import "github.com/google/uuid"

type User struct {
	ID 			uuid.UUID
	FirstName 	string
	LastName	string
	Patronymic	string
}

func NewUser(firstName, lastName, patronymic string) *User {
	return &User{
		FirstName:	firstName,
		LastName:	lastName,
		Patronymic: patronymic,
	}
}