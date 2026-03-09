package repository

import "sync"

type URLRepository interface {
	Save(id, url string)
	Get(id string) (string, bool)
}

type MemoryRepository struct {
	store map[string]string
	mu    sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		store: make(map[string]string),
	}
}

func (r *MemoryRepository) Save(id, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[id] = url
}

func (r *MemoryRepository) Get(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RLock()

	url, ok := r.store[id]
	return url, ok
}
