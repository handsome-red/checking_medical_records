package templates

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

type TemplatesManager struct {
	templates *template.Template
}

var FuncMap = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"mul": func(a, b float64) float64 {
		return a * b
	},
	"div": func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		return a / b
	},
}

func NewTemplatesManager(pattern string) (*TemplatesManager, error) {
	slog.Debug("Loading templates", "pattern", pattern)

	// Создаём новый шаблон с функциями
	tmpl := template.New("").Funcs(FuncMap)

	// Парсим все файлы
	tmpl, err := tmpl.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	slog.Debug("Templates loaded successfully")

	return &TemplatesManager{
		templates: tmpl,
	}, nil
}

func (tm *TemplatesManager) ExecuteTemplate(w http.ResponseWriter, name string, data any) error {
	if tm.templates == nil {
		return fmt.Errorf("templates not initialized")
	}

	// Устанавливаем правильный Content-Type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	return tm.templates.ExecuteTemplate(w, name, data)
}

// GetTemplate возвращает внутренний шаблон (если нужно)
func (tm *TemplatesManager) GetTemplate() *template.Template {
	return tm.templates
}
