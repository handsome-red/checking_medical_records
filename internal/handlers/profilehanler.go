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

func (h *ProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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
		"IsAdmin":      user.IsAdmin(),
	}

	_ = h.template.ExecuteTemplate(w, "profile.html", data)
}

type ProfileBook struct {
	BookID        uuid.UUID `json:"book_id"`
	BookName      string    `json:"book_name"`
	BestScore     int       `json:"best_score"`
	MaxScore      int       `json:"max_score"`
	Percent       float64   `json:"percent"`
	AttemptsCount int       `json:"attempts_count"`
}
