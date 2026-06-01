// internal/handlers/testhandler.go
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"med_book/internal/middleware"
	"med_book/internal/model"
	"med_book/internal/service"
	"med_book/internal/templates"

	"github.com/google/uuid"
)

type TestHandler struct {
	sessionService *service.SessionService
	bookService    *service.BookService
	template       *templates.TemplatesManager
}

func NewTestHandler(
	sessionService *service.SessionService,
	bookService *service.BookService,
	template *templates.TemplatesManager,
) *TestHandler {
	return &TestHandler{
		sessionService: sessionService,
		bookService:    bookService,
		template:       template,
	}
}

func (h *TestHandler) ShowStartPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.GetUserIDFromContext(r.Context()); ok {
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"Error": nil,
	}

	if err := h.template.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GET /test - показать текущий вопрос
func (h *TestHandler) ShowTest(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		fmt.Printf("getSession ошибка: %v\n", err)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	// Проверяем, не завершена ли сессия
	if session.IsCompleted() {
		fmt.Println("Сессия уже завершена")
		http.Redirect(w, r, "/test/result", http.StatusSeeOther)
		return
	}

	// Проверяем, не истекла ли сессия
	if session.IsExpired() {
		fmt.Println("Сессия истекла по времени")
		if err := h.sessionService.CompleteSession(r.Context(), session.ID); err != nil {
			fmt.Printf("Ошибка завершения сессии: %v\n", err)
		}
		http.Redirect(w, r, "/test/result", http.StatusSeeOther)
		return
	}

	// Получаем текущий вопрос
	currentQuestion, err := h.sessionService.GetCurrentQuestion(r.Context(), session.ID)
	if err != nil {
		if err.Error() == "all questions answered" {
			h.completeTest(r.Context(), session.ID)
			http.Redirect(w, r, "/test/result", http.StatusSeeOther)
			return
		}
		fmt.Printf("GetCurrentQuestion ошибка: %v\n", err)
		http.Error(w, "Ошибка получения вопроса", http.StatusInternalServerError)
		return
	}

	// Получаем книгу для отображения страниц
	book, err := h.bookService.GetBookByID(r.Context(), session.BookID)
	if err != nil {
		fmt.Printf("GetBookByID ошибка: %v\n", err)
		http.Error(w, "Книга не найдена", http.StatusInternalServerError)
		return
	}

	// Получаем прогресс (сколько ответил)
	answers, err := h.sessionService.GetAnswersBySession(r.Context(), session.ID)
	if err != nil {
		fmt.Printf("GetAnswersBySession ошибка: %v\n", err)
		answers = []model.UserAnswer{}
	}

	totalQuestions, err := h.bookService.CountQuestions(r.Context(), session.BookID)
	if err != nil {
		fmt.Printf("CountQuestions ошибка: %v\n", err)
		totalQuestions = 0
	}

	// Получаем страницы книги
	pages, err := h.bookService.GetPagesByBookID(r.Context(), session.BookID)
	if err != nil {
		fmt.Printf("GetPagesByBookID ошибка: %v\n", err)
		pages = []model.BookPage{}
	}

	fmt.Println("TIME: ")
	fmt.Println(session.GetTimeRemaining().Seconds())

	var expiresAtUnix int64
	if session.ExpiresAt != nil {
		expiresAtUnix = session.ExpiresAt.Unix()
	} else {
		expiresAtUnix = 0
	}

	data := map[string]any{
		"question":       currentQuestion,
		"pages":          pages,
		"current":        len(answers) + 1,
		"total":          totalQuestions,
		"progress":       h.getProgress(len(answers), totalQuestions),
		"expiresAt":      expiresAtUnix,
		"book_title":     book.Title,
		"time_remaining": session.GetTimeRemainingFormatted(),
		"question_index": len(answers) + 1,
	}

	if err := h.template.ExecuteTemplate(w, "test.html", data); err != nil {
		fmt.Printf("ExecuteTemplate ошибка: %v\n", err)
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
	}
}

// POST /test/submit - обработать ответ
// internal/handlers/testhandler.go

