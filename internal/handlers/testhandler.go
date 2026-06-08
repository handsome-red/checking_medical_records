// internal/handlers/testhandler.go
package handlers

import (
	"context"
	"fmt"
	"log"
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
		"title":       "Тестирование медицинских книжек",
		"description": "Система для сотрудников Роспотребнадзора по проверке корректности оформления медицинских книжек.",
	}

	if err := h.template.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *TestHandler) ShowTest(w http.ResponseWriter, r *http.Request) {

	log.Println("=== ShowTest START ===")

	session, err := h.getSession(r)
	if err != nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	log.Printf("Session ID: %v, Status: %v, Completed: %v", session.ID, session.Status, session.IsCompleted())

	if session.IsCompleted() || session.Status == model.SessionStatusExpired {
		http.Redirect(w, r, "/test/result", http.StatusSeeOther)
		return
	}

	if session.IsExpired() {
		_ = h.sessionService.ExpireSession(r.Context(), session.ID)
		http.Redirect(w, r, "/test/result?expired=1", http.StatusSeeOther)
		return
	}

	currentQuestion, err := h.sessionService.GetCurrentQuestion(r.Context(), session.ID)
	if err != nil {
		if err.Error() == "all questions answered" {
			http.Redirect(w, r, "/test/result", http.StatusSeeOther)
			return
		}
		http.Error(w, "Ошибка получения вопроса", http.StatusInternalServerError)
		return
	}

	log.Printf("Current question ID: %v, Error: %v", currentQuestion.ID, err)

	book, err := h.bookService.GetBookByID(r.Context(), session.BookID)
	if err != nil {
		http.Error(w, "Книга не найдена", http.StatusInternalServerError)
		return
	}

	answers, err := h.sessionService.GetAnswersBySession(r.Context(), session.ID)
	if err != nil {
		answers = []model.UserAnswer{}
	}

	answeredQuestions, _ := h.countAnsweredQuestions(answers)
	totalQuestions, err := h.bookService.CountQuestions(r.Context(), session.BookID)
	if err != nil {
		totalQuestions = 0
	}

	pages, err := h.bookService.GetPagesByBookID(r.Context(), session.BookID)
	if err != nil {
		pages = []model.BookPage{}
	}

	var expiresAtUnix int64
	if session.ExpiresAt != nil {
		expiresAtUnix = session.ExpiresAt.Unix()
	}

	data := map[string]any{
		"question":       currentQuestion,
		"pages":          pages,
		"current":        answeredQuestions + 1,
		"total":          totalQuestions,
		"progress":       h.getProgress(answeredQuestions, totalQuestions),
		"expiresAt":      expiresAtUnix,
		"book_title":     book.Title,
		"time_remaining": session.GetTimeRemainingFormatted(),
	}

	if err := h.template.ExecuteTemplate(w, "test.html", data); err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
	}
}

func (h *TestHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
    log.Println("=== SubmitAnswer START ===")
    
    session, err := h.getSession(r)
    if err != nil {
        log.Printf("ERROR getting session: %v", err)
        http.Redirect(w, r, "/profile", http.StatusSeeOther)
        return
    }
    log.Printf("Session ID: %v, Status: %v", session.ID, session.Status)

    if session.IsExpired() {
        log.Println("Session expired")
        _ = h.sessionService.ExpireSession(r.Context(), session.ID)
        http.Redirect(w, r, "/test/result?expired=1", http.StatusSeeOther)
        return
    }

    if err := r.ParseForm(); err != nil {
        log.Printf("ERROR parsing form: %v", err)
        http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
        return
    }

    questionIDStr := r.FormValue("question_id")
    log.Printf("Question ID from form: %s", questionIDStr)
    
    if questionIDStr == "" {
        log.Println("ERROR: question_id is empty")
        http.Error(w, "ID вопроса не указан", http.StatusBadRequest)
        return
    }

    questionID, err := uuid.Parse(questionIDStr)
    if err != nil {
        log.Printf("ERROR parsing question ID: %v", err)
        http.Error(w, "Неверный ID вопроса", http.StatusBadRequest)
        return
    }

    optionIDs := make([]uuid.UUID, 0, len(r.Form["answers"]))
    for _, answerIDStr := range r.Form["answers"] {
        optionID, err := uuid.Parse(answerIDStr)
        if err != nil {
            log.Printf("WARNING: skipping invalid answer ID: %s", answerIDStr)
            continue
        }
        optionIDs = append(optionIDs, optionID)
    }
    log.Printf("Saving answer for question %v with %d options", questionID, len(optionIDs))

    if err := h.sessionService.SubmitQuestionAnswer(r.Context(), session.ID, questionID, optionIDs); err != nil {
        log.Printf("ERROR saving answer: %v", err)
        http.Error(w, "Ошибка сохранения ответа", http.StatusInternalServerError)
        return
    }
    log.Println("Answer saved successfully")

    // Получаем обновленную сессию
    session, err = h.sessionService.GetSessionByID(r.Context(), session.ID)
    if err != nil {
        log.Printf("ERROR getting updated session: %v", err)
        http.Error(w, "Ошибка получения сессии", http.StatusInternalServerError)
        return
    }
    log.Printf("Updated session - Status: %v, Completed: %v", session.Status, session.IsCompleted())

    if session.IsCompleted() || session.Status == model.SessionStatusExpired {
        log.Println("Session completed or expired, redirecting to /test/result")
        http.Redirect(w, r, "/test/result", http.StatusSeeOther)
        return
    }

    log.Println("Redirecting to /test")
    http.Redirect(w, r, "/test", http.StatusSeeOther)
}

