package handlers

import (
	"html/template"
	"med_book/internal/middleware"
	"med_book/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
	templates   *template.Template
}

func NewAuthHandler(
	authService *service.AuthService,
	userService *service.UserService,
) *AuthHandler {
	templates := template.Must(template.ParseGlob("internal/templates/*.html"))

	return &AuthHandler{
		authService: authService,
		userService: userService,
		templates:   templates,
	}
}

func (h *AuthHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.GetUserIDFromContext(r.Context()); ok {
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"Error": nil,
	}

	if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Login()            {}
func (h *AuthHandler) ShowRegisterPage() {}
func (h *AuthHandler) Register()         {}
