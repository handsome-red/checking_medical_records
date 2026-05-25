package handlers

import (
	"net/http"
)

func (h *TestHandler) StartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]any{
		"title":       "Добро пожаловать на тест",
		"description": "Проверьте свои знания о медицинских книжках",
		// "total_questions": len(h.sessionService.),
	}

	h.template.ExecuteTemplate(w, "index.html", data)
}
