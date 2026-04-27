package app

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/MCMCXCII/url_shortener/internal/closer"
	"github.com/MCMCXCII/url_shortener/internal/config"
	"github.com/MCMCXCII/url_shortener/internal/logger"
	"go.uber.org/zap"
)

type App struct {
	diContainer *diContainer
	httpServer  *http.Server
}

func New(cfg *config.Config) *App {
	a := &App{
		diContainer: newDIContainer(cfg),
	}

	a.initDeps()

	return a
}

// initDeps последовательно вызывает функции инициализации.
func (a *App) initDeps() {
	inits := []func(){
		a.initHTTPServer,
	}

	for _, fn := range inits {
		fn()
	}
}

// initHTTPServer создаёт HTTP-сервер.
func (a *App) initHTTPServer() {
	a.httpServer = &http.Server{
		Addr:    a.diContainer.cfg.ServerAddress,
		Handler: a.diContainer.Handler().Routes(),
	}
}

// Run запускает HTTP-сервер с graceful shutdown.
//
// Что происходит:
//  1. signal.NotifyContext перехватывает SIGINT (Ctrl+C) и SIGTERM (Kubernetes)
//  2. HTTP-сервер запускается в отдельной горутине
//  3. Main-горутина ждёт сигнал через <-ctx.Done()
//  4. При сигнале: server.Shutdown дожидается текущих запросов
//  5. closer.CloseAll закрывает все ресурсы в обратном порядке (LIFO)
//
// Паттерн "двойной Ctrl+C":
//   - Первый Ctrl+C → graceful shutdown
//   - stop() снимает custom handler после первого сигнала
//   - Второй Ctrl+C → ОС убивает процесс мгновенно (для разработки, когда shutdown завис)
func (a *App) Run() error {
	// 1. Перехват сигналов.
	// signal.NotifyContext создаёт канал с ёмкостью 1 (буферизованный).
	// Если бы канал был unbuffered — сигнал мог бы потеряться,
	// пока main ещё инициализирует зависимости и не слушает канал.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Log.Info("server runs", zap.String("address", a.diContainer.cfg.ServerAddress))

	// 2. Запуск сервера в горутине.
	// ListenAndServe блокирует — поэтому запускаем в горутине, а main ждёт сигнал.
	// http.ErrServerClosed — нормальное завершение (мы сами вызвали Shutdown), не ошибка.
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("mistake of server", zap.Error(err))
		}
	}()

	// 3. Ожидание сигнала.
	<-ctx.Done()
	logger.Log.Info("get signal, end...")

	// Паттерн "двойной Ctrl+C": снимаем custom handler.
	// Теперь второй Ctrl+C убьёт процесс мгновенно (дефолтное поведение ОС).

	stop()

	// 4. Graceful shutdown HTTP-сервера.
	// Таймаут 15 секунд. Используем context.Background(), а не ctx — тот уже отменён.
	//
	// Что делает server.Shutdown внутри:
	//   1) Закрывает listeners — новые TCP-соединения невозможны
	//   2) Закрывает idle connections (keep-alive без активных запросов)
	//   3) Ждёт активные connections — пока handler вернёт ответ
	//   4) Если контекст истёк — возвращает ошибку, НО handlers продолжают работать в фоне

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("ошибка при приостановки сервера", zap.Error(err))
	}

	logger.Log.Info("сервер остановлен")

	// 5. Закрытие всех ресурсов через глобальный closer (LIFO).
	// Отдельный контекст с таймаутом 10 секунд — свой бюджет для ресурсов.
	// Суммарно: 15с (сервер) + 10с (ресурсы) = 25с из 30с Kubernetes grace period.
	closerCtx, closerCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closerCancel()

	if err := closer.CloseAll(closerCtx); err != nil {
		logger.Log.Error("ошибка при закрытии ресурсов", zap.Error(err))
	}

	return nil
}
