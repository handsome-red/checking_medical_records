// internal/handlers/register.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"med_book/internal/model"
	"med_book/internal/service"
	"med_book/internal/templates"

	"github.com/google/uuid"
)

type RegistrationHandler struct {
	userService    *service.UserService
	sessionService *service.SessionService
	bookService    *service.BookService
	template       *templates.TemplatesManager
}

func NewRegistrationHandler(
	userService *service.UserService,
	sessionService *service.SessionService,
	bookService *service.BookService,
	template *templates.TemplatesManager,
) *RegistrationHandler {
	return &RegistrationHandler{
		userService:    userService,
		sessionService: sessionService,
		bookService:    bookService,
		template:       template,
	}
}

func (h *RegistrationHandler) ShowRegistrationForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := h.template.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		http.Error(w, "Ошибка загрузки формы регистрации", http.StatusInternalServerError)
	}
}

func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Printf("Метод %s, нужен метод POST\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Printf("Не удалось распарсить форму\n")
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	patronymic := r.FormValue("patronymic")
	password := r.FormValue("password")
	email := r.FormValue("email")

	// Если email не указан, генерируем временный
	if email == "" {
		email = fmt.Sprintf("%s.%s@temp.com", firstName, lastName)
	}

	ctx := r.Context()

	// 1. Находим или создаём пользователя
	user, err := h.findOrCreateUser(ctx, firstName, lastName, patronymic, email, password)
	if err != nil {
		data := map[string]any{
			"Error":      err.Error(),
			"FirstName":  firstName,
			"LastName":   lastName,
			"Patronymic": patronymic,
		}
		h.template.ExecuteTemplate(w, "register.html", data)
		return
	}

	// 2. Проверяем, может ли пользователь начать новый тест сегодня
	canStart, err := h.sessionService.CanStartNewTest(ctx, user.ID)
	if err != nil {
		data := map[string]any{
			"Error":      "Ошибка проверки доступности теста: " + err.Error(),
			"FirstName":  firstName,
			"LastName":   lastName,
			"Patronymic": patronymic,
		}
		h.template.ExecuteTemplate(w, "register.html", data)
		return
	}

	if !canStart {
		nextTime, _ := h.sessionService.GetNextAvailableTime(ctx, user.ID)
		data := map[string]any{
			"Error":      fmt.Sprintf("Вы уже проходили тест сегодня. Следующая попытка доступна %s", nextTime.Format("02.01.2006 15:04")),
			"FirstName":  firstName,
			"LastName":   lastName,
			"Patronymic": patronymic,
		}
		h.template.ExecuteTemplate(w, "register.html", data)
		return
	}

	// 3. Получаем первую книгу для теста
	books, err := h.bookService.GetAllBooks(ctx)
	if err != nil || len(books) == 0 {
		data := map[string]any{
			"Error":      "Книги не найдены",
			"FirstName":  firstName,
			"LastName":   lastName,
			"Patronymic": patronymic,
		}
		h.template.ExecuteTemplate(w, "register.html", data)
		return
	}

	// Берём первую книгу (или можно позволить пользователю выбрать)
	bookID := books[0].ID

	// 4. Создаём сессию
	session, err := h.sessionService.StartTest(ctx, user.ID, bookID, 30*time.Minute)
	if err != nil {
		data := map[string]any{
			"Error":      "Ошибка создания сессии: " + err.Error(),
			"FirstName":  firstName,
			"LastName":   lastName,
			"Patronymic": patronymic,
		}
		h.template.ExecuteTemplate(w, "register.html", data)
		return
	}

	// 5. Устанавливаем cookies
	h.setAuthCookie(w, user.ID)
	h.setSessionCookie(w, session.ID)

	fmt.Printf("✅ Пользователь %s %s (ID: %v) зарегистрирован и начал сессию %v\n",
		user.LastName, user.FirstName, user.ID, session.ID)

	http.Redirect(w, r, "/test", http.StatusSeeOther)
}

// findOrCreateUser находит пользователя по ФИО или создаёт нового
func (h *RegistrationHandler) findOrCreateUser(
	ctx context.Context,
	firstName, lastName, patronymic, email, password string,
) (*model.User, error) {
	// Пытаемся найти пользователя по ФИО
	user, err := h.userService.FindByEmail(ctx, email)
	if err == nil && user != nil {
		return user, nil
	}

	// Если не найден, создаём нового
	user, err = h.userService.Register(ctx, email, firstName, lastName, patronymic, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *RegistrationHandler) setAuthCookie(w http.ResponseWriter, userID uuid.UUID) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userID.String(),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 24 часа
	})
}

func (h *RegistrationHandler) setSessionCookie(w http.ResponseWriter, sessionID uuid.UUID) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   1800, // 30 минут
	})
}

func handleError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case "user with this name already exists":
		http.Error(w, err.Error(), http.StatusConflict)
	case "invalid name", "too short password":
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func respondWithJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
