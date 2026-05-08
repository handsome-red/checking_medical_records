package handlers

import (
	"encoding/json"
	"log"
	"med_book/internal/service"
	"net/http"

	"html/template"
	// "github.com/gin-gonic/gin"
)

var indexTemplate *template.Template

func init() {
	// Путь указан относительно корня проекта (там, откуда запускается main)
	tmpl, err := template.ParseFiles("internal/handlers/templates/index.html")
	if err != nil {
		log.Fatal("Ошибка загрузки шаблона: ", err)
	}
	indexTemplate = tmpl
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	questions, err := service.NewTest("")
	if err != nil {
		log.Printf("Error loading test: %v", err)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error":   "failed to load questions",
			"details": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = indexTemplate.Execute(w, questions)
	if err != nil {
		log.Printf("Ошибка рендеринга шаблона: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}
