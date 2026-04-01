package service

import (
	"math/rand"

	"github.com/MCMCXCII/url_shortener/internal/repository"
)

type Shortener struct {
	repo repository.URLRepository
}

func NewShortener(repo repository.URLRepository) *Shortener {
	return &Shortener{repo: repo}
}

func generateShortID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 8)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func (s *Shortener) Create(url string) (string, error) {
	id := generateShortID()
	err := s.repo.Save(id, url)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Shortener) Get(id string) (string, bool) {
	return s.repo.Get(id)
}

func (s *Shortener) Ping() error {
	if pinger, ok := s.repo.(repository.Pinger); ok {
		return pinger.Ping()
	}
	return nil
}

type BatchInput struct {
	CorrelationID string
	OriginalURL   string
}

type BatchOutput struct {
	CorrelationID string
	ShortURL      string
}

func (s *Shortener) CreateBatch(items []BatchInput, baseURL string) ([]BatchOutput, error) {
	repoItems := make([]repository.BatchItem, 0, len(items))
	resp := make([]BatchOutput, 0, len(items))

	for _, item := range items {
		id := generateShortID()

		repoItems = append(repoItems, repository.BatchItem{
			ID:  id,
			URL: item.OriginalURL,
		})

		resp = append(resp, BatchOutput{
			CorrelationID: item.CorrelationID,
			ShortURL:      baseURL + "/" + id,
		})
	}

	if err := s.repo.SaveBatch(repoItems); err != nil {
		return nil, err
	}

	return resp, nil
}
