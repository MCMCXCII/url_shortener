package repository

import "errors"

type URLRepository interface {
	Save(id, url string) error
	SaveBatch(items []BatchItem) error
	Get(id string) (string, bool)
	GetByOriginal(original string) (string, bool)
}

type Pinger interface {
	Ping() error
}

type Loader interface {
	Load() error
}

type BatchItem struct {
	ID  string
	URL string
}

var ErrOriginalURLExists = errors.New("original url already exists")
