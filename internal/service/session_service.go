// internal/service/session.go
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"med_book/internal/model"
	"med_book/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	sessionDuration  = time.Hour
	numberOfQuestion = 5
)

type SessionServiceInterface struct {
	sessionRepo *repository.SessionRepository
	bookRepo    *repository.BookRepository
	answerRepo  *repository.AnswerRepository
}

type SessionService struct {
	sessionRepo *repository.SessionRepository
	bookRepo    *repository.BookRepository
	answerRepo  *repository.AnswerRepository
}

func NewSessionService(
	sessionRepo *repository.SessionRepository,
	bookRepo *repository.BookRepository,
	answerRepo *repository.AnswerRepository,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		bookRepo:    bookRepo,
		answerRepo:  answerRepo,
	}
}

// ========== СОЗДАНИЕ СЕССИИ ==========

// CreateSession создаёт новую сессию с книгами и вопросами
func (s *SessionService) CreateSession(userID, bookID uuid.UUID) (*model.Session, error) {
	// 1. Создаём сессию
	session := s

	if err := s.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 2. Сохраняем книги и вопросы
	var allQuestions []model.SessionQuestion
	globalIndex := 0

	for bookIdx, bookID := range bookIDs {
		// Сохраняем книгу
		sessionBook := &model.SessionBook{
			SessionID:  session.ID,
			BookID:     bookID,
			OrderIndex: bookIdx,
		}
		if err := s.repo.CreateSessionBook(sessionBook); err != nil {
			return nil, fmt.Errorf("failed to create session book: %w", err)
		}

		// Получаем вопросы из книги
		book, err := s.testService.GetBook(bookID)
		if err != nil {
			return nil, err
		}

		selectedQuestions := s.selectRandomQuestions(book.Questions, numberOfQuestion)

		// Сохраняем вопросы
		for _, question := range selectedQuestions {
			sessionQuestion := &model.SessionQuestion{
				SessionID:  session.ID,
				BookID:     bookID,
				QuestionID: question.ID,
				OrderIndex: globalIndex,
			}
			if err := s.repo.CreateSessionQuestion(sessionQuestion); err != nil {
				return nil, fmt.Errorf("failed to create session question: %w", err)
			}
			allQuestions = append(allQuestions, *sessionQuestion)
			globalIndex++
		}
	}

	// 3. Создаём прогресс (начинаем с первого вопроса первой книги)
	if len(allQuestions) == 0 {
		return nil, errors.New("no questions in session")
	}

	progress := &model.SessionProgress{
		SessionID:         session.ID,
		CurrentBookIndex:  0,
		CurrentQIndex:     0,
		CurrentBookID:     bookIDs[0],
		CurrentQuestionID: allQuestions[0].QuestionID,
		CompletedAt:       nil,
	}

	if err := s.repo.CreateProgress(progress); err != nil {
		return nil, fmt.Errorf("failed to create progress: %w", err)
	}

	return session, nil
}

