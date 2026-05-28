package main

import (
	"log"
	"net/http"
	"time"

	"med_book/internal/database"
	"med_book/internal/handlers"
	"med_book/internal/middleware"
	"med_book/internal/repository"
	"med_book/internal/service"
	"med_book/internal/templates"

	"github.com/go-chi/chi/v5"
)

const (
	defaultServerAddress = ":8080"
	shutdownTimeout      = 30 * time.Second
)

func main() {
	// Подключение к БД
	db, err := database.NewDatabaseFromEnv()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Миграция
	if err := database.Migrate(db.GetDB()); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db.GetDB())
	sessionRepo := repository.NewSessionRepository(db.GetDB())
	bookRepo := repository.NewBookRepository(db.GetDB())
	answerRepo := repository.NewAnswerRepository(db.GetDB())

	// Инициализация сервисов
	userService := service.NewUserService(userRepo)
	sessionService := service.NewSessionService(sessionRepo, bookRepo, answerRepo)
	bookService := service.NewBookService(bookRepo)

	// Инициализация шаблонов
	templateManager, err := templates.NewTemplatesManager("internal/handlers/templates/*.html")
	if err != nil {
		log.Fatal("Failed to create TemplateManager:", err)
	}

	// Инициализация хендлеров
	testHandler := handlers.NewTestHandler(sessionService, bookService, templateManager)
	authHandler := handlers.NewAuthHandler(userService, templateManager)
	registerHandler := handlers.NewRegistrationHandler(userService, sessionService, bookService, templateManager.GetTemplate())
	profileHandler := handlers.NewProfileHandler(userService, sessionService, bookService, templateManager)

	// Мидлвары
	authMiddleware := middleware.AuthMiddleware(userService)

	// Роутер
	r := chi.NewRouter()

	// Статические файлы
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Публичные маршруты
	r.Get("/", testHandler.ShowTest)
	r.Get("/login", authHandler.ShowLogin)
	r.Get("/register", registerHandler.ShowRegistrationForm)
	r.Post("/register", registerHandler.Register)
	r.Post("/login", authHandler.Login)

	// Защищённые маршруты
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/logout", authHandler.Logout)

		r.Route("/profile", func(r chi.Router) {
			r.Get("/", profileHandler.GetUserProfile)
		})

		r.Route("/test", func(r chi.Router) {
			r.Get("/", testHandler.ShowTest)
			r.Post("/submit", testHandler.SubmitAnswer)
			r.Get("/result", testHandler.ShowResult)
			r.Post("/abandon", testHandler.AbandonTest)
		})

		r.Route("/books", func(r chi.Router) {
			r.Get("/", bookHandler.GetAllBooks)
			r.Get("/{id}", bookHandler.GetBookByID)
			r.Get("/{id}/start", testHandler.StartTest)
		})
	})

	log.Println("🚀 Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Сервер завершился с ошибкой: %v", err)
	}
}
