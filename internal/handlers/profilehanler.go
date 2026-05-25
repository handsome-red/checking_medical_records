// internal/handlers/profile_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"med_book/internal/middleware"
	"med_book/internal/model"
	"med_book/internal/service"
	"med_book/internal/templates"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	profileService *service.ProfileService
	sessionService *service.SessionService
	userService    *service.UserService
	testService    *service.TestService
	templates      *templates.TemplatesManager
}

func NewProfileHandler(
	profileService *service.ProfileService,
	sessionService *service.SessionService,
	userService *service.UserService,
	testService *service.TestService,
	templates *templates.TemplatesManager,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		sessionService: sessionService,
		userService:    userService,
		testService:    testService,
		templates:      templates,
	}
}

// GET /profile - страница со списком сессий
func (h *ProfileHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из сессии
	userID, err := h.getUserIDFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем пользователя
	user, err := h.userService.GetUser(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Получаем все сессии пользователя
	sessions, err := h.sessionService.GetUserSessions(userID)
	if err != nil {
		http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
		return
	}

	// Формируем данные для шаблона
	var sessionInfos []map[string]interface{}
	totalCorrect := 0
	totalQuestions := 0
	bestPercentage := 0.0

	for i, session := range sessions {
		// Получаем прогресс
		progress, err := h.sessionService.GetProgress(session.ID)
		if err != nil {
			continue
		}

		// Пропускаем незавершённые сессии для статистики
		if progress.CompletedAt == nil {
			continue
		}

		// Получаем книгу
		book, err := h.testService.GetBook(progress.CurrentBookID)
		if err != nil {
			continue
		}

		// Получаем вопросы для подсчёта правильных ответов
		questions, err := h.sessionService.GetSessionQuestions(session.ID)
		if err != nil {
			continue
		}

		// Получаем ответы
		answers, err := h.sessionService.GetAnswersBySession(session.ID)
		if err != nil {
			continue
		}

		answersMap := make(map[int][]int)
		for _, a := range answers {
			answersMap[a.QuestionID] = append(answersMap[a.QuestionID], a.AnswerID)
		}

		// Считаем правильные ответы
		correctCount := 0
		for _, q := range questions {
			userAnswers := answersMap[q.QuestionID]
			if h.isAnswerCorrect(userAnswers, q.QuestionID) {
				correctCount++
			}
		}

		total := len(questions)
		percentage := float64(correctCount) / float64(total) * 100

		totalCorrect += correctCount
		totalQuestions += total

		if percentage > bestPercentage {
			bestPercentage = percentage
		}

		// Вычисляем время прохождения
		duration := session.ExpiresAt.Sub(session.CreatedAt)

		sessionInfo := map[string]interface{}{
			"Number":     i + 1,
			"SessionID":  session.ID.String(),
			"BookName":   book.Name,
			"Correct":    correctCount,
			"Total":      total,
			"Percentage": percentage,
			"Date":       session.CreatedAt.Format("02.01.2006"),
			"Duration":   formatDurationShort(duration),
			"Completed":  progress.CompletedAt != nil,
		}

		sessionInfos = append(sessionInfos, sessionInfo)
	}

	avgScore := 0.0
	if totalQuestions > 0 {
		avgScore = float64(totalCorrect) / float64(totalQuestions) * 100
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"FullName": fmt.Sprintf("%s %s %s", user.LastName, user.FirstName, user.Patronymic),
		},
		"Sessions":      sessionInfos,
		"TotalSessions": len(sessionInfos),
		"AvgScore":      int(avgScore),
		"BestScore":     int(bestPercentage),
	}

	h.templates.ExecuteTemplate(w, "profile.html", data)
}

func formatDurationShort(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%d мин %d сек", minutes, seconds)
	}
	return fmt.Sprintf("%d сек", seconds)
}

