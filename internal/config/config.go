package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	LogLevel      string
	FileStorage   string
	Dsn           string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "172.19.9.107:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "Base URL for shortened links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Level log")
	flag.StringVar(&cfg.FileStorage, "f", "/tmp/short-url-db.json", "save to file")
	flag.StringVar(&cfg.Dsn, "d", "", "addres BD")

	flag.Parse()

	if env := os.Getenv("SERVER_ADDRESS"); env != "" {
		cfg.ServerAddress = env
	}

	if env := os.Getenv("BASE_URL"); env != "" {
		cfg.BaseURL = env
	}

	if env := os.Getenv("FILE_STORAGE_PATH"); env != "" {
		cfg.FileStorage = env
	}

	if env := os.Getenv("DATABASE_DSN"); env != "" {
		cfg.Dsn = env
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://" + cfg.ServerAddress
	}

	return cfg
}
