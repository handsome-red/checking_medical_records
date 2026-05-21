package handlers

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

// Global variable for parsed templates
var Templates *template.Template

func init() {
	// Добавляем функцию add1 для увеличения индекса на 1
	funcMap := template.FuncMap{
		"add1": func(i int) int {
			return i + 1
		},
	}

	Templates = template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"))
}
