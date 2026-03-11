package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "", "Base URL for shortened links")

	flag.Parse()

	if env := os.Getenv("SERVER_ADDRESS"); env != "" {
		cfg.ServerAddress = env
	}

	if env := os.Getenv("BASE_URL"); env != "" {
		cfg.BaseURL = env
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://" + cfg.ServerAddress
	}

	return cfg
}
