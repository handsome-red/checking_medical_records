// internal/model/session.go
package model

// import (
// 	"time"

// 	"github.com/google/uuid"
// )

// // Session - основная сессия
// type Session struct {
// 	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
// 	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
// 	CreatedAt time.Time
// 	ExpiresAt time.Time
// }

// func (Session) TableName() string {
// 	return "sessions"
// }

// // session_progress - текущее состояние (1:1 с session)
// type SessionProgress struct {
// 	SessionID         uuid.UUID `gorm:"primary_key"`
// 	CurrentBookIndex  int       `gorm:"default:0"`
// 	CurrentQIndex     int       `gorm:"default:0"`
// 	CurrentBookID     int       `gorm:"default:0"` // денормализация
// 	CurrentQuestionID int       `gorm:"default:0"` // денормализация
// 	CompletedAt       *time.Time
// }

// func (SessionProgress) TableName() string {
// 	return "session_progresses"
// }

// // SessionBook - книги, которые пользователь проходит в этой сессии
// type SessionBook struct {
// 	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
// 	SessionID  uuid.UUID `gorm:"type:uuid;index;not null"`
// 	BookID     int       `gorm:"not null"`
// 	OrderIndex int       `gorm:"not null"`
// }

// func (SessionBook) TableName() string {
// 	return "session_books"
// }

// // SessionQuestion - вопросы в рамках сессии и книги
// type SessionQuestion struct {
// 	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
// 	SessionID  uuid.UUID `gorm:"type:uuid;index;not null"`
// 	BookID     int       `gorm:"not null"`
// 	QuestionID int       `gorm:"not null"`
// 	OrderIndex int       `gorm:"not null"`
// }

// func (SessionQuestion) TableName() string {
// 	return "session_questions"
// }

// // SessionAnswer - ответы пользователя на вопросы
// type SessionAnswer struct {
// 	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
// 	SessionID  uuid.UUID `gorm:"type:uuid;index;not null"`
// 	QuestionID int       `gorm:"not null"`
// 	AnswerID   int       `gorm:"not null"`
// }

// func (SessionAnswer) TableName() string {
// 	return "session_answers"
// }
