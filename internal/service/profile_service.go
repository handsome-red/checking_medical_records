// internal/service/profile_service.go
package service

import (
	"context"
	"fmt"

	"med_book/internal/repository"

	"github.com/google/uuid"
)

type ProfileService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	bookRepo    *repository.BookRepository
	answerRepo  *repository.AnswerRepository
}

func NewProfileService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	bookRepo *repository.BookRepository,
	answerRepo *repository.AnswerRepository,
) *ProfileService {
	return &ProfileService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		bookRepo:    bookRepo,
		answerRepo:  answerRepo,
	}
}

type Profile struct {
	FirstName    string        `json:"first_name"`
	LastName     string        `json:"last_name"`
	Patronymic   string        `json:"patronymic"`
	ProfileBooks []ProfileBook `json:"profile_books"`
}

type ProfileBook struct {
	BookID        uuid.UUID `json:"book_id"`
	BookName      string    `json:"book_name"`
	BestScore     int       `json:"best_score"`
	MaxScore      int       `json:"max_score"`
	Percent       float64   `json:"percent"`
	AttemptsCount int       `json:"attempts_count"`
}

// GetUserProfile возвращает профиль пользователя с группировкой по книгам
func (s *ProfileService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	// 1. Получаем пользователя
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Получаем все книги
	books, err := s.bookRepo.GetAllBooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
	}

	// 3. Для каждой книги собираем статистику
	profileBooks := make([]ProfileBook, 0, len(books))
	for _, book := range books {
		// Получаем лучший результат пользователя по этой книге
		bestScore, err := s.sessionRepo.GetUserBestScoreForBook(ctx, userID, book.ID)
		if err != nil {
			// Если ошибка, просто пропускаем книгу или ставим 0
			bestScore = 0
		}

		// Получаем максимальный балл за книгу
		maxScore, err := s.bookRepo.GetMaxScore(ctx, book.ID)
		if err != nil {
			maxScore = 0
		}

		// Получаем количество попыток
		attemptsCount, err := s.sessionRepo.GetUserAttempts(ctx, userID, book.ID)
		if err != nil {
			attemptsCount = 0
		}

		// Вычисляем процент
		percent := 0.0
		if maxScore > 0 {
			percent = float64(bestScore) / float64(maxScore) * 100
		}

		// Добавляем только книги, которые пользователь хоть раз проходил
		if attemptsCount > 0 || bestScore > 0 {
			profileBooks = append(profileBooks, ProfileBook{
				BookID:        book.ID,
				BookName:      book.Title,
				BestScore:     bestScore,
				MaxScore:      maxScore,
				Percent:       percent,
				AttemptsCount: attemptsCount,
			})
		}
	}

	return &Profile{
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Patronymic:   user.Patronymic,
		ProfileBooks: profileBooks,
	}, nil
}
