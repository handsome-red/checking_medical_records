// internal/service/session_service.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
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

// StartTest начинает новый тест с конкретной книгой
func (s *SessionService) StartTest(ctx context.Context, userID, bookID uuid.UUID, duration time.Duration) (*model.Session, error) {
	unfinished, err := s.sessionRepo.FindUnfinishedByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check unfinished session: %w", err)
	}
	if unfinished != nil {
		return nil, errors.New("you have an unfinished session")
	}

	maxScore, err := s.bookRepo.GetMaxScore(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get max score: %w", err)
	}

	questions, err := s.bookRepo.GetQuestionsWithOptions(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to load questions: %w", err)
	}
	if len(questions) == 0 {
		return nil, errors.New("book has no questions")
	}

	order := make([]uuid.UUID, len(questions))
	for i, q := range questions {
		order[i] = q.ID
	}
	rand.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to encode question order: %w", err)
	}

	var session *model.Session
	if duration > 0 {
		session = model.NewSession(userID, bookID, duration, maxScore)
	} else {
		session = model.NewUnlimitedSession(userID, bookID, maxScore)
	}
	session.QuestionOrder = string(orderJSON)

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *SessionService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*model.Session, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsExpired() && session.IsInProgress() {
		if err := session.Expire(); err == nil {
			_ = s.recalculateSessionScore(ctx, session)
			_ = s.sessionRepo.Update(ctx, session)
		}
	}

	return session, nil
}

func (s *SessionService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error) {
	return s.sessionRepo.FindByUserID(ctx, userID)
}

func (s *SessionService) GetCurrentQuestion(ctx context.Context, sessionID uuid.UUID) (*model.Question, error) {
	session, err := s.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsCompleted() || session.Status == model.SessionStatusExpired {
		return nil, errors.New("session already completed")
	}
	if session.IsExpired() {
		return nil, errors.New("session expired")
	}

	answeredQuestions, err := s.answerRepo.GetAnsweredQuestionIDs(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	order, err := s.parseQuestionOrder(session)
	if err != nil {
		return nil, err
	}

	for _, questionID := range order {
		if !containsUUID(answeredQuestions, questionID) {
			question, err := s.bookRepo.GetQuestionByID(ctx, questionID)
			if err != nil {
				return nil, err
			}
			return question, nil
		}
	}

	if err := s.CompleteSession(ctx, sessionID); err != nil {
		return nil, err
	}

	return nil, errors.New("all questions answered")
}

func (s *SessionService) GetAnswersBySession(ctx context.Context, sessionID uuid.UUID) ([]model.UserAnswer, error) {
	return s.answerRepo.FindBySessionID(ctx, sessionID)
}

func (s *SessionService) GetUserBooksStats(ctx context.Context, userID uuid.UUID) ([]*repository.UserBookStat, error) {
	return s.sessionRepo.GetUserBooksStats(ctx, userID)
}

// SubmitQuestionAnswer сохраняет все выбранные варианты для одного вопроса
func (s *SessionService) SubmitQuestionAnswer(
    ctx context.Context,
    sessionID uuid.UUID,
    questionID uuid.UUID,
    optionIDs []uuid.UUID,
) error {
    log.Printf("SubmitQuestionAnswer START at %v", time.Now())
    
    // Замеряем каждый шаг
    startTime := time.Now()
    stepStart := time.Now()
    
    tx, err := s.sessionRepo.Begin(ctx)
    if err != nil {
        return err
    }
    log.Printf("  [1] Begin tx: %v", time.Since(stepStart))
    defer tx.Rollback()

    stepStart = time.Now()
    session, err := tx.FindByID(sessionID)
    if err != nil {
        return err
    }
    log.Printf("  [2] FindByID: %v", time.Since(stepStart))

    if session == nil {
        return errors.New("session not found")
    }

    stepStart = time.Now()
    // Сохраняем ответы
    now := time.Now()
    for i, optionID := range optionIDs {
        answer := &model.UserAnswer{
            SessionID:  sessionID,
            QuestionID: questionID,
            OptionID:   optionID,
            AnsweredAt: now,
        }
        if err := tx.CreateAnswer(answer); err != nil {
            return err
        }
        log.Printf("  [3.%d] Save answer: %v", i, time.Since(stepStart))
    }

    stepStart = time.Now()
    answeredCount, err := tx.CountAnsweredQuestions(sessionID)
    if err != nil {
        return err
    }
    log.Printf("  [4] CountAnsweredQuestions: %v (count=%d)", time.Since(stepStart), answeredCount)

    stepStart = time.Now()
    session.AnsweredCount = answeredCount
    if err := tx.Update(session); err != nil {
        return err
    }
    log.Printf("  [5] Update session: %v", time.Since(stepStart))

    stepStart = time.Now()
    if err := tx.Commit(); err != nil {
        return err
    }
    log.Printf("  [6] Commit: %v", time.Since(stepStart))

    log.Printf("SubmitQuestionAnswer END - TOTAL: %v", time.Since(startTime))
    return nil
}

func (s *SessionService) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if err := session.Complete(); err != nil {
		return err
	}

	if err := s.recalculateSessionScore(ctx, session); err != nil {
		return err
	}

	return s.sessionRepo.Update(ctx, session)
}

