package repository

type URLRepository interface {
	Save(id, url string) error
	SaveBatch(items []BatchItem) error
	Get(id string) (string, bool)
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
