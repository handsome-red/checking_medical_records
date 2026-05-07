package server

import (
    "context"
    "net/http"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "go.uber.org/fx"
    
    "med_book/internal/config"
)

type Server struct {
	router	*chi.Mux
	config	*config.Config
	http 	*http.Server
}

func NewServer(lc fx.Lifecycle, cfg *config.Config) *Server {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)

	srv := &Server{
		router: router,
		config:	cfg,
		http:	&http.Server{
			Addr:         cfg.Server.Port,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					// Logging
				}
			} ()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.http.Shutdown(ctx)
		},
	})

	return srv
}

func (s *Server) RegisterRoutes(register func(chi.Route)) {
	register(s.router)
}

func (s *Server) Router() *chi.Mux {
	return s.router
}