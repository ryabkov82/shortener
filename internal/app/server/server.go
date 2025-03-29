package server

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/ping"
	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/shortenapi"
	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service"
	storage "github.com/ryabkov82/shortener/internal/app/storage/inmemory"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"

	"github.com/go-chi/chi/v5"
)

// StartServer запускает HTTP-сервер.
func StartServer(log *zap.Logger, cfg *config.Config) {

	pg, err := postgres.NewPostgresStorage(cfg.DBConnect)

	if err != nil {
		panic(err)
	}

	st, err := storage.NewInMemoryStorage(cfg.FileStorage)

	if err != nil {
		panic(err)
	}

	// загружаем сохраненные данные из файла..
	if err := st.Load(cfg.FileStorage); err != nil {
		panic(err)
	}

	service := service.NewService(st)

	router := chi.NewRouter()
	router.Use(mwlogger.RequestLogging(log))
	router.Use(mwgzip.Gzip)

	router.Post("/", shorturl.GetHandler(service, cfg.BaseURL, log))
	router.Get("/{id}", redirect.GetHandler(service, log))

	router.Post("/api/shorten", shortenapi.GetHandler(service, cfg.BaseURL, log))

	router.Get("/ping", ping.GetHandler(pg, log))

	log.Info("Server started", zap.String("address", cfg.HTTPServerAddr))

	if err := http.ListenAndServe(cfg.HTTPServerAddr, router); err != nil {
		log.Error("failed to serve server", zap.Error(err))
	}

}
