package main

import (
	"log"
	"net/http"

	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/handler"
	"github.com/MCMCXCII/url_shortener/internal/repository"
	"github.com/MCMCXCII/url_shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfig()

	repo := repository.NewMemoryRepository()
	svc := service.NewShortener(repo)
	h := handler.NewHandler(svc, cfg)

	r := chi.NewRouter()
	r.Get("/{id}", h.HandlerGet)
	r.Post("/", h.HandlerPost)
	log.Printf("Server starts: %s", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatal(err)
	}
}
