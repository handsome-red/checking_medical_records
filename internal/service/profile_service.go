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
	// books, err := s.bookRepo.GetAllBooks(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get books: %w", err)
	// }

	// 3. Для каждой книги собираем статистику
	bookStats, err := s.sessionRepo.GetUserBooksStats(ctx, userID)

	profileBooks := make([]ProfileBook, 0, len(bookStats))
	for _, book := range bookStats {
		if err != nil {
			return nil, fmt.Errorf("failed to get user`s book stats: %w", err)
		}

		profileBooks = append(profileBooks, ProfileBook{
			BookID:        book.BookID,
			BookName:      book.BookName,
			BestScore:     book.BestScore,
			MaxScore:      book.MaxScore,
			Percent:       book.Percent,
			AttemptsCount: book.AttemptsCount,
		})

	}

	return &Profile{
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Patronymic:   user.Patronymic,
		ProfileBooks: profileBooks,
	}, nil
}
