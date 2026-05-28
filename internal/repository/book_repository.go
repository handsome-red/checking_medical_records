// internal/repository/book_repository.go
package repository

import (
	"context"
	"med_book/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{db: db}
}

// FindByID находит книгу по ID
func (r *BookRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	var book model.Book
	err := r.db.WithContext(ctx).
		Preload("Pages").
		First(&book, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetAllBooks возвращает все книги
func (r *BookRepository) GetAllBooks(ctx context.Context) ([]model.Book, error) {
	var books []model.Book
	err := r.db.WithContext(ctx).Find(&books).Error
	return books, err
}

// GetMaxScore возвращает максимальный балл за книгу
func (r *BookRepository) GetMaxScore(ctx context.Context, bookID uuid.UUID) (int, error) {
	var maxScore int
	err := r.db.WithContext(ctx).
		Model(&model.Question{}).
		Where("book_id = ?", bookID).
		Select("COALESCE(SUM(points), 0)").
		Scan(&maxScore).Error
	return maxScore, err
}

// CountQuestions возвращает количество вопросов в книге
func (r *BookRepository) CountQuestions(ctx context.Context, bookID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Question{}).
		Where("book_id = ?", bookID).
		Count(&count).Error
	return int(count), err
}

// GetQuestionsWithOptions возвращает все вопросы книги с вариантами ответов
func (r *BookRepository) GetQuestionsWithOptions(ctx context.Context, bookID uuid.UUID) ([]model.Question, error) {
	var questions []model.Question
	err := r.db.WithContext(ctx).
		Preload("Options").
		Where("book_id = ?", bookID).
		Order("sort_order ASC").
		Find(&questions).Error
	return questions, err
}

// GetQuestionByID возвращает вопрос по ID
func (r *BookRepository) GetQuestionByID(ctx context.Context, questionID uuid.UUID) (*model.Question, error) {
	var question model.Question
	err := r.db.WithContext(ctx).
		Preload("Options").
		First(&question, "id = ?", questionID).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// GetOptionByID возвращает вариант ответа по ID
func (r *BookRepository) GetOptionByID(ctx context.Context, optionID uuid.UUID) (*model.Option, error) {
	var option model.Option
	err := r.db.WithContext(ctx).First(&option, "id = ?", optionID).Error
	if err != nil {
		return nil, err
	}
	return &option, nil
}

// GetPagesByBookID возвращает страницы книги
func (r *BookRepository) GetPagesByBookID(ctx context.Context, bookID uuid.UUID) ([]model.BookPage, error) {
	var pages []model.BookPage
	err := r.db.WithContext(ctx).
		Where("book_id = ?", bookID).
		Order("page_number ASC").
		Find(&pages).Error
	return pages, err
}
