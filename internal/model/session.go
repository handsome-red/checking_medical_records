package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_status,priority:1"`
	BookID uuid.UUID `gorm:"type:uuid;not null;index"`

	StartedAt        time.Time  `gorm:"default:now();index"`
	CompletedAt      *time.Time `gorm:"index"`
	ExpiresAt        *time.Time `gorm:"index"`
	TimeLimitMinutes int        `gorm:"default:0"`

	Status SessionStatus `gorm:"type:varchar(20);default:'in_progress';index:idx_user_status,priority:2"`

	Score    int `gorm:"default:0"`
	MaxScore int `gorm:"default:0;check:max_score >= 0"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Связи
	User    User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Book    Book         `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	Answers []UserAnswer `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE"`
}

type SessionStatus string

const (
	SessionStatusInProgress SessionStatus = "in_progress"
	SessionStatusCompleted  SessionStatus = "completed"
	SessionStatusAbandoned  SessionStatus = "abandoned"
	SessionStatusExpired    SessionStatus = "expired"
)

func (Session) TableName() string {
	return "user_book_sessions"
}

// NewSession создаёт новую сессию
func NewSession(userID, bookID uuid.UUID, duration time.Duration, maxScore int) *Session {
	expiresAt := time.Now().Add(duration)
	return &Session{
		UserID:           userID,
		BookID:           bookID,
		StartedAt:        time.Now(),
		ExpiresAt:        &expiresAt,
		TimeLimitMinutes: int(duration.Minutes()),
		MaxScore:         maxScore,
		Status:           SessionStatusInProgress,
		Score:            0,
	}
}

// NewUnlimitedSession создаёт сессию без ограничения по времени
func NewUnlimitedSession(userID, bookID uuid.UUID, maxScore int) *Session {
	return &Session{
		UserID:           userID,
		BookID:           bookID,
		StartedAt:        time.Now(),
		TimeLimitMinutes: 0,
		MaxScore:         maxScore,
		Status:           SessionStatusInProgress,
		Score:            0,
	}
}

// IsCompleted проверяет, завершена ли сессия
func (s *Session) IsCompleted() bool {
	return s.Status == SessionStatusCompleted
}

// IsInProgress проверяет, активна ли сессия
func (s *Session) IsInProgress() bool {
	return s.Status == SessionStatusInProgress
}

// IsAbandoned проверяет, брошена ли сессия
func (s *Session) IsAbandoned() bool {
	return s.Status == SessionStatusAbandoned
}

// IsExpired проверяет, истекло ли время
func (s *Session) IsExpired() bool {
	if s.Status == SessionStatusExpired {
		return true
	}
	if s.IsCompleted() {
		return false
	}
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// Complete завершает сессию
func (s *Session) Complete() error {
	if s.IsCompleted() {
		return errors.New("session already completed")
	}
	if s.IsExpired() {
		return errors.New("cannot complete expired session")
	}
	now := time.Now()
	s.CompletedAt = &now
	s.Status = SessionStatusCompleted
	return nil
}

// Abandon бросает сессию
func (s *Session) Abandon() error {
	if s.IsCompleted() {
		return errors.New("cannot abandon completed session")
	}
	if s.IsExpired() {
		return errors.New("cannot abandon expired session")
	}
	s.Status = SessionStatusAbandoned
	return nil
}

// Expire принудительно завершает по таймауту
func (s *Session) Expire() error {
	if s.IsCompleted() {
		return errors.New("cannot expire completed session")
	}
	s.Status = SessionStatusExpired
	now := time.Now()
	s.CompletedAt = &now
	return nil
}

// AddPoints добавляет баллы
func (s *Session) AddPoints(points int) error {
	if s.IsCompleted() {
		return errors.New("cannot add points to completed session")
	}
	if s.IsExpired() {
		return errors.New("cannot add points to expired session")
	}
	if points < 0 {
		return errors.New("points cannot be negative")
	}
	s.Score += points
	if s.Score > s.MaxScore {
		s.Score = s.MaxScore
	}
	return nil
}

// ScoreProgress возвращает прогресс по баллам (0-100)
func (s *Session) ScoreProgress() float64 {
	if s.MaxScore == 0 {
		return 0
	}
	return float64(s.Score) / float64(s.MaxScore) * 100
}

// IsPassed проверяет, пройден ли тест (порог 70%)
func (s *Session) IsPassed() bool {
	return s.ScoreProgress() >= 70.0
}

// Duration возвращает длительность сессии
func (s *Session) Duration() time.Duration {
	end := time.Now()
	if s.CompletedAt != nil {
		end = *s.CompletedAt
	}
	return end.Sub(s.StartedAt)
}

// GetTimeRemaining возвращает оставшееся время
func (s *Session) GetTimeRemaining() time.Duration {
	if s.IsCompleted() || s.IsExpired() {
		return 0
	}
	if s.ExpiresAt == nil {
		return 0
	}
	remaining := time.Until(*s.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetTimeRemainingFormatted возвращает форматированное время "MM:SS"
func (s *Session) GetTimeRemainingFormatted() string {
	remaining := s.GetTimeRemaining()
	if remaining <= 0 {
		return "00:00"
	}
	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// IsExpiringSoon проверяет, истекает ли сессия в ближайшие N минут
func (s *Session) IsExpiringSoon(minutes int) bool {
	if s.IsCompleted() || s.IsExpired() {
		return false
	}
	if s.ExpiresAt == nil {
		return false
	}
	remaining := s.GetTimeRemaining()
	return remaining > 0 && remaining <= time.Duration(minutes)*time.Minute
}

// BeforeCreate GORM хук
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.Status == "" {
		s.Status = SessionStatusInProgress
	}

	// Если есть TimeLimitMinutes, но нет ExpiresAt
	if s.TimeLimitMinutes > 0 && s.ExpiresAt == nil {
		expiresAt := time.Now().Add(time.Duration(s.TimeLimitMinutes) * time.Minute)
		s.ExpiresAt = &expiresAt
	}

	return s.Validate()
}

// BeforeUpdate GORM хук
func (s *Session) BeforeUpdate(tx *gorm.DB) error {
	// Запрещаем изменение MaxScore
	if tx.Statement.Changed("MaxScore") {
		return errors.New("max_score cannot be changed after creation")
	}

	// Запрещаем изменение StartedAt
	if tx.Statement.Changed("StartedAt") {
		return errors.New("started_at cannot be changed")
	}

	return nil
}

// Validate валидирует сессию
func (s *Session) Validate() error {
	if s.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if s.BookID == uuid.Nil {
		return errors.New("book_id is required")
	}
	if s.MaxScore < 0 {
		return errors.New("max_score cannot be negative")
	}
	if s.TimeLimitMinutes < 0 {
		return errors.New("time_limit_minutes cannot be negative")
	}
	if !s.Status.IsValid() {
		return errors.New("invalid status")
	}
	return nil
}

// IsValid проверяет валидность статуса
func (s SessionStatus) IsValid() bool {
	switch s {
	case SessionStatusInProgress, SessionStatusCompleted, SessionStatusAbandoned, SessionStatusExpired:
		return true
	}
	return false
}
