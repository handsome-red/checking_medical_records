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
	// Parse all templates at startup
	Templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))
}