func (s *SessionService) selectRandomQuestions(questions []model.Question, count int) []model.Question {
	if count <= 0 || len(questions) == 0 {
		return []model.Question{}
	}

	// Если запрошено больше, чем есть вопросов - берём все
	if count >= len(questions) {
		count = len(questions)
	}

	shuffled := make([]model.Question, len(questions))
	copy(shuffled, questions)
	rand.NewSource(time.Now().UnixNano())
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

// ========== ПОЛУЧЕНИЕ ТЕКУЩИХ ДАННЫХ ==========

func (s *SessionService) GetAllSessions() ([]*model.Session, error) {
	return s.repo.GetAllSessions()
}

// GetSessionByID возвращает сессию по ID
func (s *SessionService) GetSessionByID(sessionID uuid.UUID) (*model.Session, error) {
	return s.repo.FindByID(sessionID)
}

// GetUserSessions возвращает все сессии пользователя
func (s *SessionService) GetUserSessions(userID uuid.UUID) ([]*model.Session, error) {
	return s.repo.FindByUserID(userID)
}

// GetCurrentQuestionID возвращает ID текущего вопроса
func (s *SessionService) GetCurrentQuestionID(sessionID uuid.UUID) (int, error) {
	progress, err := s.repo.GetProgress(sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get progress: %w", err)
	}

	return progress.CurrentQuestionID, nil
}

// GetCurrentQuestion возвращает полный объект текущего вопроса
func (s *SessionService) GetCurrentQuestion(sessionID uuid.UUID) (*model.Question, error) {
	questionID, err := s.GetCurrentQuestionID(sessionID)
	if err != nil {
		return nil, err
	}

	return s.testService.GetQuestion(questionID)
}

// GetUserAttempts возвращает количество попыток прохождения книги пользователем
func (s *SessionService) GetUserAttempts(userID uuid.UUID, bookID int) (int, error) {
	book, err := s.testService.GetBook(bookID)
	if err != nil {
		return 0, fmt.Errorf("book")
	}

	return s.testService.GetUserAttempts(userID, bookID)
}

// GetCurrentProgress возвращает текущий прогресс
// func (s *SessionService) GetCurrentProgress(sessionID uuid.UUID) (*model.SessionProgress, error) {
// 	return s.repo.GetProgress(sessionID)
// }

// GetSessionQuestions возвращает все вопросы сессии
func (s *SessionService) GetSessionQuestions(sessionID uuid.UUID) ([]model.SessionQuestion, error) {
	return s.repo.FindQuestionsBySessionID(sessionID)
}

// GetSessionBooks возвращает все книги сессии
func (s *SessionService) GetSessionBooks(sessionID uuid.UUID) ([]model.SessionBook, error) {
	return s.repo.FindBooksBySessionID(sessionID)
}

// ========== НАВИГАЦИЯ ==========

// MoveToNextQuestion перемещает к следующему вопросу
func (s *SessionService) MoveToNextQuestion(sessionID uuid.UUID) error {
	// 1. Получаем прогресс
	progress, err := s.repo.GetProgress(sessionID)
	if err != nil {
		return err
	}

	// 2. Получаем все вопросы
	questions, err := s.repo.FindQuestionsBySessionID(sessionID)
	if err != nil {
		return err
	}

	totalQuestions := len(questions)

	// 3. Проверяем, не последний ли вопрос
	if progress.CurrentQIndex+1 >= totalQuestions {
		// Тест закончен
		return s.CompleteSession(sessionID)
	}

	// 4. Переходим к следующему вопросу
	progress.CurrentQIndex++
	progress.CurrentQuestionID = questions[progress.CurrentQIndex].QuestionID

	// Обновляем книгу, если нужно (если книга изменилась)
	if questions[progress.CurrentQIndex].BookID != progress.CurrentBookID {
		progress.CurrentBookID = questions[progress.CurrentQIndex].BookID
		progress.CurrentBookIndex++
	}

	return s.repo.UpdateProgress(progress)
}

// MoveToPreviousQuestion перемещает к предыдущему вопросу (если нужно)
func (s *SessionService) MoveToPreviousQuestion(sessionID uuid.UUID) error {
	progress, err := s.repo.GetProgress(sessionID)
	if err != nil {
		return err
	}

	if progress.CurrentQIndex == 0 {
		return errors.New("already at first question")
	}

	questions, err := s.repo.FindQuestionsBySessionID(sessionID)
	if err != nil {
		return err
	}

	progress.CurrentQIndex--
	progress.CurrentQuestionID = questions[progress.CurrentQIndex].QuestionID

	return s.repo.UpdateProgress(progress)
}

// ========== ОТВЕТЫ ==========

// SaveAnswer сохраняет ответ пользователя
func (s *SessionService) SaveAnswer(sessionID uuid.UUID, questionID, answerID int) error {
	answer := &model.SessionAnswer{
		SessionID:  sessionID,
		QuestionID: questionID,
		AnswerID:   answerID,
	}
	return s.repo.CreateAnswer(answer)
}

// GetAnswersBySession возвращает все ответы пользователя
func (s *SessionService) GetAnswersBySession(sessionID uuid.UUID) ([]model.SessionAnswer, error) {
	return s.repo.FindAnswersBySessionID(sessionID)
}

// GetAnswerByQuestion возвращает ответ на конкретный вопрос
func (s *SessionService) GetAnswerByQuestion(sessionID uuid.UUID, questionID int) (*model.SessionAnswer, error) {
	answers, err := s.repo.FindAnswersBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	for _, answer := range answers {
		if answer.QuestionID == questionID {
			return &answer, nil
		}
	}

	return nil, errors.New("answer not found")
}

// ========== ЗАВЕРШЕНИЕ СЕССИИ ==========

// CompleteSession завершает сессию
func (s *SessionService) CompleteSession(sessionID uuid.UUID) error {
	progress, err := s.repo.GetProgress(sessionID)
	if err != nil {
		return err
	}

	now := time.Now()
	progress.CompletedAt = &now

	return s.repo.UpdateProgress(progress)
}

// IsSessionCompleted проверяет, завершена ли сессия
func (s *SessionService) IsSessionCompleted(sessionID uuid.UUID) (bool, error) {
	progress, err := s.repo.GetProgress(sessionID)
	if err != nil {
		return false, err
	}

	return progress.CompletedAt != nil, nil
}

// ========== ПРОГРЕСС ==========

// GetProgress возвращает прогресс
func (s *SessionService) GetProgress(sessionID uuid.UUID) (*model.SessionProgress, error) {
	return s.repo.GetProgress(sessionID)
}

// GetProgressPercentage возвращает процент прохождения теста
func (s *SessionService) GetProgressPercentage(sessionID uuid.UUID) (float64, error) {
	questions, err := s.repo.FindQuestionsBySessionID(sessionID)
	if err != nil {
		return 0, err
	}

	answers, err := s.repo.FindAnswersBySessionID(sessionID)
	if err != nil {
		return 0, err
	}

	total := len(questions)
	if total == 0 {
		return 0, nil
	}

	return float64(len(answers)) / float64(total) * 100, nil
}

// ========== ЛИМИТ ==========

// CanStartNewTest проверяет, может ли пользователь начать новый тест сегодня
func (s *SessionService) CanStartNewTest(userID uuid.UUID) (bool, error) {
	// Проверяем последний завершённый тест
	_, progress, err := s.repo.FindLastCompletedByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil // нет завершённых тестов
		}
		return false, err
	}

	// Если тест завершён сегодня - нельзя
	if progress.CompletedAt != nil && isSameDay(*progress.CompletedAt, time.Now()) {
		return false, nil
	}

	return true, nil
}

// GetNextAvailableTime возвращает время следующей доступной попытки
func (s *SessionService) GetNextAvailableTime(userID uuid.UUID) (time.Time, error) {
	_, progress, err := s.repo.FindLastCompletedByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	if progress.CompletedAt == nil {
		return time.Time{}, nil
	}

	return progress.CompletedAt.Add(24 * time.Hour), nil
}

// HasUnfinishedSession проверяет, есть ли у пользователя незавершённая сессия
func (s *SessionService) HasUnfinishedSession(userID uuid.UUID) (bool, *model.Session, error) {
	session, err := s.repo.FindUnfinishedByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, session, nil
}

func isSameDay(t time.Time, ref time.Time) bool {
	return t.Year() == ref.Year() && t.YearDay() == ref.YearDay()
}
