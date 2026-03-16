package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/repository"
	"github.com/MCMCXCII/url_shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func TestHandlerGet(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		FileStorage:   "/tmp/short-url-db.json",
	}
	storage, err := repository.NewFileStorage(cfg.FileStorage)
	if err != nil {
		t.Errorf("error file")
	}
	repo := repository.NewMemoryRepository(storage)
	svc := service.NewShortener(repo)
	originalURL := "http://practicticum.yandex.ru/"

	// Create возвращает ID
	id := svc.Create(originalURL)

	h := NewHandler(svc, cfg)

	r := chi.NewRouter()
	r.Get("/{id}", h.HandlerGet)
	req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()

	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected 307, got %d", res.StatusCode)
	}

	if res.Header.Get("Location") != originalURL {
		t.Errorf("wrong Location header, got %s", res.Header.Get("Location"))
	}
}

func TestHandlerJSONPost(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		FileStorage:   "/tmp/short-url-db.json",
	}
	storage, err := repository.NewFileStorage(cfg.FileStorage)
	if err != nil {
		t.Errorf("error file")
	}
	repo := repository.NewMemoryRepository(storage)
	svc := service.NewShortener(repo)
	h := NewHandler(svc, cfg)
	jsonBody := `{"url":"http://practicticum.yandex.ru/"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.HandlerJSONPost(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	expectedPrefix := cfg.BaseURL + "/"
	if !strings.Contains(string(body), expectedPrefix) {
		t.Errorf("expected response to contain %q, got %s", expectedPrefix, string(body))
	}
}
