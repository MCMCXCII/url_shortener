package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/logger"
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

func (h *Handler) HandlerPingGet(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Postgres_repo.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

type ShortenRequest struct {
	Url string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func (h *Handler) HandlerJSONPost(w http.ResponseWriter, r *http.Request) {
	var body ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		logger.Log.Debug("json decode error")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	id := h.service.Create(body.Url)

	shortURL := h.cfg.BaseURL + "/" + id
	resp := ShortenResponse{
		Result: shortURL,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
