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
	defaultAddress := "localhost:8080"
	defaultBaseURL := ""
	flag.StringVar(&cfg.ServerAddress, "a", defaultAddress, "HTTP server address (host:port)")
	flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL for shortened links")

	flag.Parse()

	// сначала проверяем переменные окружения
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.ServerAddress = envAddr
	}

	if envBase := os.Getenv("BASE_URL"); envBase != "" {
		cfg.BaseURL = envBase
	}

	// если BaseURL всё ещё пустой — строим по ServerAddress
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://" + cfg.ServerAddress
	}
	return cfg
}
