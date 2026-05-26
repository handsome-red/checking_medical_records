package model

import "github.com/google/uuid"

// model/book.go
type Book struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title    string    `gorm:"not null;size:200;index"`
	BookPath string    `gorm:"size:500"`

	// Связи
	Pages     []BookPage `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	Images    []Image    `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	Questions []Question `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	Sessions  []Session  `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
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
