package main

import (
	"log/slog"
	"net/http"
	"med_book/pkg/logger"
	"med_book/internal/config"
	"med_book/internal/handlers"
)

func main() {
	
	logger.Init()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return
	}
	slog.Info("Config loaded successfully", "port", cfg.Server.Port)

	http.HandleFunc("/", handlers.HelloWorldHandler)

	slog.Info("Server starting", "port", cfg.Server.Port)
	if err := http.ListenAndServe(cfg.Server.Port, nil); err != nil {
		slog.Error("Server failed to start", "error", err)
		return
	}
}