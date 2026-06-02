package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"errors"
	"med_book/internal/quiz/models"
	"med_book/internal/quiz/repository"
	"med_book/internal/quiz/service"
)

type WebHandler struct {
	svc  *service.QuizService
	tmpl *template.Template
}

func NewWebHandler(svc *service.QuizService, tmpl *template.Template) *WebHandler {
	return &WebHandler{svc: svc, tmpl: tmpl}
}

// Video — показ видео и переход к тестированию
func (h *WebHandler) Video(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.render(w, "start.html", nil)
}

// Index — форма регистрации
func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/reg" {
		http.NotFound(w, r)
		return
	}
	h.render(w, "index.html", nil)
}

// Start — обработка формы регистрации, создание сессии
func (h *WebHandler) Start(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	videoAgreement := true
	if len(r.FormValue("video_agreement")) == 0 {
		videoAgreement = false
	}

	grade, err := strconv.Atoi(r.FormValue("grade"))
	if err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}
	p := models.Participant{
		LastName:    strings.TrimSpace(r.FormValue("last_name")),
		FirstName:   strings.TrimSpace(r.FormValue("first_name")),
		MiddleName:  strings.TrimSpace(r.FormValue("middle_name")),
		District:    strings.TrimSpace(r.FormValue("district")),
		City:        strings.TrimSpace(r.FormValue("city")),
		School:      strings.TrimSpace(r.FormValue("school")),
		Grade:       grade,
		VideoAgreed: videoAgreement,
	}

	// Валидация обязательных полей
	if p.LastName == "" || p.FirstName == "" || p.District == "" || p.City == "" || p.School == "" || p.Grade == "" || p.VideoAgreed == "" {
		h.render(w, "index.html", map[string]interface{}{
			"Error":       "Пожалуйста, заполните все обязательные поля.",
			"Participant": p,
		})
		return
	}

	sess := h.svc.CreateSession(p)
	http.Redirect(w, r, "/quiz/"+sess.ID+"/0", http.StatusSeeOther)
}

// Quiz — показ вопроса
func (h *WebHandler) Quiz(w http.ResponseWriter, r *http.Request) {
	// Парсим /quiz/{sessionID}/{index}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/quiz/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	sessionID := parts[0]
	idx, err := strconv.Atoi(parts[1])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	sess, ok := h.svc.GetSession(sessionID)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if idx >= len(sess.Questions) {
		http.Redirect(w, r, "/result/"+sessionID, http.StatusSeeOther)
		return
	}

	q := sess.Questions[idx]
	// Определяем тип вопроса: если несколько правильных — checkbox, иначе radio
	correctCount := 0
	for _, opt := range q.Options {
		if opt.IsCorrect {
			correctCount++
		}
	}

	// Ранее выбранные ответы для кнопки "Назад"
	prevAnswers := sess.Answers[q.ID]

	var remainingSeconds int
	elapsed := time.Since(sess.StartedAt)
	if elapsed >= service.TestTimeLimit {
		remainingSeconds = 0
	} else {
		remainingSeconds = int((service.TestTimeLimit - elapsed).Seconds())
	}

	totalSeconds := service.TestTimeLimit.Seconds()
	remainingPercent := float64(remainingSeconds) / float64(totalSeconds) * 100

	data := map[string]interface{}{
		"SessionID":        sessionID,
		"Question":         q,
		"Index":            idx,
		"Total":            len(sess.Questions),
		"Num":              idx + 1,
		"IsMulti":          correctCount > 1,
		"PrevAnswers":      prevAnswers,
		"HasPrev":          idx > 0,
		"RemainingSeconds": remainingSeconds,
		"TotalSeconds":     totalSeconds,
		"RemainingPercent": remainingPercent,
	}
	h.render(w, "quiz.html", data)
}

// SubmitAnswer — сохранение ответа и переход к следующему вопросу
func (h *WebHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}

	sessionID := r.FormValue("session_id")
	questionIDStr := r.FormValue("question_id")
	// indexStr := r.FormValue("index")

	questionID, _ := strconv.Atoi(questionIDStr)
	// index, _ := strconv.Atoi(indexStr)

	// Собираем выбранные варианты
	rawAnswers := r.Form["answer"]
	var optionIDs []int
	for _, v := range rawAnswers {
		id, err := strconv.Atoi(v)
		if err == nil {
			optionIDs = append(optionIDs, id)
		}
	}

	sess, ok := h.svc.GetSession(sessionID)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Сохраняем ответ
	err := h.svc.SaveAnswer(sessionID, questionID, optionIDs)
	if err != nil {
		if errors.Is(err, alert.ErrTimeLimitOut) {
			// При таймауте сохраняем результаты и показываем страницу с результатами
			sess, _ := h.svc.GetSession(sessionID)
			if sess != nil {
				if appendErr := repository.AppendResult(sess); appendErr != nil {
					log.Printf("csv write error on timeout: %v", appendErr)
				}
			}
			http.Redirect(w, r, "/result/"+sessionID+"?timeout=1", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return
	}

	// Обновляем текущий индекс
	nextIndex := sess.Current

	if nextIndex >= len(sess.Questions) {
		if err := repository.AppendResult(sess); err != nil {
			log.Printf("csv write error: %v", err)
		}
		http.Redirect(w, r, "/result/"+sessionID, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/quiz/"+sessionID+"/"+strconv.Itoa(nextIndex), http.StatusSeeOther)
}

// Result — страница результатов
func (h *WebHandler) Result(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/result/")
	sess, ok := h.svc.GetSession(sessionID)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	isTimeout := r.URL.Query().Get("timeout") == "1"

	data := map[string]interface{}{
		"IsTimeout":   isTimeout,
		"IsCompleted": !sess.CompleteAt.IsZero(),
	}

	h.render(w, "result.html", data)
}

// Download — загрузка таблицы
func (h *WebHandler) Download(w http.ResponseWriter, r *http.Request) {
	filePath := "./results.csv"

	// Запрещаем кэширование
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.Header().Set("Content-Disposition", "attachment; filename=results.csv")
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")

	http.ServeFile(w, r, filePath)
}

func (h *WebHandler) render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Ошибка шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}
