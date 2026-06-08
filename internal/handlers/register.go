// internal/handlers/register.go
package handlers

import (
	"context"
	"fmt"
	"net/http"

	"med_book/internal/model"
	"med_book/internal/service"
	"med_book/internal/templates"
)

type RegistrationHandler struct {
	userService *service.UserService
	authService *service.AuthService
	adminEmail  string
	template    *templates.TemplatesManager
}

func NewRegistrationHandler(
	userService *service.UserService,
	authService *service.AuthService,
	adminEmail string,
	template *templates.TemplatesManager,
) *RegistrationHandler {
	return &RegistrationHandler{
		userService: userService,
		authService: authService,
		adminEmail:  adminEmail,
		template:    template,
	}
}

func (h *RegistrationHandler) ShowRegistrationForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.ExecuteTemplate(w, "register.html", nil); err != nil {
		http.Error(w, "Ошибка загрузки формы регистрации", http.StatusInternalServerError)
	}
}

func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	patronymic := r.FormValue("patronymic")
	password := r.FormValue("password")
	email := r.FormValue("email")

	ctx := r.Context()

	user, err := h.findOrCreateUser(ctx, firstName, lastName, patronymic, email, password)
	if err != nil {
		data := map[string]any{
			"Error":      err.Error(),
			"Email":      email,
			"FirstName":  firstName,
			"LastName":   lastName,
			"Patronymic": patronymic,
		}
		_ = h.template.ExecuteTemplate(w, "register.html", data)
		return
	}

	if h.adminEmail != "" && email == h.adminEmail {
		_ = h.userService.PromoteToAdmin(ctx, email)
	}

	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Ошибка авторизации", http.StatusInternalServerError)
		return
	}

	setAuthTokenCookie(w, token)
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *RegistrationHandler) findOrCreateUser(
	ctx context.Context,
	firstName, lastName, patronymic, email, password string,
) (*model.User, error) {
	existing, err := h.userService.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("пользователь с таким email уже существует")
	}

	user, err := h.userService.Register(ctx, email, firstName, lastName, patronymic, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func setAuthTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   24 * 60 * 60,
	})
}
