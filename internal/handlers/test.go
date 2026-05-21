package handlers

import (
	"log"
	"net/http"

	"html/template"
	// "github.com/gin-gonic/gin"
)

var indexTemplate *template.Template

func init() {
	tmpl, err := template.ParseFiles("internal/handlers/templates/index.html")
	if err != nil {
		log.Fatal("Ошибка загрузки шаблона: ", err)
	}
	indexTemplate = tmpl
}

func (h *TestHandler) StartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]any{
		"title":       "Добро пожаловать на тест",
		"description": "Проверьте свои знания о медицинских книжках",
		// "total_questions": len(h.sessionService.),
	}

	Templates.ExecuteTemplate(w, "index.html", data)
}
