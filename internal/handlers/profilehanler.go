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

	stats, err := h.sessionService.GetUserBooksStats(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	// Получаем статистику по книгам
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

	data := map[string]any{
		"FirstName":    user.FirstName,
		"LastName":     user.LastName,
		"Patronymic":   user.Patronymic,
		"ProfileBooks": profileBooks,
	}

	h.template.ExecuteTemplate(w, "profile.html", data)
}

// GetUserProfile возвращает профиль пользователя
func (h *ProfileHandler) Statistics(w http.ResponseWriter, r *http.Request) {
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

	if !user.IsAdmin() {
		http.Error(w, "User is not admin", http.StatusNotFound)
		return
	}

	statistics := h.sessionService.GetGlobalUsersStat()

	data := map[string]any{
		"statistics": statistics,
	}

	h.template.ExecuteTemplate(w, "statistics.html", data)
}

type ProfileBook struct {
	BookID        uuid.UUID `json:"book_id"`
	BookName      string    `json:"book_name"`
	BestScore     int       `json:"best_score"`
	MaxScore      int       `json:"max_score"`
	Percent       float64   `json:"percent"`
	AttemptsCount int       `json:"attempts_count"`
}
