package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	BookID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	StartedAt   time.Time  `gorm:"default:now();index"`
	CompletedAt *time.Time // NULL если ещё не завершён
	Status      string     `gorm:"type:varchar(20);default:'in_progress';index"` // in_progress, completed, abandoned
	Score       int        `gorm:"default:0"`                                    // итоговый балл
	MaxScore    int        `gorm:"default:0"`                                    // максимальный возможный балл
	CreatedAt   time.Time  `gorm:"default:now()"`
	UpdatedAt   time.Time  `gorm:"default:now()"`

	// Связи
	User    User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Book    Book         `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	Answers []UserAnswer `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
}

func (Session) TableName() string {
	return "user_book_sessions"
}
