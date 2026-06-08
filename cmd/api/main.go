package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"med_book/internal/config"
	"med_book/internal/database"
	_import "med_book/internal/import"
	"med_book/internal/handlers"
	"med_book/internal/middleware"
	"med_book/internal/repository"
	"med_book/internal/service"
	"med_book/internal/templates"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	importService := _import.NewImportService(db.GetDB())
	if err := importService.ImportBooksIfNotExists("pkg/questions/questions.json"); err != nil {
		log.Printf("Warning: failed to import questions: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db.GetDB())
	bookRepo := repository.NewBookRepository(db.GetDB())
	answerRepo := repository.NewAnswerRepository(db.GetDB())

	authService := service.NewAuthService(cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	sessionService := service.NewSessionService(sessionRepo, bookRepo, answerRepo)
	bookService := service.NewBookService(bookRepo)
	adminService := service.NewAdminService(sessionService)

	ctx := context.Background()
	if cfg.AdminEmail != "" {
		if err := userService.PromoteToAdmin(ctx, cfg.AdminEmail); err != nil {
			log.Printf("Admin promotion skipped for %s: %v", cfg.AdminEmail, err)
		} else {
			log.Printf("User %s promoted to admin", cfg.AdminEmail)
		}
	}

	templateManager, err := templates.NewTemplatesManager("internal/handlers/templates/*.html")
	if err != nil {
		log.Fatal("Failed to create TemplateManager:", err)
	}

	testHandler := handlers.NewTestHandler(sessionService, bookService, templateManager)
	authHandler := handlers.NewAuthHandler(authService, userService, cfg.AdminEmail, templateManager)
	registerHandler := handlers.NewRegistrationHandler(userService, authService, cfg.AdminEmail, templateManager)
	profileHandler := handlers.NewProfileHandler(userService, sessionService, bookService, templateManager)
	adminHandler := handlers.NewAdminHandler(userService, sessionService, adminService, bookService, templateManager)

	authMiddleware := middleware.AuthMiddleware(authService)

	r := chi.NewRouter()

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	r.Get("/", testHandler.ShowStartPage)
	r.Get("/login", authHandler.ShowLoginPage)
	r.Post("/login", authHandler.Login)
	r.Get("/register", registerHandler.ShowRegistrationForm)
	r.Post("/register", registerHandler.Register)
	r.Get("/api/registration", registerHandler.ShowRegistrationForm)
	r.Post("/api/registration", registerHandler.Register)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/logout", authHandler.Logout)
		r.Get("/profile", profileHandler.GetUserProfile)

		r.Route("/test", func(r chi.Router) {
			r.Get("/", testHandler.ShowTest)
			r.Post("/submit", testHandler.SubmitAnswer)
			r.Get("/result", testHandler.ShowResult)
			r.Get("/expire", testHandler.ExpireTest)
			r.Post("/abandon", testHandler.AbandonTest)
		})

		r.Get("/books/{id}/start", testHandler.StartTest)

		r.Route("/admin", func(r chi.Router) {
			r.Get("/", adminHandler.ShowAdminPanel)
			r.Get("/export", adminHandler.ExportExcel)
		})
	})

	addr := ":" + cfg.ServerPort
	log.Printf("Server started on http://localhost%s", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout: 60 * time.Second, // Увеличили с 30 до 60
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}
