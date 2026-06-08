package model

import (
	"time"

	"github.com/google/uuid"
)

type UserAnswer struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	SessionID  uuid.UUID `gorm:"type:text;not null;index"`
	QuestionID uuid.UUID `gorm:"type:text;not null;index"`
	OptionID   uuid.UUID `gorm:"type:text;not null"`
	IsCorrect  bool      `gorm:"default:false;index"`      // правильный ли ответ (дублируем для скорости)
	Points     int       `gorm:"default:0"`                // сколько баллов получил
	AnsweredAt time.Time `gorm:"autoCreateTime;index"`

	Session  Session  `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
	Question Question `gorm:"foreignKey:QuestionID"`
	Option   Option   `gorm:"foreignKey:OptionID"`
}

func (UserAnswer) TableName() string {
	return "user_answers"
}
