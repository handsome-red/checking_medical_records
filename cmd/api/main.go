package main

import (
	"log"
	"net/http"

	"med_book/internal/config"
	"med_book/internal/database"
	"med_book/internal/handlers"
	"med_book/internal/repository"
	"med_book/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {

	// dbConfig := database.Config{
	// 	User:     "atlas", // ваш пользователь
	// 	Password: "123",   // ваш пароль
	// 	Host:     "localhost",
	// 	Port:     "5432",
	// 	DBName:   "med_book", // имя базы данных
	// }

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

	r := chi.NewRouter()

	r.Handle("/uploads/*", http.StripPrefix("/uploads/",
		http.FileServer(http.Dir("uploads"))))
	r.Handle("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static"))))

	r.Get("/", testHandler.StartHandler)
	r.Route("/api", func(r chi.Router) {
		r.Route("/registration", func(r chi.Router) {
			r.Get("/", registrationHandler.ShowRegistrationForm)
			r.Post("/", registrationHandler.Register)
		})

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
