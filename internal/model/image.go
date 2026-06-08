package model

import "github.com/google/uuid"

type Image struct {
	ID        uuid.UUID `gorm:"type:text;primaryKey"`
	BookID    uuid.UUID `gorm:"type:text;not null;index"`
	ImagePath string    `gorm:"not null;size:500"` // /uploads/books/{book_id}/img_1.jpg
	// ImageURL    string    `gorm:"not null;size:500"` // /static/books/{book_id}/img_1.jpg
	ContentType string `gorm:"size:100"` // image/jpeg, image/png
	FileSize    int64  // размер в байтах
	SortOrder   int    `gorm:"default:0"` // порядок отображения

	// Связь
	Book Book `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
}

func (Image) TableName() string {
	return "images"
}

func (i *Image) GetImageURL() string {
	return "/static/" + i.ImagePath
}
