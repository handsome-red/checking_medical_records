// internal/handlers/profile_handler.go
package handlers

import (
	"net/http"

	"med_book/internal/middleware"
	"med_book/internal/service"
	"med_book/internal/templates"

	"github.com/google/uuid"
)

type ProfileHandler struct {
	userService    *service.UserService
	sessionService *service.SessionService
	bookService    *service.BookService
	template       *templates.TemplatesManager
}

func NewProfileHandler(
	userService *service.UserService,
	sessionService *service.SessionService,
	bookService *service.BookService,
	template *templates.TemplatesManager,
) *ProfileHandler {
	return &ProfileHandler{
		userService:    userService,
		sessionService: sessionService,
		bookService:    bookService,
		template:       template,
	}
}

// GetUserProfile возвращает профиль пользователя
func (h *ProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем пользователя
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Получаем все сессии пользователя
	sessions, err := h.sessionService.GetUserSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
		return
	}

	// Получаем статистику по книгам
	profileBooks := make([]ProfileBook, 0)
	for _, session := range sessions {
		if !session.IsCompleted() {
			continue
		}
		
		book, err := h.bookService.GetBookByID(r.Context(), session.BookID)
		if err != nil {
			continue
		}

		percent := float64(session.Score) / float64(session.MaxScore) * 100
		
		profileBooks = append(profileBooks, ProfileBook{
			BookID:        book.ID,
			BookName:      book.Title,
			BestScore:     session.Score,
			MaxScore:      session.MaxScore,
			Percent:       percent,
			AttemptsCount: 1,
		})
	}

	data := map[string]any{
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"patronymic":    user.Patronymic,
		"profile_books": profileBooks,
	}

	h.template.ExecuteTemplate(w, "profile.html", data)
}

type ProfileBook struct {
	BookID        uuid.UUID `json:"book_id"`
	BookName      string    `json:"book_name"`
	BestScore     int       `json:"best_score"`
	MaxScore      int       `json:"max_score"`
	Percent       float64   `json:"percent"`
	AttemptsCount int       `json:"attempts_count"`
}
