package model

import "github.com/google/uuid"

type Book struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title string    `gorm:"not null;index"`
	Pages []BookPage
}

type BookPage struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BookID      uuid.UUID `gorm:"type:uuid;not null;index"`
	PageNumber  int       `gorm:"not null"`
	ImagePath   string    `gorm:"not null"` // /uploads/books/{book_id}/page_1.jpg
	ImageURL    string    `gorm:"not null"` // /static/books/{book_id}/page_1.jpg
	ContentType string    `gorm:"size:100"` // image/jpeg
	Book        Book      `gorm:"foreignKey:BookID"`
}

func (Book) TableName() string {
	return "books"
}
