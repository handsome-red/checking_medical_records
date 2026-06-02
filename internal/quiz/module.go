// internal/quiz/module.go
package quiz

import (
	"embed"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"med_book/internal/quiz/handlers"
	"med_book/internal/quiz/repository"
	"med_book/internal/quiz/service"
)

//go:embed questions.json
var questionsJSON []byte

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Module struct {
	router chi.Router
}

func NewModule() *Module {
	m := &Module{
		router: chi.NewRouter(),
	}

	m.setupRoutes()
	return m
}

func (m *Module) Router() chi.Router {
	return m.router
}

func (m *Module) setupRoutes() {
	// Инициализация квиза
	store, err := repository.NewSQLiteStore("asd")
	if err != nil {
		log.Fatalf("can not create repository")
	}
	svc, _ := service.NewQuizService(store, questionsJSON)
	handler := handlers.NewWebHandler(svc, templatesFS)

	// Статика квиза
	m.router.Handle("/static/*", http.StripPrefix("/quiz/static/",
		http.FileServer(http.FS(staticFS))))

	// Маршруты квиза (все с префиксом /quiz уже добавлен в main)
	m.router.Get("/", handler.Index)             // /quiz/
	m.router.Get("/reg", handler.Registration)   // /quiz/reg
	m.router.Get("/start", handler.Start)        // /quiz/start
	m.router.Get("/{id}", handler.Quiz)          // /quiz/1
	m.router.Post("/submit", handler.Submit)     // /quiz/submit
	m.router.Get("/result/{id}", handler.Result) // /quiz/result/123
	m.router.Get("/download", handler.Download)  // /quiz/download
}