func (h *TestHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		fmt.Printf("❌ getSession error: %v\n", err)
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}
	fmt.Printf("✅ Session: %s\n", session.ID)

	if err := r.ParseForm(); err != nil {
		fmt.Printf("❌ ParseForm error: %v\n", err)
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	questionIDStr := r.FormValue("question_id")
	fmt.Printf("📝 question_id: %s\n", questionIDStr)

	if questionIDStr == "" {
		fmt.Println("❌ question_id is empty")
		http.Error(w, "ID вопроса не указан", http.StatusBadRequest)
		return
	}

	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		fmt.Printf("❌ uuid.Parse error: %v\n", err)
		http.Error(w, "Неверный ID вопроса", http.StatusBadRequest)
		return
	}

	// Получаем выбранные варианты
	answers := r.Form["answers"]
	fmt.Printf("📋 answers: %v (len=%d)\n", answers, len(answers))

	// Сохраняем каждый ответ
	for i, answerIDStr := range answers {
		fmt.Printf("   Processing answer %d: %s\n", i, answerIDStr)
		optionID, err := uuid.Parse(answerIDStr)
		if err != nil {
			fmt.Printf("   ❌ uuid.Parse error: %v\n", err)
			continue
		}

		if err := h.sessionService.SubmitAnswer(r.Context(), session.ID, questionID, optionID); err != nil {
			fmt.Printf("   ❌ SubmitAnswer error: %v\n", err)
		} else {
			fmt.Printf("   ✅ Answer saved\n")
		}
	}

	// Получаем обновлённую сессию
	session, err = h.sessionService.GetSessionByID(r.Context(), session.ID)
	if err != nil {
		fmt.Printf("❌ GetSessionByID error: %v\n", err)
		http.Error(w, "Ошибка получения сессии", http.StatusInternalServerError)
		return
	}
	fmt.Printf("📊 Session status: %s, completed: %v\n", session.Status, session.IsCompleted())

	if session.IsCompleted() {
		fmt.Println("🏁 Redirect to result")
		http.Redirect(w, r, "/test/result", http.StatusSeeOther)
		return
	}

	fmt.Println("➡️ Redirect to next question")
	http.Redirect(w, r, "/test", http.StatusSeeOther)
}

// GET /test/result - показать результаты
func (h *TestHandler) ShowResult(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	// Получаем детальные результаты
	results, err := h.getDetailedResults(r.Context(), session)
	if err != nil {
		fmt.Printf("Ошибка получения результатов: %v\n", err)
		http.Error(w, "Ошибка получения результатов", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"results":       results.Questions,
		"correct":       results.CorrectCount,
		"total":         results.TotalQuestions,
		"percentage":    results.Percentage,
		"score":         session.Score,
		"max_score":     session.MaxScore,
		"score_percent": session.ScoreProgress(),
	}

	h.template.ExecuteTemplate(w, "result.html", data)
}

// GET /books/{id}/start - начать новый тест с конкретной книгой
func (h *TestHandler) StartTest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем bookID из URL
	bookIDStr := r.PathValue("id")
	if bookIDStr == "" {
		bookIDStr = r.URL.Query().Get("book_id")
	}

	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		http.Error(w, "Неверный ID книги", http.StatusBadRequest)
		return
	}

	_, err = h.bookService.GetBookByID(r.Context(), bookID)
	if err != nil {
		http.Error(w, "Книга не найдена", http.StatusBadRequest)
		return
	}

	fmt.Printf("StartTest для пользователя %v, книга %v\n", userID, bookID)

	// 1. Проверяем, есть ли незавершённая сессия
	hasUnfinished, existingSession, err := h.sessionService.HasUnfinishedSession(r.Context(), userID)
	if err != nil {
		http.Error(w, "Ошибка проверки сессии", http.StatusInternalServerError)
		return
	}

	if hasUnfinished && existingSession != nil {
		// Продолжаем существующую сессию
		h.setSessionCookie(w, existingSession.ID)
		http.Redirect(w, r, "/test", http.StatusSeeOther)
		return
	}

	// 2. Создаём сессию
	session, err := h.sessionService.StartTest(r.Context(), userID, bookID, 30*time.Minute)
	if err != nil {
		fmt.Printf("Ошибка создания сессии: %v\n", err)
		http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
		return
	}

	// 3. Устанавливаем cookie
	h.setSessionCookie(w, session.ID)

	// 4. Редирект на тест
	http.Redirect(w, r, "/test", http.StatusSeeOther)
}

