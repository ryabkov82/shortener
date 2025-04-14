package server

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/batch"
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
)

// StartServer запускает HTTP-сервер.
func StartServer(log *zap.Logger, cfg *config.Config) {

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

	log.Info("Server started", zap.String("address", cfg.HTTPServerAddr))

	if err := http.ListenAndServe(cfg.HTTPServerAddr, router); err != nil {
		log.Error("failed to serve server", zap.Error(err))
	}

}