// GET /profile/session/{id} - получить детальную информацию о сессии
func (h *ProfileHandler) GetSessionDetails(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Получаем сессию
	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Получаем пользователя
	user, err := h.userService.GetUser(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Получаем прогресс
	progress, err := h.sessionService.GetProgress(sessionID)
	if err != nil {
		http.Error(w, "Progress not found", http.StatusNotFound)
		return
	}

	// Получаем книгу
	book, err := h.testService.GetBook(progress.CurrentBookID)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Получаем вопросы
	questions, err := h.sessionService.GetSessionQuestions(sessionID)
	if err != nil {
		http.Error(w, "Questions not found", http.StatusNotFound)
		return
	}

	// Получаем ответы
	answers, err := h.sessionService.GetAnswersBySession(sessionID)
	if err != nil {
		http.Error(w, "Answers not found", http.StatusNotFound)
		return
	}

	answersMap := make(map[int][]int)
	for _, a := range answers {
		answersMap[a.QuestionID] = append(answersMap[a.QuestionID], a.AnswerID)
	}

	// Подсчёт правильных ответов
	correctCount := 0
	questionDetails := []map[string]interface{}{}

	for _, q := range questions {
		userAnswers := answersMap[q.QuestionID]
		isCorrect := h.isAnswerCorrect(userAnswers, q.QuestionID)

		if isCorrect {
			correctCount++
		}

		questionDetails = append(questionDetails, map[string]interface{}{
			"question_id":  q.QuestionID,
			"user_answers": userAnswers,
			"is_correct":   isCorrect,
		})
	}

	duration := session.ExpiresAt.Sub(session.CreatedAt)

	response := map[string]interface{}{
		"session_id":     session.ID.String(),
		"session_number": h.getSessionNumberForUser(session.UserID, session.ID),
		"user": map[string]interface{}{
			"id":        user.ID.String(),
			"full_name": fmt.Sprintf("%s %s %s", user.LastName, user.FirstName, user.Patronymic),
		},
		"book": map[string]interface{}{
			"id":   book.ID,
			"name": book.Name,
		},
		"statistics": map[string]interface{}{
			"correct":    correctCount,
			"total":      len(questions),
			"percentage": float64(correctCount) / float64(len(questions)) * 100,
		},
		"timing": map[string]interface{}{
			"start_time": session.CreatedAt.Format("02.01.2006 15:04:05"),
			"end_time": func() string {
				if progress.CompletedAt != nil {
					return progress.CompletedAt.Format("02.01.2006 15:04:05")
				}
				return "Не завершён"
			}(),
			"duration": formatDuration(duration),
		},
		"questions": questionDetails,
		"status": func() string {
			if progress.CompletedAt != nil {
				return "completed"
			}
			return "in_progress"
		}(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Вспомогательные методы
func (h *ProfileHandler) getUserIDFromSession(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return uuid.Nil, err
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}

	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		return uuid.Nil, err
	}

	return session.UserID, nil
}

func (h *ProfileHandler) getSessionNumber(sessions []*model.Session, sessionID uuid.UUID) int {
	for i, s := range sessions {
		if s.ID == sessionID {
			return i + 1
		}
	}
	return 0
}

func (h *ProfileHandler) getSessionNumberForUser(userID uuid.UUID, sessionID uuid.UUID) int {
	sessions, _ := h.sessionService.GetUserSessions(userID)
	for i, s := range sessions {
		if s.ID == sessionID {
			return i + 1
		}
	}
	return 0
}

func (h *ProfileHandler) isAnswerCorrect(userAnswers []int, questionID int) bool {
	question, err := h.testService.GetQuestion(questionID)
	if err != nil {
		return false
	}

	correctMap := make(map[int]bool)
	for i, opt := range question.Options {
		if opt.Correct {
			correctMap[i] = true
		}
	}

	if len(userAnswers) != len(correctMap) {
		return false
	}

	for _, ans := range userAnswers {
		if !correctMap[ans] {
			return false
		}
	}

	return true
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dч %dм %dс", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dм %dс", minutes, seconds)
	}
	return fmt.Sprintf("%dс", seconds)
}

type ProfileResponse struct {
	ID         uuid.UUID
	FirstName  string
	LastName   string
	Patronymic string
	Books      []UserBookStat
}

type UserBookStat struct {
	BookID          int
	BookName        string
	AttemptsCount   int      // Количество попыток
	BestScore       int      // Лучшее количество правильных ответов
	TotalQuestions  int      // Всего вопросов в книге
	BestPercentage  float64  // Лучший процент (для отображения)
}

func (h *ProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Получучаем userID
	log.Println("🚀 GetUserProfile START")
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем ФИО пользователя
	user, err := h.userService.GetUser(userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	books := h.testService.GetBooks()
	booksStats := make([]UserBookStat, 0, len(books))

	for _, book := range books {
		booksStats = append(booksStats, UserBookStat{
			BookID: book.ID,
			BookName: book.Name,
			AttemptsCount: h.sessionService.CompleteSession()
			BestScore      int     // Лучшее количество правильных ответов
			TotalQuestions int     // Всего вопросов в книге
			BestPercentage
		})
	}

	response := ProfileResponse{
		ID:         user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Patronymic: user.Patronymic,
		Books:     	
	}

	if err := h.templates.ExecuteTemplate(w, "profile.html", response); err != nil {
		log.Printf("❌ ExecuteTemplate error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
