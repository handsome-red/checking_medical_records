// internal/service/book_service.go
package service

import (
	"context"

	"med_book/internal/model"
	"med_book/internal/repository"

	"github.com/google/uuid"
)

type BookService struct {
	bookRepo *repository.BookRepository
}

func NewBookService(bookRepo *repository.BookRepository) *BookService {
	return &BookService{
		bookRepo: bookRepo,
	}
}

// GetAllBooks возвращает все книги
func (s *BookService) GetAllBooks(ctx context.Context) ([]model.Book, error) {
	return s.bookRepo.GetAllBooks(ctx)
}

// GetBookByID возвращает книгу по ID
func (s *BookService) GetBookByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	return s.bookRepo.FindByID(ctx, id)
}

// GetPagesByBookID возвращает страницы книги
func (s *BookService) GetPagesByBookID(ctx context.Context, bookID uuid.UUID) ([]model.BookPage, error) {
	return s.bookRepo.GetPagesByBookID(ctx, bookID)
}

// GetQuestionsWithOptions возвращает вопросы с вариантами ответов
func (s *BookService) GetQuestionsWithOptions(ctx context.Context, bookID uuid.UUID) ([]model.Question, error) {
	return s.bookRepo.GetQuestionsWithOptions(ctx, bookID)
}

// CountQuestions возвращает количество вопросов в книге
func (s *BookService) CountQuestions(ctx context.Context, bookID uuid.UUID) (int, error) {
	return s.bookRepo.CountQuestions(ctx, bookID)
}

// GetMaxScore возвращает максимальный балл за книгу
func (s *BookService) GetMaxScore(ctx context.Context, bookID uuid.UUID) (int, error) {
	return s.bookRepo.GetMaxScore(ctx, bookID)
}
