// internal/service/profile_service.go
package service

import (
	"fmt"
	"med_book/internal/repository"
	"time"

	"github.com/google/uuid"
)

type ProfileService struct {
	sessionRepo    repository.SessionRepositoryInterface
	sessionService *SessionService
}

func NewProfileService(
	sessionRepo repository.SessionRepositoryInterface,
	sessionService *SessionService,
) *ProfileService {
	return &ProfileService{
		sessionRepo:    sessionRepo,
		sessionService: sessionService,
	}
}

type SessionInfo struct {
	SessionID      string
	UserFullName   string
	BookID         int
	BookNumber     int
	CorrectAnswers int
	TotalQuestions int
	StartTime      string
	EndTime        string
	Duration       string
	Completed      bool
}

func (s *ProfileService) GetUserSessions(userID uuid.UUID) ([]SessionInfo, error) {
	sessions, err := s.sessionRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	var result []SessionInfo

	for _, session := range sessions {
		// Получаем прогресс сессии
		progress, err := s.sessionService.GetProgress(session.ID)
		if err != nil {
			continue
		}

		// Получаем вопросы сессии
		questions, err := s.sessionRepo.FindQuestionsBySessionID(session.ID)
		if err != nil {
			continue
		}

		// Получаем ответы
		answers, err := s.sessionRepo.FindAnswersBySessionID(session.ID)
		if err != nil {
			continue
		}

		// Считаем правильные ответы для текущей книги
		answersMap := make(map[int][]int)
		for _, a := range answers {
			answersMap[a.QuestionID] = append(answersMap[a.QuestionID], a.AnswerID)
		}

		correct := 0
		for _, q := range questions {
			// Пропускаем вопросы не из текущей книги
			if q.BookID != progress.CurrentBookID {
				continue
			}

			userAnswers := answersMap[q.QuestionID]
			if s.isAnswerCorrect(userAnswers, q.QuestionID) {
				correct++
			}
		}

		// Получаем книгу
		BookNumber := progress.CurrentBookIndex
		// Здесь нужно получить имя книги из BookService

		// Вычисляем длительность
		duration := session.ExpiresAt.Sub(session.CreatedAt)

		endTime := ""
		if progress.CompletedAt != nil {
			endTime = progress.CompletedAt.Format("02.01.2006 15:04:05")
		} else {
			endTime = "Не завершён"
		}

		info := SessionInfo{
			SessionID:      session.ID.String(),
			UserFullName:   "", // нужно из userRepo
			BookID:         progress.CurrentBookID,
			BookNumber:     BookNumber,
			CorrectAnswers: correct,
			TotalQuestions: len(questions),
			StartTime:      session.CreatedAt.Format("02.01.2006 15:04:05"),
			EndTime:        endTime,
			Duration:       formatDuration(duration),
			Completed:      progress.CompletedAt != nil,
		}

		result = append(result, info)
	}

	return result, nil
}

func (s *ProfileService) isAnswerCorrect(userAnswers []int, questionID int) bool {
	question, _ := s.sessionService.testService.GetQuestion(questionID)

	correctAnswers := make(map[int]bool, 0)
	for i, answer := range question.Options {
		if answer.Correct {
			correctAnswers[i] = true
		}
	}

	if len(correctAnswers) != len(userAnswers) {
		return false
	}

	for _, answer := range userAnswers {
		if !correctAnswers[answer] {
			return false
		}
	}

	return true
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dч %dм %dс", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dм %dс", minutes, seconds)
	}
	return fmt.Sprintf("%dс", seconds)
}
