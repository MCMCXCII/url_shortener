package main

import (
	"log"
	"net/http"

	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/handler"
	"github.com/MCMCXCII/url_shortener/internal/logger"
	"github.com/MCMCXCII/url_shortener/internal/middleware"
	"github.com/MCMCXCII/url_shortener/internal/repository"
	"github.com/MCMCXCII/url_shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfig()

	repo := repository.NewMemoryRepository()
	svc := service.NewShortener(repo)
	h := handler.NewHandler(svc, cfg)
	logger.Initialize(cfg.LogLevel)

	r := chi.NewRouter()

	r.With(middleware.ResponseLogger).Post("/", h.HandlerPost)
	r.With(middleware.ResponseLogger).Post("/api/shorten", h.HandlerJSONPost)
	r.With(middleware.RequestLogger).Get("/{id}", h.HandlerGet)

	r.Use(middleware.GzipMiddleware)

	log.Printf("Server starts: %s", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatal(err)
	}
}
