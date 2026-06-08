// internal/quiz/module.go
package quiz

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"med_book/internal/quiz/handlers"
	"med_book/internal/quiz/models"
	"med_book/internal/quiz/repository"
	"med_book/internal/quiz/service"
)

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
	// Загрузка вопросов из файла
	questions, err := loadQuestionsFromFile("pkg/questions/questions.json")
	if err != nil {
		log.Fatalf("can not load questions: %v", err)
	}

	// Преобразование в JSON для совместимости с существующим сервисом
	questionsJSON, err := json.Marshal(questions)
	if err != nil {
		log.Fatalf("can not marshal questions: %v", err)
	}

	// Инициализация квиза
	store, err := repository.NewSQLiteStore("asd")
	if err != nil {
		log.Fatalf("can not create repository: %v", err)
	}
	
	svc, err := service.NewQuizService(store, questionsJSON)
	if err != nil {
		log.Fatalf("can not create quiz service: %v", err)
	}
	
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

// loadQuestionsFromFile загружает вопросы из JSON файла
func loadQuestionsFromFile(filePath string) ([]models.Question, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var questions []models.Question
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}