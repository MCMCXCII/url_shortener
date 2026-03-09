package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *service.Shortener
	cfg     *config.Config
}

func NewHandler(service *service.Shortener, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

func (h *Handler) HandlerPost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	id := h.service.Create(string(body))

	shortURL := h.cfg.BaseURL + "/" + id
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func (h *Handler) HandlerGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	url, ok := h.service.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
