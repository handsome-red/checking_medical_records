package main

import (
	// "log/slog"
	"net/http"

	// "med_book/pkg/logger"
	"med_book/internal/handlers"
	"med_book/internal/service"

	"github.com/go-chi/chi/v5"
	// "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	mux *http.ServeMux
}

func main() {
	userService := service.NewUserService()
	registrationHandler := handlers.NewRegistrationHandler(userService)

	r := chi.NewRouter()

	r.Get("/", handlers.TestHandler)
	r.Get("/registration", registrationHandler.ShowRegistrationForm)
	r.Post("/registration", registrationHandler.Register)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	http.ListenAndServe(":8080", r)
}
