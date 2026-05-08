package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewSession(userID uuid.UUID, duration time.Duration) *Session {
	return &Session{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
