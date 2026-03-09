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

func (s *Shortener) Create(url string) string {
	id := generateShortID()
	s.repo.Save(id, url)
	return id
}

func (s *Shortener) Get(id string) (string, bool) {
	return s.repo.Get(id)
}
