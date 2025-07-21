package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/bootstrap"
	"go.uber.org/zap"
)

func main() {
	cfg, appLogger := bootstrap.Bootstrap()
	defer func() {
		if err := appLogger.Sync(); err != nil {
			log.Printf("Ошибка при закрытии логгера: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := bootstrap.RunMigrations(ctx, cfg, appLogger); err != nil {
		appLogger.Fatal(ctx, "Ошибка при выполнении миграций", zap.Error(err))
	}

	repos, err := bootstrap.InitializeRepositories(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal(ctx, "Ошибка инициализации репозиториев", zap.Error(err))
	}

	services := bootstrap.InitializeServices(ctx, cfg, repos, appLogger)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := StartGRPCServer(ctx, cfg, services, appLogger); err != nil {
			appLogger.Error(ctx, "Критическая ошибка gRPC сервера", zap.Error(err))
			cancel()
		}
	}()

	go func() {
		defer wg.Done()
		if err := StartHTTPServer(ctx, cfg, appLogger); err != nil {
			appLogger.Error(ctx, "Критическая ошибка HTTP сервера", zap.Error(err))
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	wg.Wait()
	appLogger.Info(ctx, "Серверы остановлены")
}
