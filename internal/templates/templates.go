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

func NewTemplatesManager(pattern string) (*TemplatesManager, error) {
	slog.Debug("Loading templates", "pattern", pattern)

	tmpl, err := template.ParseGlob(pattern)
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
	return tm.templates.ExecuteTemplate(w, name, data)
}
