package repository

type URLRepository interface {
	Save(id, url string) error
	Get(id string) (string, bool)
}

type Pinger interface {
	Ping() error
}

type Loader interface {
	Load() error
}
