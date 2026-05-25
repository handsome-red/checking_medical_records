package model

import (
	"time"

	"github.com/google/uuid"
)

type Question struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BookID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Text        string    `gorm:"not null;type:text"` // текст вопроса
	Explanation string    `gorm:"type:text"`          // объяснение после ответа
	SortOrder   int       `gorm:"default:0;index"`    // порядок вопроса в тесте
	Points      int       `gorm:"default:1"`          // баллы за вопрос
	CreatedAt   time.Time `gorm:"default:now()"`
	UpdatedAt   time.Time `gorm:"default:now()"`

	// Связи
	Book    Book         `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	Options []Option     `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
	Answers []UserAnswer `gorm:"foreignKey:QuestionID"`
}

func (Question) TableName() string {
	return "questions"
}
