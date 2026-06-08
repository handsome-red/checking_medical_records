package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Option struct {
	ID         uuid.UUID `gorm:"type:text;primaryKey"`
	QuestionID uuid.UUID `gorm:"type:text;not null;index"`
	Text       string    `gorm:"not null;type:text"`  // текст варианта ответа
	IsCorrect  bool      `gorm:"default:false;index"` // правильный ли ответ
	SortOrder  int       `gorm:"default:0"`           // порядок отображения

	// Связи
	Question Question     `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
	Answers  []UserAnswer `gorm:"foreignKey:OptionID"`
}

func (Option) TableName() string {
	return "options"
}

func (o *Option) BeforeCreate(tx *gorm.DB) error {
	ensureUUID(&o.ID)
	return nil
}
