package model

import (
	"time"

	"github.com/google/uuid"
)

type UserAnswer struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	SessionID  uint      `gorm:"not null;index"`           // ссылка на сессию (INT для производительности)
	QuestionID uuid.UUID `gorm:"type:uuid;not null;index"` // ссылка на вопрос
	OptionID   uuid.UUID `gorm:"type:uuid;not null"`       // выбранный вариант ответа
	IsCorrect  bool      `gorm:"default:false;index"`      // правильный ли ответ (дублируем для скорости)
	Points     int       `gorm:"default:0"`                // сколько баллов получил
	AnsweredAt time.Time `gorm:"default:now();index"`      // когда ответил

	Session  Session  `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
	Question Question `gorm:"foreignKey:QuestionID"`
	Option   Option   `gorm:"foreignKey:OptionID"`
}

func (UserAnswer) TableName() string {
	return "user_answers"
}
