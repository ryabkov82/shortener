package server

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/batch"
	"github.com/ryabkov82/shortener/internal/app/handlers/deluserurls"
	"github.com/ryabkov82/shortener/internal/app/handlers/ping"
	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/shortenapi"
	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	"github.com/ryabkov82/shortener/internal/app/handlers/userurls"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/storage/inmemory"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"

	"github.com/go-chi/chi/v5"

	"github.com/ryabkov82/shortener/internal/app/pprof"
)

// StartServer запускает HTTP-сервер.
func StartServer(log *zap.Logger, cfg *config.Config) {

	pprof.StartPProf(log, cfg.ConfigPProf)

	srv := &service.Service{}

	if cfg.DBConnect != "" {
		pg, err := postgres.NewPostgresStorage(cfg.DBConnect)

		if err != nil {
			panic(err)
		}
		srv = service.NewService(pg)
		log.Info("Storage postgres")
	} else {
		st, err := inmemory.NewInMemoryStorage(cfg.FileStorage)

		if err != nil {
			panic(err)
		}
		// загружаем сохраненные данные из файла..
		if err := st.Load(cfg.FileStorage); err != nil {
			panic(err)
		}
		srv = service.NewService(st)
		log.Info("Storage inmemory", zap.String("FileStorage", cfg.FileStorage))
	}

	router := chi.NewRouter()
	router.Use(mwlogger.RequestLogging(log))
	router.Use(mwgzip.Gzip)

	/*
		// Группа с автоматической аутентификацией
		router.Group(func(router chi.Router) {
			router.Use(auth.JWTAutoIssue([]byte(cfg.JwtKey)))

			router.Post("/", shorturl.GetHandler(srv, cfg.BaseURL, log))
			router.Get("/{id}", redirect.GetHandler(srv, log))

			router.Post("/api/shorten", shortenapi.GetHandler(srv, cfg.BaseURL, log))

			router.Get("/ping", ping.GetHandler(srv, log))
			router.Post("/api/shorten/batch", batch.GetHandler(srv, cfg.BaseURL, log))
		})

		// Группа со строгой аутентификацией
		router.Group(func(router chi.Router) {
			router.Use(auth.StrictJWTAutoIssue([]byte(cfg.JwtKey)))
			router.Get("/api/user/urls", userurls.GetHandler(srv, cfg.BaseURL, log))
		})
	*/

	router.Use(auth.JWTAutoIssue([]byte(cfg.JwtKey)))

	router.Post("/", shorturl.GetHandler(srv, cfg.BaseURL, log))
	router.Get("/{id}", redirect.GetHandler(srv, log))

	router.Post("/api/shorten", shortenapi.GetHandler(srv, cfg.BaseURL, log))

	router.Get("/ping", ping.GetHandler(srv, log))
	router.Post("/api/shorten/batch", batch.GetHandler(srv, cfg.BaseURL, log))
	router.Get("/api/user/urls", userurls.GetHandler(srv, cfg.BaseURL, log))
	router.Delete("/api/user/urls", deluserurls.GetHandler(srv, cfg.BaseURL, log))

	log.Info("Server started", zap.String("address", cfg.HTTPServerAddr))

	// Запуск HTTP-сервера в отдельной горутине

	server := &http.Server{
		Addr:    cfg.HTTPServerAddr,
		Handler: router, // Ваш роутер
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to serve server", zap.Error(err))
		}
	}()

	// Обработка сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Остановка HTTP-сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Info("HTTP server shutdown error", zap.Error(err))
	}

	// корректное завершение работы воркеров сервиса
	srv.GracefulStop(5 * time.Second)

	log.Info("Server shutdown complete")

}
