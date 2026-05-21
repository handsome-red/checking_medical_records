package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"med_book/internal/model"
	"med_book/internal/service"
	"net/http"
)

type RegistrationHandler struct {
	userService     *service.UserService
	sessionService  *service.SessionService
	questionService *service.QuestionService
	testService     *service.TestService
}

func NewRegistrationHandler(
	userService *service.UserService,
	sessionService *service.SessionService,
	questionService *service.QuestionService,
	testService *service.TestService,
) *RegistrationHandler {
	return &RegistrationHandler{
		userService:     userService,
		sessionService:  sessionService,
		questionService: questionService,
		testService:     testService,
	}
}

func (h *RegistrationHandler) ShowRegistrationForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := Templates.ExecuteTemplate(w, "register.html", nil)
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
		fmt.Printf("Не удалось начать парсить форму\n")
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	patronymic := r.FormValue("patronymic")
	password := r.FormValue("password")

	var user *model.User // 👈 объявляем здесь

	person, err := h.userService.FindByFIO(firstName, lastName, patronymic)
	if err != nil {
		// Если ошибка не "пользователь не найден", показываем ошибку
		if !errors.Is(err, errors.New("user not found")) {
			data := map[string]any{
				"Error":      "Failed to find user: " + err.Error(),
				"FirstName":  firstName,
				"LastName":   lastName,
				"Patronymic": patronymic,
			}
			Templates.ExecuteTemplate(w, "register.html", data)
			return
		}
	}

	// Если пользователь найден
	if person != nil {
		// Проверяем, может ли пользователь начать новый тест сегодня
		canStart, err := h.sessionService.CanStartNewTest(person.ID)
		if err != nil {
			data := map[string]any{
				"Error":      "Failed to check test availability: " + err.Error(),
				"FirstName":  firstName,
				"LastName":   lastName,
				"Patronymic": patronymic,
			}
			Templates.ExecuteTemplate(w, "register.html", data)
			return
		}

		if !canStart {
			nextTime, _ := h.sessionService.GetNextAvailableTime(person.ID)
			data := map[string]any{
				"Error":      fmt.Sprintf("Вы уже проходили тест сегодня. Следующая попытка доступна %s", nextTime.Format("02.01.2006 15:04")),
				"FirstName":  firstName,
				"LastName":   lastName,
				"Patronymic": patronymic,
			}
			Templates.ExecuteTemplate(w, "register.html", data)
			return
		}

		user = person // 👈 ВАЖНО: присваиваем существующего пользователя
	} else {
		user, err = h.userService.RegisterUser(firstName, lastName, patronymic, password)
		if err != nil {
			data := map[string]any{
				"Error":      err.Error(),
				"FirstName":  firstName,
				"LastName":   lastName,
				"Patronymic": patronymic,
			}
			Templates.ExecuteTemplate(w, "register.html", data)
			return
		}
		fmt.Printf("User %v: %s %s %s зарегистрирован\n", user.ID, user.LastName, user.FirstName, user.Patronymic)
	}

	// Получаем ID всех книг для теста
	books := h.testService.GetBooks()
	bookIDs := make([]int, len(books))
	for i, book := range books {
		bookIDs[i] = book.ID
	}
	fmt.Printf("Найдено %d книг\n", len(bookIDs))

	session, err := h.sessionService.CreateSession(user.ID, bookIDs)
	if err != nil {
		data := map[string]any{
			"Error":      "Failed to create session: " + err.Error(),
			"FirstName":  firstName,
			"LastName":   lastName,
			"Patronymic": patronymic,
		}
		Templates.ExecuteTemplate(w, "register.html", data)
		return
	}
	fmt.Printf("Создана новая сессия id: %v\n", session.ID)

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
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
