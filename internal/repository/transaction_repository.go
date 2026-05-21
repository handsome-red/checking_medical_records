package repository

import (
	"errors"
	"med_book/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionInterface interface {
	// Основные операции с сессией
	CreateSession(session *model.Session) error
	FindSessionByID(id uuid.UUID) (*model.Session, error)
	UpdateSession(session *model.Session) error

	// Операции с вопросами
	CreateSessionQuestion(question *model.SessionQuestion) error

	// Операции с ответами
	CreateAnswer(answer *model.SessionAnswer) error

	// Операции с прогрессом
	GetProgress(sessionID uuid.UUID) (*model.SessionProgress, error)
	CreateProgress(progress *model.SessionProgress) error
	UpdateProgress(progress *model.SessionProgress) error

	// Управление транзакцией
	Commit() error
	Rollback() error
}

type SessionTxRepository struct {
	tx *gorm.DB
}

func NewSessionTxRepository(tx *gorm.DB) *SessionTxRepository {
	return &SessionTxRepository{tx: tx}
}

func (r *SessionTxRepository) CreateSession(session *model.Session) error {
	return r.tx.Create(session).Error
}

func (r *SessionTxRepository) FindSessionByID(id uuid.UUID) (*model.Session, error) {
	var session model.Session
	err := r.tx.First(&session, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("session not found")
	}
	return &session, err
}

func (r *SessionTxRepository) UpdateSession(session *model.Session) error {
	return r.tx.Save(session).Error
}

func (r *SessionTxRepository) CreateSessionQuestion(question *model.SessionQuestion) error {
	return r.tx.Create(question).Error
}

func (r *SessionTxRepository) CreateAnswer(answer *model.SessionAnswer) error {
	return r.tx.Create(answer).Error
}

func (r *SessionTxRepository) GetProgress(sessionID uuid.UUID) (*model.SessionProgress, error) {
	var progress model.SessionProgress
	err := r.tx.First(&progress, "session_id = ?", sessionID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("progress not found")
	}
	return &progress, err
}

func (r *SessionTxRepository) CreateProgress(progress *model.SessionProgress) error {
	return r.tx.Create(progress).Error
}

func (r *SessionTxRepository) UpdateProgress(progress *model.SessionProgress) error {
	return r.tx.Save(progress).Error
}

func (r *SessionTxRepository) Commit() error {
	return r.tx.Commit().Error
}

func (r *SessionTxRepository) Rollback() error {
	return r.tx.Rollback().Error
}