func (s *SessionService) ExpireSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.IsCompleted() || session.Status == model.SessionStatusExpired {
		return nil
	}

	if err := session.Expire(); err != nil {
		return err
	}

	if err := s.recalculateSessionScore(ctx, session); err != nil {
		return err
	}

	return s.sessionRepo.Update(ctx, session)
}

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

func (s *SessionService) GetAdminReports(ctx context.Context, userID, bookID *uuid.UUID) ([]repository.AdminSessionReport, error) {
	return s.sessionRepo.GetAdminReports(ctx, userID, bookID)
}

func (s *SessionService) parseQuestionOrder(session *model.Session) ([]uuid.UUID, error) {
	if session.QuestionOrder != "" {
		var order []uuid.UUID
		if err := json.Unmarshal([]byte(session.QuestionOrder), &order); err == nil && len(order) > 0 {
			return order, nil
		}
	}

	questions, err := s.bookRepo.GetQuestionsWithOptions(context.Background(), session.BookID)
	if err != nil {
		return nil, err
	}

	order := make([]uuid.UUID, len(questions))
	for i, q := range questions {
		order[i] = q.ID
	}
	return order, nil
}

func (s *SessionService) recalculateSessionScore(ctx context.Context, session *model.Session) error {
	answers, err := s.answerRepo.FindBySessionID(ctx, session.ID)
	if err != nil {
		return err
	}

	questions, err := s.bookRepo.GetQuestionsWithOptions(ctx, session.BookID)
	if err != nil {
		return err
	}

	answersByQuestion := groupAnswersByQuestion(answers)
	correctCount := 0

	for _, question := range questions {
		selected := make(map[uuid.UUID]model.Option)
		for _, ua := range answersByQuestion[question.ID] {
			for _, opt := range question.Options {
				if opt.ID == ua.OptionID {
					selected[opt.ID] = opt
				}
			}
		}
		if isQuestionFullyCorrect(&question, selected) {
			correctCount++
		}
	}

	session.Score = correctCount
	return nil
}

func isQuestionFullyCorrect(question *model.Question, selected map[uuid.UUID]model.Option) bool {
	correctCount := 0
	for _, opt := range question.Options {
		if opt.IsCorrect {
			correctCount++
		}
	}

	if len(selected) != correctCount {
		return false
	}

	for _, opt := range selected {
		if !opt.IsCorrect {
			return false
		}
	}

	return correctCount > 0 || len(selected) == 0
}

func groupAnswersByQuestion(answers []model.UserAnswer) map[uuid.UUID][]model.UserAnswer {
	result := make(map[uuid.UUID][]model.UserAnswer)
	for _, answer := range answers {
		result[answer.QuestionID] = append(result[answer.QuestionID], answer)
	}
	return result
}

func containsUUID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}