// POST /test/abandon - бросить тест
func (h *TestHandler) AbandonTest(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	if err := h.sessionService.AbandonSession(r.Context(), session.ID); err != nil {
		fmt.Printf("Ошибка отмены сессии: %v\n", err)
		http.Error(w, "Ошибка отмены теста", http.StatusInternalServerError)
		return
	}

	h.clearSessionCookie(w)
	http.Redirect(w, r, "/books", http.StatusSeeOther)
}

// ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========

func (h *TestHandler) getSession(r *http.Request) (*model.Session, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, fmt.Errorf("no session cookie")
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID")
	}

	session, err := h.sessionService.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (h *TestHandler) setSessionCookie(w http.ResponseWriter, sessionID uuid.UUID) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   1800,
	})
}

func (h *TestHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func (h *TestHandler) getProgress(current, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(current) / float64(total) * 100
}

func (h *TestHandler) completeTest(ctx context.Context, sessionID uuid.UUID) error {
	return h.sessionService.CompleteSession(ctx, sessionID)
}

// DetailedResults детальные результаты
type DetailedResults struct {
	Questions      []QuestionResult `json:"questions"`
	CorrectCount   int              `json:"correct_count"`
	TotalQuestions int              `json:"total_questions"`
	Percentage     float64          `json:"percentage"`
}

type QuestionResult struct {
	QuestionText   string   `json:"question_text"`
	UserAnswers    []string `json:"user_answers"`
	CorrectAnswers []string `json:"correct_answers"`
	IsCorrect      bool     `json:"is_correct"`
}

func (h *TestHandler) getDetailedResults(ctx context.Context, session *model.Session) (*DetailedResults, error) {
	// Получаем все ответы сессии
	answers, err := h.sessionService.GetAnswersBySession(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	// Получаем все вопросы книги
	questions, err := h.bookService.GetQuestionsWithOptions(ctx, session.BookID)
	if err != nil {
		return nil, err
	}

	// Группируем ответы по вопросам
	answersByQuestion := make(map[uuid.UUID][]model.UserAnswer)
	for _, answer := range answers {
		answersByQuestion[answer.QuestionID] = append(answersByQuestion[answer.QuestionID], answer)
	}

	results := make([]QuestionResult, 0, len(questions))
	correctCount := 0

	for _, question := range questions {
		userAnswers := answersByQuestion[question.ID]

		// Получаем тексты ответов пользователя
		userAnswersText := make([]string, 0)
		for _, ua := range userAnswers {
			for _, opt := range question.Options {
				if opt.ID == ua.OptionID {
					userAnswersText = append(userAnswersText, opt.Text)
				}
			}
		}

		// Получаем правильные ответы
		correctAnswersText := make([]string, 0)
		for _, opt := range question.Options {
			if opt.IsCorrect {
				correctAnswersText = append(correctAnswersText, opt.Text)
			}
		}

		// Проверяем, правильный ли ответ
		isCorrect := true
		if len(userAnswers) != len(correctAnswersText) {
			isCorrect = false
		} else {
			for _, ua := range userAnswers {
				if !ua.IsCorrect {
					isCorrect = false
					break
				}
			}
		}

		if isCorrect {
			correctCount++
		}

		results = append(results, QuestionResult{
			QuestionText:   question.Text,
			UserAnswers:    userAnswersText,
			CorrectAnswers: correctAnswersText,
			IsCorrect:      isCorrect,
		})
	}

	total := len(questions)
	percentage := 0.0
	if total > 0 {
		percentage = float64(correctCount) / float64(total) * 100
	}

	return &DetailedResults{
		Questions:      results,
		CorrectCount:   correctCount,
		TotalQuestions: total,
		Percentage:     percentage,
	}, nil
}
