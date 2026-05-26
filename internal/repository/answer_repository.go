package repository

import "med_book/internal/database"

type AnswerRepositoryInterface interface {
}

type AnswerRepository struct {
	db *database.Database
}

func NewAnswerRepository(db *database.Database) *AnswerRepository {
	return &AnswerRepository{db: db}
}

// func Create
