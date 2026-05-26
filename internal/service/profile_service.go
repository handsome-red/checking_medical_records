// internal/service/profile_service.go
package service

import (
	"context"
	"fmt"
	"med_book/internal/repository"

	"github.com/google/uuid"
)

type ProfileService struct {
	userRepo    repository.UserRepositoryInterface
	sessionRepo repository.SessionRepositoryInterface
	bookRepo    repository.BookRepositoryInterface
	answerRepo  repository.AnswerRepositoryInterface
}

func NewProfileService(
	userRepo repository.UserRepositoryInterface,
	sessionRepo repository.SessionRepositoryInterface,
	bookRepo repository.BookRepositoryInterface,
	answerRepo repository.AnswerRepositoryInterface,
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

	// 2. Получаем статистику по всем книгам ОДНИМ ЗАПРОСОМ
	stats, err := s.sessionRepo.GetUserBooksStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// 3. Преобразуем в ProfileBook
	profileBooks := make([]ProfileBook, 0, len(stats))
	for _, stat := range stats {
		profileBooks = append(profileBooks, ProfileBook{
			BookID:        stat.BookID,
			BookName:      stat.BookName,
			BestScore:     stat.BestScore,
			MaxScore:      stat.MaxScore,
			Percent:       stat.Percent,
			AttemptsCount: stat.AttemptsCount,
		})
	}

	return &Profile{
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Patronymic:   user.Patronymic,
		ProfileBooks: profileBooks,
	}, nil
}
