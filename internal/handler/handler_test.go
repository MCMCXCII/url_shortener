package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/repository"
	"github.com/MCMCXCII/url_shortener/internal/service"
)

func TestHandlerPost(t *testing.T) {
	body := "http://practicticum.yandex.ru/"

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	repo := repository.NewMemoryRepository()
	svc := service.NewShortener(repo)
	cfg := config.NewConfig()
	h := NewHandler(svc, cfg)

	h.HandlerPost(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", res.StatusCode)
	}
	if !strings.HasPrefix(w.Body.String(), "http://localhost:8080/") {
		t.Errorf("wrong short url format")
	}
}

func TestHandlerGet(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := service.NewShortener(repo)
	cfg := config.NewConfig()
	originalURL := "http://practicticum.yandex.ru/"
	id := svc.Create(originalURL)

	h := NewHandler(svc, cfg)
	req := httptest.NewRequest(
		http.MethodGet,
		"/"+id,
		nil,
	)
	w := httptest.NewRecorder()
	h.HandlerGet(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected 307, got %d", res.StatusCode)
	}

	if res.Header.Get("Location") != originalURL {
		t.Errorf("wrong Location header")
	}
}
