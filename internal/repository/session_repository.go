// internal/repository/session_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// CountAnswers в транзакции — количество отвеченных вопросов
func (tx *SessionTx) CountAnsweredQuestions(sessionID uuid.UUID) (int, error) {
	var count int64
	err := tx.tx.Model(&model.UserAnswer{}).
		Where("session_id = ?", sessionID).
		Distinct("question_id").
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

// GetUserBooksStats возвращает статистику по всем книгам пользователя
func (r *SessionRepository) GetUserBooksStats(ctx context.Context, userID uuid.UUID) ([]*UserBookStat, error) {
	var stats []*UserBookStat

	query := `
		SELECT
			b.id AS book_id,
			b.title AS book_name,
			COALESCE(MAX(s.score), 0) AS best_score,
			COALESCE(q.question_count, 0) AS max_score,
			COALESCE(COUNT(DISTINCT s.id), 0) AS attempts_count,
			ROUND(
				COALESCE(CAST(MAX(s.score) AS REAL), 0) /
				NULLIF(CAST(q.question_count AS REAL), 0) * 100,
				1
			) AS percent
		FROM books b
		LEFT JOIN (
			SELECT book_id, COUNT(*) AS question_count
			FROM questions
			GROUP BY book_id
		) q ON q.book_id = b.id
		LEFT JOIN user_book_sessions s
			ON s.book_id = b.id
			AND s.user_id = ?
			AND s.status IN ('completed', 'expired')
		GROUP BY b.id, b.title, q.question_count
		ORDER BY b.title
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
		Where("user_id = ? AND book_id = ? AND status IN ?", userID, bookID, []model.SessionStatus{
			model.SessionStatusCompleted,
			model.SessionStatusExpired,
		}).
		Count(&count).Error
	return int(count), err
}

type AdminSessionReport struct {
	SessionID   uuid.UUID           `json:"session_id"`
	LastName    string              `json:"last_name"`
	FirstName   string              `json:"first_name"`
	Patronymic  string              `json:"patronymic"`
	BookTitle   string              `json:"book_title"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt *time.Time          `json:"completed_at"`
	Score       int                 `json:"score"`
	MaxScore    int                 `json:"max_score"`
	Answers     []AdminAnswerReport `json:"answers"`
}

type AdminTextReport struct {
	Text       string `json:"text"`
	IsCorrect  bool   `json:"is_correct"`
	IsSelected bool   `json:"is_selected"`
}

type AdminAnswerReport struct {
	QuestionText  string            `json:"question_text"`
	SelectedTexts []AdminTextReport `json:"selected_texts"`
	// IsCorrect     bool              `json:"is_all_correct"`
}

func (r *SessionRepository) GetAdminReports(ctx context.Context, userID, bookID *uuid.UUID) ([]AdminSessionReport, error) {
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("Book").
		Preload("Answers").
		Preload("Answers.Question").
		Preload("Answers.Option").
		Where("status IN ?", []model.SessionStatus{
			model.SessionStatusCompleted,
			model.SessionStatusExpired,
		}).
		Order("started_at DESC")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if bookID != nil {
		query = query.Where("book_id = ?", *bookID)
	}

	var sessions []model.Session
	if err := query.Find(&sessions).Error; err != nil {
		return nil, err
	}

	reports := make([]AdminSessionReport, 0, len(sessions))
	for _, session := range sessions {
		report := AdminSessionReport{
			SessionID:   session.ID,
			LastName:    session.User.LastName,
			FirstName:   session.User.FirstName,
			Patronymic:  session.User.Patronymic,
			BookTitle:   session.Book.Title,
			StartedAt:   session.StartedAt,
			CompletedAt: session.CompletedAt,
			Score:       session.Score,
			MaxScore:    session.MaxScore,
		}

		answersByQuestion := make(map[uuid.UUID][]model.UserAnswer)
		for _, answer := range session.Answers {
			answersByQuestion[answer.QuestionID] = append(answersByQuestion[answer.QuestionID], answer)
		}

		for questionID, userAnswers := range answersByQuestion {
			if len(userAnswers) == 0 {
				continue
			}
			questionText := userAnswers[0].Question.Text

			var correctOptions []model.Option
			r.db.WithContext(ctx).
				Where("question_id = ? AND is_correct = ?", questionID, true).
				Order("sort_order").
				Find(&correctOptions)

			answersMap := make(map[string]AdminTextReport)
			for _, opt := range correctOptions {
				answersMap[opt.Text] = AdminTextReport{
					Text:       opt.Text,
					IsCorrect:  true,
					IsSelected: false,
				}
			}

			for _, ua := range userAnswers {

				isCorrect := ua.Option.IsCorrect

				if isCorrect {
					if report, exist := answersMap[ua.Option.Text]; exist {
						report.IsSelected = true
						answersMap[ua.Option.Text] = report
					}
				} else {
					answersMap[ua.Option.Text] = AdminTextReport{
						Text:       ua.Option.Text,
						IsCorrect:  false,
						IsSelected: true,
					}
				}
			}

			allAnswers := make([]AdminTextReport, 0, len(answersMap))
			for _, answer := range answersMap {
				allAnswers = append(allAnswers, answer)
			}

			report.Answers = append(report.Answers, AdminAnswerReport{
				QuestionText:  questionText,
				SelectedTexts: allAnswers,
			})
		}

		reports = append(reports, report)
	}

	return reports, nil
}
