// internal/handlers/profile_handler.go
package handlers

import (
	"log"
	"net/http"

	"med_book/internal/middleware"
	"med_book/internal/service"
	"med_book/internal/templates"
)

type ProfileHandler struct {
	profileService *service.ProfileService
	sessionService *service.SessionService
	userService    *service.UserService
	templates      *templates.TemplatesManager
}

func NewProfileHandler(
	profileService *service.ProfileService,
	sessionService *service.SessionService,
	userService *service.UserService,
	templates *templates.TemplatesManager,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		sessionService: sessionService,
		userService:    userService,
		templates:      templates,
	}
}

func (h *ProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Получучаем userID
	log.Println("🚀 GetUserProfile START")
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.profileService.GetUserProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "User`s profile not found", http.StatusBadRequest)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "profile.html", profile); err != nil {
		log.Printf("❌ ExecuteTemplate error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
