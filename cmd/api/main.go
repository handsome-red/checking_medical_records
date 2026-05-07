package main

import (
	// "log/slog"
	"net/http"

	// "med_book/pkg/logger"
	"med_book/internal/handlers"
	"github.com/go-chi/chi/v5"
    // "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
    mux *http.ServeMux
}

func main() {
	r := chi.NewRouter()

	r.Get("/", handlers.HelloWorldHandler)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	http.ListenAndServe(":8080", r)
}