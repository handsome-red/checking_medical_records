package repository

import (
	"context"
	"errors"
	"fmt"
	"med_book/internal/database"
	"med_book/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookRepositoryInterface interface {
	// CRUD
	Create(book *model.Book) error
	GetByID(bookID uuid.UUID) (*model.Book, error)
	Update(book *model.Book) error
	Delete(bookID uuid.UUID) error

	GetAllBooks(ctx context.Context) ([]*model.Book, error)

	FindByID(ctx context.Context, bookID uuid.UUID) (*model.Book, error)
}

type BookRepository struct {
	db *database.Database
}

func NewBookRepository(db *database.Database) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(book *model.Book) error {
	return r.db.GetDB().Create(book).Error
}

func (r *BookRepository) GetByID(bookID uuid.UUID) (*model.Book, error) {
	var book model.Book
	result := r.db.GetDB().Where("id = ?", bookID).First(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("book with id: %v not found", bookID)
		}
	}
	return &book, nil
}

func (r *BookRepository) Update(book *model.Book) error {
	return r.db.GetDB().Save(book).Error
}

func (r *BookRepository) Delete(bookID uuid.UUID) error {
	return r.db.GetDB().Delete(&model.Book{}, "id = ?", bookID).Error
}

func (r *BookRepository) GetAllBooks(ctx context.Context) ([]*model.Book, error) {
	var books []*model.Book
	result := r.db.GetDB().WithContext(ctx).Find(&books)
	if result.Error != nil {
		return nil, result.Error
	}
	return books, nil
}
