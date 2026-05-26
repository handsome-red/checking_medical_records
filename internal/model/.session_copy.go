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
// type Session struct {
// 	SessionID         uuid.UUID `gorm:"primary_key"`
// 	CurrentBookIndex  int       `gorm:"default:0"`
// 	CurrentQIndex     int       `gorm:"default:0"`
// 	CurrentBookID     int       `gorm:"default:0"` // денормализация
// 	CurrentQuestionID int       `gorm:"default:0"` // денормализация
// 	CompletedAt       *time.Time
// }

// func (Session) TableName() string {
// 	return "session_progresses"
// }

// // Book - книги, которые пользователь проходит в этой сессии
// type Book struct {
// 	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
// 	SessionID  uuid.UUID `gorm:"type:uuid;index;not null"`
// 	BookID     int       `gorm:"not null"`
// 	OrderIndex int       `gorm:"not null"`
// }

// func (Book) TableName() string {
// 	return "session_books"
// }

// // Question - вопросы в рамках сессии и книги
// type Question struct {
// 	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
// 	SessionID  uuid.UUID `gorm:"type:uuid;index;not null"`
// 	BookID     int       `gorm:"not null"`
// 	QuestionID int       `gorm:"not null"`
// 	OrderIndex int       `gorm:"not null"`
// }

// func (Question) TableName() string {
// 	return "session_questions"
// }

// // UserAnswer - ответы пользователя на вопросы
// type UserAnswer struct {
// 	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
// 	SessionID  uuid.UUID `gorm:"type:uuid;index;not null"`
// 	QuestionID int       `gorm:"not null"`
// 	AnswerID   int       `gorm:"not null"`
// }

// func (UserAnswer) TableName() string {
// 	return "session_answers"
// }
