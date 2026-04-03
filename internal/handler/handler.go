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

	id, exists, err := h.service.Create(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortURL := h.cfg.BaseURL + "/" + id
	w.Header().Set("Content-Type", "text/plain")
	if exists {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
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
	if err := h.service.Ping(); err != nil {
		http.Error(w, "database connection error", http.StatusInternalServerError)
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
	id, exists, err := h.service.Create(body.Url)
	if err != nil {
		logger.Log.Debug("db select error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	shortURL := h.cfg.BaseURL + "/" + id
	resp := ShortenResponse{
		Result: shortURL,
	}
	w.Header().Set("Content-Type", "application/json")
	if exists {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (h *Handler) HandlerBatchPost(w http.ResponseWriter, r *http.Request) {
	var req []BatchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, "empty batch", http.StatusBadRequest)
		return
	}

	// преобразуем в service формат
	input := make([]service.BatchInput, 0, len(req))
	for _, item := range req {
		input = append(input, service.BatchInput{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
		})
	}

	// вызываем сервис
	result, err := h.service.CreateBatch(input, h.cfg.BaseURL)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// обратно в handler формат
	resp := make([]BatchResponse, 0, len(result))
	for _, item := range result {
		resp = append(resp, BatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      item.ShortURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(resp)
}
