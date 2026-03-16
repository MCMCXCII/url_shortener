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
	logger.Initialize(cfg.LogLevel)
	log.Printf("Starting server with config: %+v", cfg)

	// Инициализация хранилища
	storage, err := initStorage(cfg.FileStorage)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Репозиторий
	repo := repository.NewMemoryRepository(storage)
	if err := repo.Load(); err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	// Сервис и хендлеры
	svc := service.NewShortener(repo)
	h := handler.NewHandler(svc, cfg)

	// Роутер и middleware
	r := chi.NewRouter()
	r.Use(middleware.GzipMiddleware)
	r.With(middleware.ResponseLogger).Post("/", h.HandlerPost)
	r.With(middleware.ResponseLogger).Post("/api/shorten", h.HandlerJSONPost)
	r.With(middleware.RequestLogger).Get("/{id}", h.HandlerGet)

	log.Printf("Server listening on %s", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}

func initStorage(filename string) (*repository.FileStorage, error) {
	if filename == "" {
		return nil, nil
	}
	return repository.NewFileStorage(filename)
}
