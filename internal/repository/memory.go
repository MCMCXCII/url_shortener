package repository

import (
	"strconv"
	"sync"

	"github.com/MCMCXCII/url_shortener/internal/logger"
	"go.uber.org/zap"
)

type MemoryRepository struct {
	store map[string]string
	mu    sync.RWMutex
	file  *FileStorage
	count int
}

func NewMemoryRepository(file *FileStorage) *MemoryRepository {
	return &MemoryRepository{
		store: make(map[string]string),
		file:  file,
	}
}

func (r *MemoryRepository) Save(id, url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[id] = url

	if r.file != nil {
		r.count++
		record := FileRecord{
			UUID:        strconv.Itoa(r.count),
			ShortURL:    id,
			OriginalURL: url,
		}
		_ = r.file.WriteToFile(record)
	}
	return nil
}

func (r *MemoryRepository) Get(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, ok := r.store[id]
	return url, ok
}

func (r *MemoryRepository) Load() error {
	if r.file == nil {
		return nil
	}

	records, err := LoadFromFile(r.file.filename)
	if err != nil {
		logger.Log.Debug("error read file", zap.String("file", r.file.filename))
		return err
	}
	for _, rec := range records {
		r.store[rec.ShortURL] = rec.OriginalURL

		id, err := strconv.Atoi(rec.UUID)
		if err == nil && id > r.count {
			r.count = id
		}
	}

	return nil

}

func (r *MemoryRepository) SaveBatch(items []BatchItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range items {
		r.store[item.ID] = item.URL

		if r.file != nil {
			r.count++
			record := FileRecord{
				UUID:        strconv.Itoa(r.count),
				ShortURL:    item.ID,
				OriginalURL: item.URL,
			}
			_ = r.file.WriteToFile(record)
		}
	}
	return nil
}

func (r *MemoryRepository) GetByOriginal(original string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for short, url := range r.store {
		if url == original {
			return short, true
		}
	}
	return "", false
}