func (h *TestHandler) ShowResult(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if session.IsInProgress() && session.IsExpired() {
		_ = h.sessionService.ExpireSession(r.Context(), session.ID)
		session, _ = h.sessionService.GetSessionByID(r.Context(), session.ID)
	}

	results, err := h.getDetailedResults(r.Context(), session)
	if err != nil {
		http.Error(w, "Ошибка получения результатов", http.StatusInternalServerError)
		return
	}

	expired := r.URL.Query().Get("expired") == "1" || session.Status == model.SessionStatusExpired

	data := map[string]any{
		"correct":    results.CorrectCount,
		"total":      results.TotalQuestions,
		"percentage": results.Percentage,
		"expired":    expired,
	}

	_ = h.template.ExecuteTemplate(w, "result.html", data)
}

func (h *TestHandler) StartTest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	bookIDStr := r.PathValue("id")
	if bookIDStr == "" {
		bookIDStr = r.URL.Query().Get("book_id")
	}

	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		http.Error(w, "Неверный ID книги", http.StatusBadRequest)
		return
	}

	if _, err = h.bookService.GetBookByID(r.Context(), bookID); err != nil {
		http.Error(w, "Книга не найдена", http.StatusBadRequest)
		return
	}

	hasUnfinished, existingSession, err := h.sessionService.HasUnfinishedSession(r.Context(), userID)
	if err != nil {
		http.Error(w, "Ошибка проверки сессии", http.StatusInternalServerError)
		return
	}

	if hasUnfinished && existingSession != nil {
		if existingSession.BookID == bookID {
			h.setSessionCookie(w, existingSession.ID)
			http.Redirect(w, r, "/test", http.StatusSeeOther)
			return
		}
		_ = h.sessionService.AbandonSession(r.Context(), existingSession.ID)
	}

	session, err := h.sessionService.StartTest(r.Context(), userID, bookID, 30*time.Minute)
	if err != nil {
		http.Error(w, "Ошибка создания сессии: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.setSessionCookie(w, session.ID)
	http.Redirect(w, r, "/test", http.StatusSeeOther)
}

func (h *TestHandler) ExpireTest(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	_ = h.sessionService.ExpireSession(r.Context(), session.ID)
	http.Redirect(w, r, "/test/result?expired=1", http.StatusSeeOther)
}

func (h *TestHandler) AbandonTest(w http.ResponseWriter, r *http.Request) {
	session, err := h.getSession(r)
	if err != nil {
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if err := h.sessionService.AbandonSession(r.Context(), session.ID); err != nil {
		http.Error(w, "Ошибка отмены теста", http.StatusInternalServerError)
		return
	}

	h.clearSessionCookie(w)
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *TestHandler) getSession(r *http.Request) (*model.Session, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, fmt.Errorf("no session cookie")
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID")
	}

	return h.sessionService.GetSessionByID(r.Context(), sessionID)
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

func (h *TestHandler) countAnsweredQuestions(answers []model.UserAnswer) (int, error) {
	seen := make(map[uuid.UUID]struct{})
	for _, answer := range answers {
		seen[answer.QuestionID] = struct{}{}
	}
	return len(seen), nil
}

type DetailedResults struct {
	CorrectCount   int
	TotalQuestions int
	Percentage     float64
}

func (h *TestHandler) getDetailedResults(ctx context.Context, session *model.Session) (*DetailedResults, error) {
	answers, err := h.sessionService.GetAnswersBySession(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	totalQuestions, err := h.bookService.CountQuestions(ctx, session.BookID)
	if err != nil {
		return nil, err
	}

	questions, err := h.bookService.GetQuestionsWithOptions(ctx, session.BookID)
	if err != nil {
		return nil, err
	}

	answersByQuestion := make(map[uuid.UUID][]model.UserAnswer)
	for _, answer := range answers {
		answersByQuestion[answer.QuestionID] = append(answersByQuestion[answer.QuestionID], answer)
	}

	correctCount := session.Score
	if correctCount == 0 && len(answers) > 0 {
		for _, question := range questions {
			userAnswers := answersByQuestion[question.ID]
			if isQuestionCorrect(question, userAnswers) {
				correctCount++
			}
		}
	}

	percentage := 0.0
	if totalQuestions > 0 {
		percentage = float64(correctCount) / float64(totalQuestions) * 100
	}

	return &DetailedResults{
		CorrectCount:   correctCount,
		TotalQuestions: totalQuestions,
		Percentage:     percentage,
	}, nil
}

func isQuestionCorrect(question model.Question, userAnswers []model.UserAnswer) bool {
	correctCount := 0
	for _, opt := range question.Options {
		if opt.IsCorrect {
			correctCount++
		}
	}
	if len(userAnswers) != correctCount {
		return false
	}
	for _, ua := range userAnswers {
		if !ua.IsCorrect {
			return false
		}
	}
	return correctCount > 0
}
