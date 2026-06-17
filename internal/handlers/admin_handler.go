package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"med_book/internal/middleware"
	"med_book/internal/service"
	"med_book/internal/templates"

	"github.com/google/uuid"
)

type AdminHandler struct {
	userService    *service.UserService
	sessionService *service.SessionService
	adminService   *service.AdminService
	bookService    *service.BookService
	template       *templates.TemplatesManager
}

func NewAdminHandler(
	userService *service.UserService,
	sessionService *service.SessionService,
	adminService *service.AdminService,
	bookService *service.BookService,
	template *templates.TemplatesManager,
) *AdminHandler {
	return &AdminHandler{
		userService:    userService,
		sessionService: sessionService,
		adminService:   adminService,
		bookService:    bookService,
		template:       template,
	}
}

func (h *AdminHandler) ShowAdminPanel(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	userFilter, bookFilter := parseAdminFilters(r)

	users, err := h.userService.GetAllUsers(ctx)
	if err != nil {
		http.Error(w, "Ошибка загрузки пользователей", http.StatusInternalServerError)
		return
	}

	books, err := h.bookService.GetAllBooks(ctx)
	if err != nil {
		http.Error(w, "Ошибка загрузки тестов", http.StatusInternalServerError)
		return
	}

	reports, err := h.sessionService.GetAdminReports(ctx, userFilter, bookFilter)
	if err != nil {
		http.Error(w, "Ошибка загрузки результатов", http.StatusInternalServerError)
		return
	}

	viewReports := make([]AdminReportView, 0, len(reports))
	for _, report := range reports {
		endTime := ""
		duration := ""
		if report.CompletedAt != nil {
			endTime = report.CompletedAt.Format("15:04")
			duration = report.CompletedAt.Sub(report.StartedAt).String()
		}

		// answers := make([]string, 0, len(report.Answers))
		// for _, answer := range report.Answers {
		// 	text := joinSelected(answer.SelectedTexts)
		// 	if answer.QuestionText != "" {
		// 		text = answer.QuestionText + ": " + text
		// 	}
		// 	answers = append(answers, text)
		// }

		questions := make([]QuestionView, 0, len(report.Answers))
		for _, answer := range report.Answers {

			answerViews := make([]AnswerView, 0, len(answer.SelectedTexts))

			for _, textReport := range answer.SelectedTexts {
				answerViews = append(answerViews, AnswerView{
					Text:      textReport.Text,
					IsCorrect: textReport.IsCorrect,
				})
			}

			questions = append(questions, QuestionView{
				Text:    answer.QuestionText,
				Answers: answerViews,
				// Correct: allCorrect,
			})
		}

		viewReports = append(viewReports, AdminReportView{
			FIO:       fmt.Sprintf("%s %s %s", report.LastName, report.FirstName, report.Patronymic),
			Date:      report.StartedAt.Format("02.01.2006"),
			StartTime: report.StartedAt.Format("15:04"),
			EndTime:   endTime,
			Duration:  duration,
			BookTitle: report.BookTitle,
			Score:     report.Score,
			MaxScore:  report.MaxScore,
			Questions: questions,
		})
	}

	selectedUserID := ""
	selectedBookID := ""
	if userFilter != nil {
		selectedUserID = userFilter.String()
	}
	if bookFilter != nil {
		selectedBookID = bookFilter.String()
	}

	exportParams := url.Values{}
	if selectedUserID != "" {
		exportParams.Set("user_id", selectedUserID)
	}
	if selectedBookID != "" {
		exportParams.Set("book_id", selectedBookID)
	}
	exportURL := "/admin/export"
	if encoded := exportParams.Encode(); encoded != "" {
		exportURL += "?" + encoded
	}

	data := map[string]any{
		"Users":          users,
		"Books":          books,
		"Reports":        viewReports,
		"SelectedUserID": selectedUserID,
		"SelectedBookID": selectedBookID,
		"ExportURL":      exportURL,
	}

	if err := h.template.ExecuteTemplate(w, "admin.html", data); err != nil {
		http.Error(w, "Ошибка рендеринга", http.StatusInternalServerError)
	}
}

func (h *AdminHandler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}

	userFilter, bookFilter := parseAdminFilters(r)
	reports, err := h.sessionService.GetAdminReports(r.Context(), userFilter, bookFilter)
	if err != nil {
		http.Error(w, "Ошибка загрузки результатов", http.StatusInternalServerError)
		return
	}

	data, err := h.adminService.ExportReportsExcel(reports)
	if err != nil {
		http.Error(w, "Ошибка формирования Excel", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("results_%s.xlsx", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *AdminHandler) isAdmin(r *http.Request) bool {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		return false
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		return false
	}

	return user.IsAdmin()
}

func parseAdminFilters(r *http.Request) (*uuid.UUID, *uuid.UUID) {
	var userFilter *uuid.UUID
	var bookFilter *uuid.UUID

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			userFilter = &id
		}
	}

	if bookIDStr := r.URL.Query().Get("book_id"); bookIDStr != "" {
		if id, err := uuid.Parse(bookIDStr); err == nil {
			bookFilter = &id
		}
	}

	return userFilter, bookFilter
}

// AnswerView представляет один ответ пользователя
type AnswerView struct {
	Text      string `json:"text"`       // Текст ответа
	IsCorrect bool   `json:"is_correct"` // Правильный ли ответ
}

// QuestionView представляет вопрос с ответами пользователя
type QuestionView struct {
	Text    string       `json:"text"`    // Текст вопроса
	Answers []AnswerView `json:"answers"` // Ответы пользователя
	// Correct bool         `json:"correct"` // Правильно ли отвечен вопрос (все ответы верны)
}

// AdminReportView представляет отчет по сессии
type AdminReportView struct {
	FIO       string         `json:"fio"`
	Date      string         `json:"date"`
	StartTime string         `json:"start_time"`
	EndTime   string         `json:"end_time"`
	Duration  string         `json:"duration"`
	BookTitle string         `json:"book_title"`
	Score     int            `json:"score"`
	MaxScore  int            `json:"max_score"`
	Questions []QuestionView `json:"questions"`
}

// func joinSelected(texts []string) QuestionView {
// 	if len(texts) == 0 {
// 		return "—"
// 	}
// 	result := texts[0]
// 	for i := 1; i < len(texts); i++ {
// 		result += ", " + texts[i]
// 	}
// 	return result
// }
