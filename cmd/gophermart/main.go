package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/handler"
	"github.com/KV2013/gophermart-loyalty/internal/logger"
	"github.com/KV2013/gophermart-loyalty/internal/repository"
	"github.com/KV2013/gophermart-loyalty/internal/router"
	"github.com/KV2013/gophermart-loyalty/internal/service"
	"github.com/KV2013/gophermart-loyalty/internal/worker"
	"go.uber.org/zap"
)

func main() {
	config, cfgErr := config.NewConfig()
	if cfgErr != nil {
		log.Fatal("Ошибка при сборке конфига")
	}
	Logger, loggerErr := logger.New(config.LogLevel)
	if loggerErr != nil {
		log.Fatal("Ошибка при сборке логгера")
	}
	defer Logger.Sync()

	repository, repoErr := repository.New(config, Logger)
	if repoErr != nil {
		Logger.Fatal("Ошибка при сборке репозитория", zap.Error(repoErr))
	}
	service := service.New(repository.User, repository.Order, repository.Balance, Logger)
	handler := handler.New(service.Auth, service.Balance, service.Order, config, Logger)

	mux := router.Init(handler, service.Auth, Logger, config)
	srv := &http.Server{
		Addr:         config.ServerAddress,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	accrualWorker := worker.New(repository.Order, repository.Balance, Logger, config.AccrualSystemAddress)
	go accrualWorker.Start(ctx)
	go func() {
		Logger.Info(
			"Сервер запущен",
			zap.String("serverAddress", config.ServerAddress),
			zap.String("logLevel", config.LogLevel),
		)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			Logger.Fatal("Ошибка при запуске сервера", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	Logger.Info("Получен сигнал завершения. Начинаем graceful shutdown...")
	shutdownCtx, cancelShutdownCtx := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdownCtx()

	err := srv.Shutdown(shutdownCtx)

	if err != nil {
		Logger.Fatal("Ошибка при завершении сервера", zap.Error(err))
		os.Exit(1)
	}
	Logger.Info("Сервер успешно завершил работу")
}
