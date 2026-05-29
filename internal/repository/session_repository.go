// internal/repository/session_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"

	"med_book/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// ========== БАЗОВЫЕ CRUD ==========

// Create создаёт новую сессию
func (r *SessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// FindByID находит сессию по ID
func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	var session model.Session
	result := r.db.WithContext(ctx).
		Preload("User").
		Preload("Book").
		First(&session, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("session with id %v not found", id)
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}

// Update обновляет сессию
func (r *SessionRepository) Update(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// Delete удаляет сессию
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Session{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session with id %v not found", id)
	}
	return nil
}

// ========== ПОИСК ==========

// FindByUserID возвращает все сессии пользователя
func (r *SessionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Session, error) {
	var sessions []*model.Session
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to find sessions for user %v: %w", userID, result.Error)
	}
	return sessions, nil
}

// FindUnfinishedByUserID находит незавершённую сессию пользователя
func (r *SessionRepository) FindUnfinishedByUserID(ctx context.Context, userID uuid.UUID) (*model.Session, error) {
	var session model.Session
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND status IN ?", userID, []string{
			string(model.SessionStatusInProgress),
		}).
		Order("created_at DESC").
		First(&session)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil // нет незавершённой сессии
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find unfinished session: %w", result.Error)
	}
	return &session, nil
}

// FindLastCompletedByUserID находит последнюю завершённую сессию
func (r *SessionRepository) FindLastCompletedByUserID(ctx context.Context, userID uuid.UUID) (*model.Session, error) {
	var session model.Session
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.SessionStatusCompleted).
		Where("completed_at IS NOT NULL").
		Order("completed_at DESC").
		First(&session)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find last completed session: %w", result.Error)
	}
	return &session, nil
}

// ========== ТРАНЗАКЦИИ ==========

// SessionTx транзакционный репозиторий
type SessionTx struct {
	tx *gorm.DB
}

// Begin начинает транзакцию
func (r *SessionRepository) Begin(ctx context.Context) (*SessionTx, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &SessionTx{tx: tx}, nil
}

// FindByID в транзакции
func (tx *SessionTx) FindByID(sessionID uuid.UUID) (*model.Session, error) {
	var session model.Session
	result := tx.tx.First(&session, "id = ?", sessionID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &session, result.Error
}

// Update в транзакции
func (tx *SessionTx) Update(session *model.Session) error {
	return tx.tx.Save(session).Error
}

// CreateAnswer в транзакции
func (tx *SessionTx) CreateAnswer(answer *model.UserAnswer) error {
	return tx.tx.Create(answer).Error
}

// CountAnswers в транзакции
func (tx *SessionTx) CountAnswers(sessionID uuid.UUID) (int, error) {
	var count int64
	err := tx.tx.Model(&model.UserAnswer{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error
	return int(count), err
}

// Commit завершает транзакцию
func (tx *SessionTx) Commit() error {
	return tx.tx.Commit().Error
}

// Rollback откатывает транзакцию
func (tx *SessionTx) Rollback() error {
	return tx.tx.Rollback().Error
}

// GetUserBooksStats возвращает статистику по всем книгам пользователя одним запросом
func (r *SessionRepository) GetUserBooksStats(ctx context.Context, userID uuid.UUID) ([]*UserBookStat, error) {
	var stats []*UserBookStat

	query := `
        SELECT 
            b.id as book_id,
            b.title as book_name,
            COALESCE(MAX(s.score), 0) as best_score,
            COALESCE(q.question_count, 0) as max_score,
            COUNT(DISTINCT s.id) as attempts_count,
            ROUND(COALESCE(MAX(s.score), 0)::DECIMAL / NULLIF(q.question_count, 0) * 100, 2) as percent
        FROM books b
        CROSS JOIN LATERAL (
            SELECT COUNT(*) as question_count 
            FROM questions 
            WHERE book_id = b.id
        ) q
        LEFT JOIN user_book_sessions s ON s.book_id = b.id 
            AND s.user_id = $1 
            AND s.status = 'completed'
        GROUP BY b.id, b.title, q.question_count
        ORDER BY percent DESC NULLS LAST
    `

	err := r.db.WithContext(ctx).Raw(query, userID).Scan(&stats).Error
	return stats, err
}

type UserBookStat struct {
	BookID        uuid.UUID `json:"book_id"`
	BookName      string    `json:"book_name"`
	BestScore     int       `json:"best_score"`
	MaxScore      int       `json:"max_score"`
	Percent       float64   `json:"percent"`
	AttemptsCount int       `json:"attempts_count"`
}

// GetUserBestScoreForBook возвращает лучший результат пользователя по книге
func (r *SessionRepository) GetUserBestScoreForBook(ctx context.Context, userID, bookID uuid.UUID) (int, error) {
	var bestScore int
	err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("user_id = ? AND book_id = ? AND status = ?", userID, bookID, model.SessionStatusCompleted).
		Select("COALESCE(MAX(score), 0)").
		Scan(&bestScore).Error
	return bestScore, err
}

// GetUserAttempts возвращает количество попыток пользователя по книге
func (r *SessionRepository) GetUserAttempts(ctx context.Context, userID, bookID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("user_id = ? AND book_id = ? AND status = ?", userID, bookID, model.SessionStatusCompleted).
		Count(&count).Error
	return int(count), err
}
