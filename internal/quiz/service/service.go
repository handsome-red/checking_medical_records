package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"med_book/internal/quiz/models"
	"med_book/internal/quiz/repository"
)

// Константы
const (
	TestTimeLimit    = 1 * time.Hour
	QuestionsPerTest = 5
)

// Ошибки
var (
	ErrSessionNotFound   = errors.New("session not found")
	ErrTimeLimitExceeded = errors.New("time limit exceeded")
	ErrInvalidQuestion   = errors.New("invalid question")
	ErrSessionCompleted  = errors.New("session already completed")
)

type QuizService struct {
	store     repository.SessionStoreInterface
	questions []models.QuestionBackend
}

func NewQuizService(store repository.SessionStoreInterface, questionsJSON []byte) (*QuizService, error) {
	var data models.QuestionsData
	if err := json.Unmarshal(questionsJSON, &data); err != nil {
		return nil, fmt.Errorf("parse questions: %w", err)
	}
	return &QuizService{
		store:     store,
		questions: data.Questions,
	}, nil
}

// CreateSession создаёт сессию с N случайными вопросами
func (s *QuizService) CreateSession(participant models.Participant) (*models.Session, error) {
	// Выбираем случайные вопросы
	shuffled := make([]models.QuestionBackend, len(s.questions))
	copy(shuffled, s.questions)
	rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	n := QuestionsPerTest
	if len(shuffled) < n {
		n = len(shuffled)
	}

	// Берём только ID вопросов
	questionIDs := make([]int, n)
	for i := 0; i < n; i++ {
		questionIDs[i] = shuffled[i].ID
	}

	// Создаём сессию
	sessionID := fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int63n(10000))
	session := models.NewSession(sessionID, "", participant, questionIDs)

	// Сохраняем
	if err := s.store.Save(session); err != nil {
		return nil, err
	}
	return session, nil
}

// GetSession возвращает сессию
func (s *QuizService) GetSession(id string) (*models.Session, error) {
	return s.store.Get(id)
}

// GetCurrentQuestion возвращает текущий вопрос для фронтенда
func (s *QuizService) GetCurrentQuestion(session *models.Session) (*models.QuestionFrontend, error) {
	if session.IsCompleted() {
		return nil, ErrSessionCompleted
	}

	questionID := session.CurrentQuestionID()
	if questionID == 0 {
		return nil, ErrInvalidQuestion
	}

	// Ищем вопрос в хранилище
	for _, q := range s.questions {
		if q.ID == questionID {
			return q.ToFrontend(), nil
		}
	}
	return nil, ErrInvalidQuestion
}

// SubmitAnswer сохраняет ответ и переходит к следующему вопросу
func (s *QuizService) SubmitAnswer(sessionID string, optionIDs []int) error {
	// Получаем сессию
	session, err := s.store.Get(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}
	if session.IsCompleted() {
		return ErrSessionCompleted
	}
	if session.IsExpired() {
		session.Complete()
		s.store.Save(session)
		return ErrTimeLimitExceeded
	}

	// Сохраняем ответы на текущий вопрос
	currentID := session.CurrentQuestionID()
	for _, optID := range optionIDs {
		session.AddAnswer(currentID, optID)
	}

	// Переходим к следующему вопросу
	isLast := !session.NextQuestion()
	if isLast {
		session.Complete()
	}

	// Сохраняем изменения
	return s.store.Save(session)
}

// GetResults возвращает результаты (правильные ответы и тотал)
func (s *QuizService) GetResults(sessionID string) (correct, total int, answers map[int][]int, err error) {
	session, err := s.store.Get(sessionID)
	if err != nil || session == nil {
		return 0, 0, nil, ErrSessionNotFound
	}
	if !session.IsCompleted() {
		return 0, 0, nil, errors.New("session not completed")
	}

	answers = session.GetAllAnswers()
	total = len(session.QuestionIDs)

	// Считаем правильные ответы
	for _, qID := range session.QuestionIDs {
		userAnswers := session.GetAnswers(qID)

		// Ищем вопрос
		var question *models.QuestionBackend
		for i := range s.questions {
			if s.questions[i].ID == qID {
				question = &s.questions[i]
				break
			}
		}
		if question == nil {
			continue
		}

		// Собираем правильные варианты
		correctOpts := []int{}
		for _, opt := range question.Options {
			if opt.IsCorrect {
				correctOpts = append(correctOpts, opt.ID)
			}
		}

		// Сравниваем
		if len(userAnswers) == len(correctOpts) {
			match := true
			for i := range userAnswers {
				found := false
				for j := range correctOpts {
					if userAnswers[i] == correctOpts[j] {
						found = true
						break
					}
				}
				if !found {
					match = false
					break
				}
			}
			if match {
				correct++
			}
		}
	}
	return correct, total, answers, nil
}

// GetProgress возвращает прогресс прохождения
func (s *QuizService) GetProgress(sessionID string) (current, total int, percentage float64, err error) {
	session, err := s.store.Get(sessionID)
	if err != nil || session == nil {
		return 0, 0, 0, ErrSessionNotFound
	}
	current = session.CurrentIndex
	total = len(session.QuestionIDs)
	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}
	return current, total, percentage, nil
}
