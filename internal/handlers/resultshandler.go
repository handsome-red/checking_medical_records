package handlers

// import (
// 	"github.com/go-chi/chi/v5"
// 	"github.com/google/uuid"
// 	"med_book/internal/model"
// 	"med_book/internal/service"
// 	"net/http"
// )

// type ResultHandler struct {
// 	resultService  *service.ResultService
// 	sessionService *service.SessionService
// }

// func NewResultHandler(resultService *service.ResultService, sessionService *service.SessionService) *ResultHandler {
// 	return &ResultHandler{
// 		resultService:  resultService,
// 		sessionService: sessionService,
// 	}
// }

// // GET /results - список всех результатов пользователя
// func (h *ResultHandler) ListResults(w http.ResponseWriter, r *http.Request) {
// 	// Получаем пользователя из сессии
// 	userID, err := h.getUserID(r)
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusSeeOther)
// 		return
// 	}

// 	results, err := h.resultService.GetByUserID(userID)
// 	if err != nil {
// 		http.Error(w, "Ошибка получения результатов", http.StatusInternalServerError)
// 		return
// 	}

// 	data := map[string]any{
// 		"results": results,
// 	}

// 	Templates.ExecuteTemplate(w, "results_list.html", data)
// }

// // GET /results/{id} - детальный результат
// func (h *ResultHandler) ShowResult(w http.ResponseWriter, r *http.Request) {
// 	resultIDStr := chi.URLParam(r, "id")
// 	resultID, err := uuid.Parse(resultIDStr)
// 	if err != nil {
// 		http.Error(w, "Неверный ID", http.StatusBadRequest)
// 		return
// 	}

// 	result, err := h.resultService.GetByID(resultID)
// 	if err != nil {
// 		http.Error(w, "Результат не найден", http.StatusNotFound)
// 		return
// 	}

// 	// Получаем детали ответов
// 	details, err := h.resultService.GetResultDetails(result.SessionID)
// 	if err != nil {
// 		http.Error(w, "Ошибка получения деталей", http.StatusInternalServerError)
// 		return
// 	}

// 	data := map[string]any{
// 		"result":  result,
// 		"details": details,
// 	}

// 	Templates.ExecuteTemplate(w, "result_detail.html", data)
// }

// func (h *ResultHandler) getUserID(r *http.Request) (uuid.UUID, error) {
// 	cookie, err := r.Cookie("session_id")
// 	if err != nil {
// 		return uuid.Nil, err
// 	}

// 	sessionID, err := uuid.Parse(cookie.Value)
// 	if err != nil {
// 		return uuid.Nil, err
// 	}

// 	session, err := h.sessionService.GetSessionByID(sessionID)
// 	if err != nil {
// 		return uuid.Nil, err
// 	}

// 	return session.UserID, nil
// }
