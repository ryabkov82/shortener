// internal/server/run.go
package server

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/pprof"
	grpcserver "github.com/ryabkov82/shortener/internal/app/server/grpc"
	httpserver "github.com/ryabkov82/shortener/internal/app/server/http"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/storage/inmemory"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

func StartServers(log *zap.Logger, cfg *config.Config) {

	pprof.StartPProf(log, cfg.ConfigPProf)

	// 1. Инициализация хранилища и сервиса
	storage, err := initStorage(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize storage", zap.Error(err))
	}

	appService := service.NewService(storage)

	// 2. Запуск серверов
	httpServer := httpserver.StartHTTPServer(log, cfg, appService)
	grpcServer := grpcserver.StartGRPCServer(log, cfg, appService)

	// 3. Graceful shutdown
	waitForShutdown(log, httpServer, grpcServer, appService)
}

// Вспомогательные функции
func initStorage(cfg *config.Config, log *zap.Logger) (service.Repository, error) {
	if cfg.DBConnect != "" {
		pg, err := postgres.NewPostgresStorage(cfg.DBConnect)
		if err != nil {
			return nil, err
		}
		log.Info("Using PostgreSQL storage")
		return pg, nil
	}

	mem, err := inmemory.NewInMemoryStorage(cfg.FileStorage)
	if err != nil {
		return nil, err
	}

	if err := mem.Load(cfg.FileStorage); err != nil {
		return nil, err
	}

	log.Info("Using in-memory storage", zap.String("file", cfg.FileStorage))
	return mem, nil
}

func waitForShutdown(
	log *zap.Logger,
	httpServer *http.Server,
	grpcServer *grpc.Server,
	service *service.Service,
) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Остановка HTTP сервера
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP server shutdown error", zap.Error(err))
	}

	// Остановка gRPC сервера
	grpcServer.GracefulStop()

	// Завершение работы сервиса
	service.GracefulStop(5 * time.Second)
	if err := service.Close(); err != nil {
		log.Error("Storage close error", zap.Error(err))
	}

	log.Info("Servers stopped gracefully")
}
