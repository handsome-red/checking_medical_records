package service

import (
	"errors"
	"med_book/internal/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SessionService struct {
	sessions map[uuid.UUID]*model.Session
	mu       sync.RWMutex
}

func NewSessionService() *SessionService {
	s := make(map[uuid.UUID]*model.Session)

	return &SessionService{
		sessions: s,
	}
}

func (s *SessionService) CreateSession(userID uuid.UUID, duration time.Duration) (*model.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session := model.NewSession(userID, duration)

	s.sessions[session.ID] = session
	return session, nil
}

func (s *SessionService) GetSession(sessionID uuid.UUID) (*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exist := s.sessions[sessionID]
	if !exist {
		return nil, errors.New("session is not exist")
	}

	return session, nil
}

func (s *SessionService) GetSessionByUserID(userID uuid.UUID) (*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

}
