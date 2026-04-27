package app

import (
	"context"

	"github.com/MCMCXCII/url_shortener/internal/closer"
	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/database"
	"github.com/MCMCXCII/url_shortener/internal/handler"
	"github.com/MCMCXCII/url_shortener/internal/logger"
	"github.com/MCMCXCII/url_shortener/internal/repository"
	"github.com/MCMCXCII/url_shortener/internal/service"
	"go.uber.org/zap"
)

// diContainer — контейнер зависимостей с ленивой инициализацией).

type diContainer struct {
	cfg *config.Config

	// Инфраструктура
	db      database.DB
	storage *repository.FileStorage

	// Репозитории
	repo repository.URLRepository

	// Сервисы
	svc *service.Shortener

	// API
	handler handler.Handler
}

func newDIContainer(cfg *config.Config) *diContainer {
	return &diContainer{cfg: cfg}
}

// DB возвращает подключение к базе данных.
func (d *diContainer) DB() database.DB {
	if d.db == nil && d.cfg.Dsn != "" {
		db, err := database.New(d.cfg.Dsn)
		if err != nil {
			logger.Log.Warn("DB unavailable, falling back to memory", zap.Error(err))
			return nil
		}

		closer.Add("база данных", func(_ context.Context) error {
			return db.Close()
		})

		d.db = db
	}
	return d.db
}

func (d *diContainer) Storage() *repository.FileStorage {
	if d.storage == nil && d.cfg.FileStorage != "" {
		storage, err := repository.NewFileStorage(d.cfg.FileStorage)
		if err != nil {
			logger.Log.Error("failed to initialize storage", zap.Error(err))
		}
		d.storage = storage
	}
	return d.storage
}

func (d *diContainer) Repo() repository.URLRepository {
	if d.repo == nil {
		if db := d.DB(); db != nil {
			logger.Log.Info("use postgres repository")
			d.repo = repository.NewPostgresRepository(db)
		} else {
			logger.Log.Info("use memory repo")
			d.repo = repository.NewMemoryRepository(d.Storage())
		}
		if loader, ok := d.repo.(repository.Loader); ok {
			if err := loader.Load(); err != nil {
				logger.Log.Error("failed to load data", zap.Error(err))
			}
		}
	}
	return d.repo
}

func (d *diContainer) Svc() *service.Shortener {
	if d.svc == nil {
		d.svc = service.NewShortener(d.Repo())
	}
	return d.svc
}

func (d *diContainer) Handler() handler.Handler {
	if d.handler == nil {
		d.handler = handler.NewHandler(d.Svc(), d.cfg)
	}
	return d.handler
}
