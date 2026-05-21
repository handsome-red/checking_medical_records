package repository

import (
	"errors"
	"med_book/internal/database"
	"med_book/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepositoryInterface interface {
	CreateSession(session *model.Session) error
	FindByID(id uuid.UUID) (*model.Session, error)
	FindByUserID(userID uuid.UUID) ([]*model.Session, error)

	Update(session *model.Session) error
	Delete(id uuid.UUID) error

	CreateSessionQuestion(question *model.SessionQuestion) error
	CreateSessionBook(book *model.SessionBook) error

	GetAllSessions() ([]*model.Session, error)

	FindQuestionsBySessionID(sessionID uuid.UUID) ([]model.SessionQuestion, error)
	FindQuestionByOrder(sessionID uuid.UUID, orderIndex int) (*model.SessionQuestion, error)
	CountQuestions(sessionID uuid.UUID) (int64, error)

	FindBooksBySessionID(sessionID uuid.UUID) ([]model.SessionBook, error)

	GetProgress(sessionID uuid.UUID) (*model.SessionProgress, error)
	UpdateProgress(progress *model.SessionProgress) error
	CreateProgress(progress *model.SessionProgress) error

	FindUnfinishedByUserID(userID uuid.UUID) (*model.Session, error)
	FindLastCompletedByUserID(userID uuid.UUID) (*model.Session, *model.SessionProgress, error)

	// Операции с ответами
	CreateAnswer(answer *model.SessionAnswer) error
	FindAnswersBySessionID(sessionID uuid.UUID) ([]model.SessionAnswer, error)

	// Транзакции
	Begin() TransactionInterface
}

type SessionRepository struct {
	db *database.Database
}

func NewSessionRepository(db *database.Database) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(session *model.Session) error {
	return r.db.GetDB().Create(session).Error
}

func (us *SessionRepository) FindByID(id uuid.UUID) (*model.Session, error) {
	var session model.Session
	result := us.db.GetDB().First(&session, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("Session not found")
	}
	return &session, nil
}

// FindByUserID возвращает все сессии пользователя
func (r *SessionRepository) FindByUserID(userID uuid.UUID) ([]*model.Session, error) {
	var sessions []*model.Session
	err := r.db.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

func (us *SessionRepository) Update(session *model.Session) error {
	result := us.db.GetDB().Save(session)
	return result.Error
}

func (r *SessionRepository) Begin() TransactionInterface {
	tx := r.db.GetDB().Begin()
	return NewSessionTxRepository(tx)
}

func (r *SessionRepository) Delete(id uuid.UUID) error {
	return r.db.GetDB().Delete(&model.Session{}, id).Error
}

func (r *SessionRepository) CreateSessionQuestion(question *model.SessionQuestion) error {
	return r.db.GetDB().Create(question).Error
}

func (r *SessionRepository) CreateSessionBook(book *model.SessionBook) error {
	return r.db.GetDB().Create(book).Error
}

func (r *SessionRepository) FindQuestionsBySessionID(sessionID uuid.UUID) ([]model.SessionQuestion, error) {
	var questions []model.SessionQuestion
	result := r.db.GetDB().Where("session_id = ?", sessionID).Order("order_index ASC").Find(&questions)
	return questions, result.Error
}

func (r *SessionRepository) FindQuestionByOrder(sessionID uuid.UUID, orderIndex int) (*model.SessionQuestion, error) {
	var question model.SessionQuestion
	result := r.db.GetDB().Where("session_id = ? AND order_index = ?", sessionID, orderIndex).First(&question)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("question not found")
	}
	return &question, result.Error
}

func (r *SessionRepository) CountQuestions(sessionID uuid.UUID) (int64, error) {
	var count int64
	result := r.db.GetDB().Model(&model.SessionQuestion{}).Where("session_id = ?", sessionID).Count(&count)
	return count, result.Error
}

func (r *SessionRepository) FindBooksBySessionID(sessionID uuid.UUID) ([]model.SessionBook, error) {
	var sessionBooks []model.SessionBook
	result := r.db.GetDB().
		Where("session_id = ?", sessionID).
		Order("order_index ASC").
		Find(&sessionBooks)

	return sessionBooks, result.Error
}

func (r *SessionRepository) CreateAnswer(answer *model.SessionAnswer) error {
	return r.db.GetDB().Create(answer).Error
}

func (r *SessionRepository) FindAnswersBySessionID(sessionID uuid.UUID) ([]model.SessionAnswer, error) {
	var answers []model.SessionAnswer
	result := r.db.GetDB().Where("session_id = ?", sessionID).Find(&answers)
	return answers, result.Error
}

func (r *SessionRepository) GetProgress(sessionID uuid.UUID) (*model.SessionProgress, error) {
	var progress model.SessionProgress
	result := r.db.GetDB().First(&progress, "session_id = ?", sessionID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("progress not found")
	}
	return &progress, result.Error
}

func (r *SessionRepository) UpdateProgress(progress *model.SessionProgress) error {
	return r.db.GetDB().Save(progress).Error
}

func (r *SessionRepository) CreateProgress(progress *model.SessionProgress) error {
	return r.db.GetDB().Create(progress).Error
}

// FindUnfinishedByUserID находит незавершённую сессию пользователя
func (r *SessionRepository) FindUnfinishedByUserID(userID uuid.UUID) (*model.Session, error) {
	var session model.Session

	err := r.db.GetDB().
		Table("sessions").
		Select("sessions.*").
		Joins("JOIN session_progresses ON session_progresses.session_id::uuid = sessions.id"). // 👈 привели к uuid
		Where("sessions.user_id = ? AND session_progresses.completed_at IS NULL", userID).
		First(&session).Error

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// FindLastCompletedByUserID находит последнюю завершённую сессию пользователя
func (r *SessionRepository) FindLastCompletedByUserID(userID uuid.UUID) (*model.Session, *model.SessionProgress, error) {
	var session model.Session
	var progress model.SessionProgress

	err := r.db.GetDB().
		Table("sessions").
		Select("sessions.*").
		Joins("JOIN session_progresses ON session_progresses.session_id::uuid = sessions.id"). // 👈 привели к uuid
		Where("sessions.user_id = ? AND session_progresses.completed_at IS NOT NULL", userID).
		Order("session_progresses.completed_at DESC").
		First(&session).Error

	if err != nil {
		return nil, nil, err
	}

	// Получаем прогресс для этой сессии
	err = r.db.GetDB().Where("session_id = ?", session.ID).First(&progress).Error
	if err != nil {
		return nil, nil, err
	}

	return &session, &progress, nil
}

func (r *SessionRepository) GetAllSessions() ([]*model.Session, error) {
	var sessions []*model.Session

	result := r.db.GetDB().
		Model(&model.Session{}).
		Find(&sessions)

	if result.Error != nil {
		return nil, result.Error
	}

	return sessions, nil
}
