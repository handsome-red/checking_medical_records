package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Question struct {
	ID          uuid.UUID `gorm:"type:text;primaryKey"`
	BookID      uuid.UUID `gorm:"type:text;not null;index"`
	Text        string    `gorm:"not null;type:text"` // текст вопроса
	Explanation string    `gorm:"type:text"`          // объяснение после ответа
	SortOrder   int       `gorm:"default:0;index"`    // порядок вопроса в тесте
	Points      int       `gorm:"default:1"`          // баллы за вопрос
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`

	// Связи
	Book    Book         `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	Options []Option     `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
	Answers []UserAnswer `gorm:"foreignKey:QuestionID"`
}

func (Question) TableName() string {
	return "questions"
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
	ensureUUID(&q.ID)
	return nil
}
