// internal/handler/test_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"med_book/internal/middleware"
	"med_book/internal/model"
	"med_book/internal/service"

	"github.com/google/uuid"
)

type TestHandler struct {
	TestService    *service.TestService
	sessionService *service.SessionService
}

func NewTestHandler(
	TestService *service.TestService,
	sessionService *service.SessionService,
) *TestHandler {
	return &TestHandler{
		TestService:    TestService,
		sessionService: sessionService,
	}
}

// GET /test - показать текущий вопрос
func (h *TestHandler) ShowTest(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		fmt.Printf("getSession ошибка: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Получаем прогресс
	progress, err := h.sessionService.GetProgress(session.ID)
	if err != nil {
		fmt.Printf("GetProgress ошибка: %v\n", err)
		http.Error(w, "ошибка получения прогресса", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Progress: CurrentBookID=%d, CurrentQIndex=%d, CurrentQuestionID=%d\n",
		progress.CurrentBookID, progress.CurrentQIndex, progress.CurrentQuestionID)

	// Проверяем, истекла ли сессия
	if progress.CompletedAt != nil {
		fmt.Printf("Сессия уже завершена")
		http.Redirect(w, r, "/test/result", http.StatusSeeOther)
		return
	}

	// Проверяем, закончен ли тест
	if session.ExpiresAt.Before(time.Now()) {
		fmt.Printf("Сессия истекла по вермени")
		if err := h.sessionService.CompleteSession(session.ID); err != nil {
			fmt.Printf("Ошибка завершения сессии: %v\n", err)
		}
		http.Redirect(w, r, "/test/result", http.StatusSeeOther)
		return
	}
	fmt.Printf("До конца сессии id: %d осталось %v\n", session.ID, session.ExpiresAt)

	// Получаем вопросы сессии
	questionsID, err := h.sessionService.GetSessionQuestions(session.ID)
	if err != nil {
		fmt.Printf("GetSessionQuestions ошибка: %v\n", err)
		http.Error(w, "Ошибка получения вопросов", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Всего в сессии %d вопросов\n", len(questionsID))

	// Получаем текущий вопрос
	currentQuestionID, err := h.sessionService.GetCurrentQuestionID(session.ID)
	if err != nil {
		fmt.Printf("GetCurrentQuestion ошибка: %v\n", err)
		http.Error(w, "Ошибка получения текущего вопроса", http.StatusInternalServerError)
		return
	}
	fmt.Printf("CurrentQuestionID: %d\n", currentQuestionID)

	currentQuestion, err := h.TestService.GetQuestion(currentQuestionID)
	if err != nil {
		fmt.Printf("Ошибка получения вопроса %v: %v\n", currentQuestionID, err)
		http.Error(w, "Вопрос не найден", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Question text: %s\n", currentQuestion.Text)

	pages, err := h.TestService.GetPagesByBookID(progress.CurrentBookID)
	if err != nil {
		fmt.Printf("Ошибка получения страниц книги с id %d\n", progress.CurrentBookID)
		http.Error(w, "Страницы книги не найдены", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Pages count: %d\n", len(pages))

	data := map[string]any{
		"question":    currentQuestion,
		"pages":       pages,
		"current":     progress.CurrentQIndex + 1,
		"total":       len(questionsID),
		"progress":    h.getProgress(progress.CurrentQIndex, len(questionsID)),
		"expiresAt":   session.ExpiresAt.Unix(),
		"currentBook": currentQuestion.BookID,
	}

	Templates.ExecuteTemplate(w, "test.html", data)
}

// POST /test/submit - обработать ответ
func (h *TestHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Получаем вопросы сессии
	questionsID, err := h.sessionService.GetSessionQuestions(session.ID)
	if err != nil {
		http.Error(w, "Ошибка получения вопросов", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	questionID, _ := strconv.Atoi(r.FormValue("question_id"))
	answers := parseAnswers(r.Form["answers"])

	// Сохраняем каждый ответ
	for _, answerID := range answers {
		if err := h.sessionService.SaveAnswer(session.ID, questionID, answerID); err != nil {
			fmt.Printf("Ошибка сохранения ответа: %v\n", err)
		}
	}

	// Получаем прогресс сессии
	progress, err := h.sessionService.GetProgress(session.ID)
	if err != nil {
		http.Error(w, "ошибка получения прогресса", http.StatusInternalServerError)
		return
	}

	// Проверяем, последний ли вопрос
	if progress.CurrentQIndex >= len(questionsID)-1 {
		fmt.Println("!!! ВЫЗЫВАЕМ completeTest !!!")
		if err := h.completeTest(session.ID); err != nil {
			fmt.Printf("Ошибка завершения теста: %v\n", err)
		} else {
			fmt.Println("!!! ТЕСТ УСПЕШНО ЗАВЕРШЁН !!!")
		}
		http.Redirect(w, r, "/test/result", http.StatusSeeOther)
		return
	}

	// Переходим к следующему вопросу
	if err := h.sessionService.MoveToNextQuestion(session.ID); err != nil {
		http.Error(w, "Ошибка перехода к следующему вопросу", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/test", http.StatusSeeOther)
}

// GET /test/result - показать результаты
func (h *TestHandler) ShowResult(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	results := h.calculateResults(session.ID)

	data := map[string]any{
		"results":    results["results"],
		"correct":    results["correct"],
		"total":      results["total"],
		"percentage": results["percentage"],
	}

	Templates.ExecuteTemplate(w, "result.html", data)
}

// POST /test/start - начать новый тест
func (h *TestHandler) StartTest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Ошибка получения userID", http.StatusInternalServerError)
		return
	}
	fmt.Printf("StartTest для пользователя: %v\n", userID)

	hasUnfinished, existingSession, err := h.sessionService.HasUnfinishedSession(userID)
	if err != nil {
		http.Error(w, "Ошибка проверки", http.StatusInternalServerError)
		return
	}

	if hasUnfinished && existingSession != nil {
		// Продолжаем существующую сессию
		http.SetCookie(w, &http.Cookie{
			Name:  "session_id",
			Value: existingSession.ID.String(),
			Path:  "/",
		})
		http.Redirect(w, r, "/test", http.StatusSeeOther)
		return
	}

	// 2. Проверяем, можно ли начать новый тест сегодня
	canStart, err := h.sessionService.CanStartNewTest(userID)
	fmt.Printf("CanStartNewTest: %v, ошибка: %v\n", canStart, err)
	if err != nil {
		http.Error(w, "Ошибка проверки", http.StatusInternalServerError)
		return
	}

	if !canStart {
		nextTime, _ := h.sessionService.GetNextAvailableTime(userID)
		data := map[string]interface{}{
			"Message":  "Вы уже проходили тест сегодня",
			"NextTime": nextTime.Format("02.01.2006 15:04"),
		}
		Templates.ExecuteTemplate(w, "cooldown.html", data)
		return
	}

	// Получаем все книги теста
	books := h.TestService.GetBooks()
	if len(books) == 0 {
		http.Error(w, "Книги не найдены", http.StatusInternalServerError)
		return
	}

	// Извлекаем ID всех вопросов
	booksID := make([]int, len(books))
	for i, q := range books {
		booksID[i] = q.ID
	}

	// Создаем сессию
	session, err := h.sessionService.CreateSession(
		userID,
		booksID,
	)

	if err != nil {
		http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie с ID сессии
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   1800, // 30 минут
	})

	http.Redirect(w, r, "/test", http.StatusSeeOther)
}

// calculateResults - вычисляет результаты сессии
func (h *TestHandler) calculateResults(sessionID uuid.UUID) map[string]any {
	type QuestionResult struct {
		QuestionText   string
		Image          string
		UserAnswers    []string
		CorrectAnswers []string
		IsCorrect      bool
	}

	// Получаем все вопросы сессии
	sessionQuestions, err := h.sessionService.GetSessionQuestions(sessionID)
	if err != nil {
		fmt.Printf("Ошибка получения вопросов сессии: %v\n", err)
		return map[string]any{"error": "Failed to get session questions"}
	}

	// Получаем все ответы сессии
	answers, err := h.sessionService.GetAnswersBySession(sessionID)
	if err != nil {
		fmt.Printf("Ошибка получения ответов сессии: %v\n", err)
		return map[string]any{"error": "Failed to get session answers"}
	}

	answersMap := make(map[int][]int)
	for _, answer := range answers {
		answersMap[answer.QuestionID] = append(answersMap[answer.QuestionID], answer.AnswerID)
	}

	results := make([]QuestionResult, 0, len(sessionQuestions))
	correctCount := 0

	for _, sq := range sessionQuestions {
		// Получаем вопрос
		question, err := h.TestService.GetQuestion(sq.QuestionID)
		if err != nil {
			fmt.Printf("Ошибка получения вопроса %d: %v\n", question.ID, err)
			continue
		}

		// Получаем ответы пользователя для этого вопроса
		userAnswerIndexes := answersMap[sq.QuestionID]

		// Преобразуем индексы в тексты ответов пользователя
		userAnswersText := make([]string, 0)
		for _, idx := range userAnswerIndexes {
			if idx >= 0 && idx < len(question.Options) {
				userAnswersText = append(userAnswersText, question.Options[idx].Text)
			}
		}

		// Находим правильные ответы (тексты)
		correctAnswersText := make([]string, 0)
		for _, opt := range question.Options {
			if opt.Correct {
				correctAnswersText = append(correctAnswersText, opt.Text)
			}
		}

		// Проверяем, правильный ли ответ
		isCorrect := compareAnswers(userAnswerIndexes, getCorrectIndexes(question.Options))
		if isCorrect {
			correctCount++
		}

		results = append(results, QuestionResult{
			QuestionText: question.Text,
			// Image:          question.Image,
			UserAnswers:    userAnswersText,
			CorrectAnswers: correctAnswersText,
			IsCorrect:      isCorrect,
		})
	}

	total := len(sessionQuestions)
	percentage := 0.0
	if total > 0 {
		percentage = float64(correctCount) / float64(total) * 100
	}

	return map[string]any{
		"results":    results,
		"correct":    correctCount,
		"total":      total,
		"percentage": percentage,
	}
}

// Вспомогательные функции
func (h *TestHandler) getSession(r *http.Request) (*model.Session, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, fmt.Errorf("no session cookie")
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID")
	}

	session, err := h.sessionService.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Получена сессия id: %d из cookie\n", sessionID)

	return session, nil
}

func (h *TestHandler) getProgress(currentIndex, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(currentIndex) / float64(total) * 100
}

// parseAnswers преобразует []string из формы в []int
func parseAnswers(answers []string) []int {
	result := make([]int, 0, len(answers))
	for _, val := range answers {
		idx, err := strconv.Atoi(val)
		if err != nil {
			fmt.Printf("Ошибка парсинга ответа %s: %v\n", val, err)
			continue
		}
		result = append(result, idx)
	}
	return result
}

// getCorrectIndexes получает индексы правильных ответов
func getCorrectIndexes(options []model.Option) []int {
	correct := make([]int, 0)
	for i, opt := range options {
		if opt.Correct {
			correct = append(correct, i)
		}
	}
	return correct
}

// compareAnswers сравнивает ответы пользователя с правильными
func compareAnswers(user, correct []int) bool {
	if len(user) != len(correct) {
		return false
	}

	correctMap := make(map[int]bool)
	for _, c := range correct {
		correctMap[c] = true
	}

	for _, u := range user {
		if !correctMap[u] {
			return false
		}
	}

	return true
}

// completeTest завершает тест (устанавливает CompletedAt)
func (h *TestHandler) completeTest(sessionID uuid.UUID) error {
	return h.sessionService.CompleteSession(sessionID)
}
