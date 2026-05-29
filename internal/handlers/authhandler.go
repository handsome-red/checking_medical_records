package handlers

import (
	"med_book/internal/middleware"
	"med_book/internal/service"
	"med_book/internal/templates"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
	templates   *templates.TemplatesManager
}

func NewAuthHandler(
	authService *service.AuthService,
	userService *service.UserService,
	templates *templates.TemplatesManager,
) *AuthHandler {
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

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := h.userService.Authenticate(r.Context(), email, password)
	if err != nil {
		data := map[string]interface{}{
			"Error": "Неверный email или пароль",
		}
		h.templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		data := map[string]interface{}{
			"Error": "Ошибка авторизации, попробуйте позже",
		}
		h.templates.ExecuteTemplate(w, "login.html", data)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,                    // Защита от XSS
		Secure:   false,                   // Для разработки false, для прода true
		SameSite: http.SameSiteStrictMode, // Защита от CSRF
		MaxAge:   24 * 60 * 60,            // 24 часа
	})
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *AuthHandler) ShowRegisterPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.GetUserIDFromContext(r.Context()); ok {
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"Error": nil,
	}

	if err := h.templates.ExecuteTemplate(w, "register.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	patronymic := r.FormValue("patronymic")
	password := r.FormValue("password")

	user, err := h.userService.Register(r.Context(), email, firstName, lastName, patronymic, password)
	if err != nil {
		http.Error(w, "error create user: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   24 * 60 * 60,
	})

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
