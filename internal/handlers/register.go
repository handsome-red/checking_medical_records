package handlers

import (
	"encoding/json"
	"med_book/internal/handlers/dto"
	"med_book/internal/service"
	"net/http"
)

type RegistrationHandler struct {
	userService *service.UserService
}

func NewRegistrationHandler(userService *service.UserService) *RegistrationHandler {
	return &RegistrationHandler{
		userService: userService,
	}
}

func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	user, err := h.userService.RegisterUser(
		req.FirstName,
		req.LastName,
		req.Patronomyc,
		req.Password,
	)

	if err != nil {
		handleError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, dto.ToUserResponse(user))
}

func (h *RegistrationHandler) ShowRegistrationForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := Templates.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		http.Error(w, "Ошибка загрузки формы регистрации", http.StatusInternalServerError)
	}
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

func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
