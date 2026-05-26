package repository

import (
	"context"
	"errors"
	"fmt"
	"med_book/internal/database"
	"med_book/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserBookStat struct {
	BookID         uuid.UUID `json:"book_id"`
	BookName       string    `json:"book_name"`
	TotalQuestions int       `json:"total_questions"`
	BestScore      int       `json:"best_score"`
	MaxScore       int       `json:"max_score"`
	AttemptsCount  int       `json:"attempts_count"`
	Percent        float64   `json:"percent"`
}

type SessionRepositoryInterface interface {
	// CRUD
	Create(session *model.Session) error
	GetByID(sessionID uuid.UUID) (*model.Session, error)
	Update(session *model.Session) error
	Delete(sessionID uuid.UUID) error

	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Session, error)

	GetUserBooksStats(ctx context.Context, userID uuid.UUID) ([]*UserBookStat, error)

	GetBookBySessionID(sessionID uuid.UUID) (*model.Book, error)
	GetNumberOfQuestions(ctx context.Context, bookID uuid.UUID) (int, error)
}

type SessionRepository struct {
	db *database.Database
}

func NewSessionRepository(db *database.Database) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *model.Session) error {
	return r.db.GetDB().Create(session).Error
}

func (r *SessionRepository) GetByID(sessionID uuid.UUID) (*model.Session, error) {
	var session model.Session
	result := r.db.GetDB().Where("id = ?", sessionID).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session with id: %v not found", sessionID)
		}
	}
	return &session, nil
}

func (r *SessionRepository) Update(session *model.Session) error {
	return r.db.GetDB().Save(session).Error
}

func (r *SessionRepository) Delete(sessionID uuid.UUID) error {
	return r.db.GetDB().Delete(&model.Session{}, "id = ?", sessionID).Error
}

func (r *SessionRepository) FindByUserID(userID uuid.UUID) (*model.Session, error) {
	var session model.Session
	result := r.db.GetDB().Where("user_id = ?", userID).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session with user_id: %v not found", userID)
		}
	}
	return &session, nil
}

func (r *SessionRepository) GetUserBestScoreForBook(ctx context.Context, userID, bookID uuid.UUID) (int, error) {
	var bestScore int

	err := r.db.GetDB().WithContext(ctx).
		Table("user_book_sessions").
		Where("user_id = ? AND book_id = ? AND status = ?",
			userID, bookID, model.SessionStatusCompleted).
		Select("COALESCE(MAX(score), 0)").
		Scan(&bestScore).Error

	return bestScore, err
}
