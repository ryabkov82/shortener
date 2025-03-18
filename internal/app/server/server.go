package server

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/shortenapi"
	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	logger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/service"
	storage "github.com/ryabkov82/shortener/internal/app/storage/inmemory"

	"github.com/go-chi/chi/v5"
)

// StartServer запускает HTTP-сервер.
func StartServer(log *zap.Logger, cfg *config.Config) {

	storage := storage.NewInMemoryStorage()

	service := service.NewService(storage)

	router := chi.NewRouter()
	router.Use(logger.RequestLogging(log))

	router.Post("/", shorturl.GetHandler(service, cfg.BaseURL, log))
	router.Get("/{id}", redirect.GetHandler(service, log))

	router.Post("/api/shorten", shortenapi.GetHandler(service, cfg.BaseURL, log))

	log.Info("Server started", zap.String("address", cfg.HTTPServerAddr))

	if err := http.ListenAndServe(cfg.HTTPServerAddr, router); err != nil {
		log.Error("failed to serve server", zap.Error(err))
	}

}
