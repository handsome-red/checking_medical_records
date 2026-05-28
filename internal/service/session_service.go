// internal/service/session_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"med_book/internal/model"
	"med_book/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

// StartTest начинает новый тест с конкретной книгой
func (s *SessionService) StartTest(ctx context.Context, userID, bookID uuid.UUID, duration time.Duration) (*model.Session, error) {
	// 1. Проверяем, нет ли незавершённой сессии
	unfinished, err := s.sessionRepo.FindUnfinishedByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check unfinished session: %w", err)
	}
	if unfinished != nil {
		return nil, errors.New("you have an unfinished session")
	}

	// 2. Проверяем лимит (один тест в день)
	canStart, err := s.CanStartNewTest(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !canStart {
		nextTime, _ := s.GetNextAvailableTime(ctx, userID)
		return nil, fmt.Errorf("you can start a new test only after %s", nextTime.Format("15:04 02.01.2006"))
	}

	// 3. Получаем максимальный балл за книгу
	maxScore, err := s.bookRepo.GetMaxScore(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get max score: %w", err)
	}

	// 4. Создаём сессию
	var session *model.Session
	if duration > 0 {
		session = model.NewSession(userID, bookID, duration, maxScore)
	} else {
		session = model.NewUnlimitedSession(userID, bookID, maxScore)
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ========== ПОЛУЧЕНИЕ ДАННЫХ ==========

// GetSessionByID возвращает сессию по ID
func (s *SessionService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*model.Session, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Проверяем истечение времени
	if session.IsExpired() {
		if err := session.Expire(); err == nil {
			s.sessionRepo.Update(ctx, session)
		}
	}

	return session, nil
}

// GetUserSessions возвращает все сессии пользователя
func (s *SessionService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error) {
	return s.sessionRepo.FindByUserID(ctx, userID)
}

// GetCurrentQuestion возвращает текущий неотвеченный вопрос
func (s *SessionService) GetCurrentQuestion(ctx context.Context, sessionID uuid.UUID) (*model.Question, error) {
	// 1. Получаем сессию
	session, err := s.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 2. Проверяем статус
	if session.IsCompleted() {
		return nil, errors.New("session already completed")
	}
	if session.IsExpired() {
		return nil, errors.New("session expired")
	}

	// 3. Получаем已回答 вопросы
	answeredQuestions, err := s.answerRepo.GetAnsweredQuestionIDs(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 4. Получаем все вопросы книги
	questions, err := s.bookRepo.GetQuestionsWithOptions(ctx, session.BookID)
	if err != nil {
		return nil, err
	}

	// 5. Находим первый неотвеченный вопрос
	for _, q := range questions {
		if !containsUUID(answeredQuestions, q.ID) {
			return &q, nil
		}
	}

	// Все вопросы已回答 - завершаем сессию
	if err := s.CompleteSession(ctx, sessionID); err != nil {
		return nil, err
	}

	return nil, errors.New("all questions answered")
}

// GetAnswersBySession возвращает все ответы пользователя
func (s *SessionService) GetAnswersBySession(ctx context.Context, sessionID uuid.UUID) ([]model.UserAnswer, error) {
	return s.answerRepo.FindBySessionID(ctx, sessionID)
}

// ========== ОТВЕТЫ ==========

// SubmitAnswer сохраняет ответ и обновляет счёт
func (s *SessionService) SubmitAnswer(ctx context.Context, sessionID uuid.UUID, questionID, optionID uuid.UUID) error {
	// Начинаем транзакцию
	tx, err := s.sessionRepo.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Получаем сессию
	session, err := tx.FindByID(sessionID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Проверяем статус
	if session.IsCompleted() {
		tx.Rollback()
		return errors.New("session already completed")
	}
	if session.IsExpired() {
		tx.Rollback()
		return errors.New("session expired")
	}

	// 3. Проверяем, не отвечал ли уже на этот вопрос
	alreadyAnswered, err := s.answerRepo.ExistsBySessionAndQuestion(ctx, sessionID, questionID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if alreadyAnswered {
		tx.Rollback()
		return errors.New("question already answered")
	}

	// 4. Получаем вариант ответа
	option, err := s.bookRepo.GetOptionByID(ctx, optionID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 5. Создаём ответ
	answer := &model.UserAnswer{
		SessionID:  sessionID,
		QuestionID: questionID,
		OptionID:   optionID,
		IsCorrect:  option.IsCorrect,
		AnsweredAt: time.Now(),
	}

	if option.IsCorrect {
		question, _ := s.bookRepo.GetQuestionByID(ctx, questionID)
		if question != nil {
			answer.Points = question.Points
		}
	}

	if err := tx.CreateAnswer(answer); err != nil {
		tx.Rollback()
		return err
	}

	// 6. Обновляем счёт сессии
	if option.IsCorrect {
		if err := session.AddPoints(answer.Points); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Update(session); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 7. Проверяем, все ли вопросы已回答
	answeredCount, err := tx.CountAnswers(sessionID)
	if err != nil {
		tx.Rollback()
		return err
	}

	totalQuestions, err := s.bookRepo.CountQuestions(ctx, session.BookID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if answeredCount >= totalQuestions {
		if err := session.Complete(); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Update(session); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// ========== ЗАВЕРШЕНИЕ СЕССИИ ==========

// CompleteSession завершает сессию
func (s *SessionService) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if err := session.Complete(); err != nil {
		return err
	}

	return s.sessionRepo.Update(ctx, session)
}

// AbandonSession бросает сессию
func (s *SessionService) AbandonSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if err := session.Abandon(); err != nil {
		return err
	}

	return s.sessionRepo.Update(ctx, session)
}

// ========== ЛИМИТЫ ==========

// CanStartNewTest проверяет, может ли пользователь начать новый тест сегодня
func (s *SessionService) CanStartNewTest(ctx context.Context, userID uuid.UUID) (bool, error) {
	lastSession, err := s.sessionRepo.FindLastCompletedByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}

	if lastSession == nil {
		return true, nil
	}

	// Если тест завершён сегодня - нельзя
	if lastSession.CompletedAt != nil && isSameDay(*lastSession.CompletedAt, time.Now()) {
		return false, nil
	}

	return true, nil
}

// GetNextAvailableTime возвращает время следующей доступной попытки
func (s *SessionService) GetNextAvailableTime(ctx context.Context, userID uuid.UUID) (time.Time, error) {
	lastSession, err := s.sessionRepo.FindLastCompletedByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	if lastSession == nil || lastSession.CompletedAt == nil {
		return time.Time{}, nil
	}

	return lastSession.CompletedAt.Add(24 * time.Hour), nil
}

// HasUnfinishedSession проверяет, есть ли у пользователя незавершённая сессия
func (s *SessionService) HasUnfinishedSession(ctx context.Context, userID uuid.UUID) (bool, *model.Session, error) {
	session, err := s.sessionRepo.FindUnfinishedByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return session != nil, session, nil
}

// ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func containsUUID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}
