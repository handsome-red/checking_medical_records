package main

import (
	"log"
	"net/http"
	"time"

	"med_book/internal/config"
	"med_book/internal/database"
	"med_book/internal/handlers"
	"med_book/internal/middleware"
	"med_book/internal/repository"
	"med_book/internal/service"

	"github.com/go-chi/chi/v5"
)

const (
	// defaultServerAddress - адрес сервера по умолчанию
	defaultServerAddress = ":8080"

	// shutdownTimeout - таймаут для graceful shutdown
	shutdownTimeout = 30 * time.Second

	// uploadsPath - путь к директории с загруженными файлами
	uploadsPath = "/uploads/*"

	// staticPath - путь к статическим файлам
	staticPath = "/static/*"

	// uploadsDir - директория для загрузок
	uploadsDir = "uploads"

	// staticDir - директория со статикой
	staticDir = "./static"
)

func main() {
	dbConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to get config:", err)
	}

	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Автоматическая миграция (создание таблиц)
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	userService := service.NewUserService(userRepo)
	testService, err := service.NewTestService("")
	if err != nil {
		log.Fatal("Failed to create testService:", err)
	}
	questionService := service.NewQuestionService(db)
	sessionService := service.NewSessionService(sessionRepo, testService)

	registrationHandler := handlers.NewRegistrationHandler(userService, sessionService, questionService, testService)
	testHandler := handlers.NewTestHandler(testService, sessionService)
	profileService := service.NewProfileService(sessionRepo, sessionService)
	profileHandler := handlers.NewProfileHandler(profileService, sessionService, userService, testService)
	authService := service.NewAuthService("my-secret-key")
	middleware := middleware.AuthMiddleware(authService)
	authHandler := handlers.NewAuthHandler(authService, userService)

	r := chi.NewRouter()

	r.Handle("/uploads/*", http.StripPrefix("/uploads/",
		http.FileServer(http.Dir("uploads"))))
	r.Handle("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static"))))

	r.Get("/", testHandler.StartHandler)
	r.Get("/login", authHandler.ShowLoginPage)
	r.Route("/api", func(r chi.Router) {
		r.Route("/registration", func(r chi.Router) {
			r.Get("/", registrationHandler.ShowRegistrationForm)
			r.Post("/", registrationHandler.Register)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware)

		r.Route("/profile", func(r chi.Router) {
			r.Get("/", profileHandler.GetUserProfile)
			r.Get("/sessions", profileHandler.GetUserSessions)
			r.Get("/session/{id}", profileHandler.GetSessionDetails)
		})

		r.Route("/test", func(r chi.Router) {
			r.Get("/", testHandler.ShowTest)
			r.Post("/submit", testHandler.SubmitAnswer)
			r.Get("/result", testHandler.ShowResult)
		})
	})

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Сервер завершился с ошибкой: %v", err)
	}
}
