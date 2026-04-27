package main

import (
	"os"

	"github.com/MCMCXCII/url_shortener/internal/app"
	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	a := app.New(cfg)

	if err := a.Run(); err != nil {
		logger.Log.Error("ошибка приложения", zap.Error(err))
		os.Exit(1)
	}
}
