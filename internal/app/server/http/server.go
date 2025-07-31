package httpserver

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/batch"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/deluserurls"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/ping"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/shortenapi"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/shorturl"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/stats"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/userurls"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/http/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/trustednet"
	"github.com/ryabkov82/shortener/internal/app/service"

	"github.com/go-chi/chi/v5"
)

// StartHTTPServer запускает HTTP-сервер.
func StartHTTPServer(log *zap.Logger, cfg *config.Config, srv *service.Service) *http.Server {

	log.Info("Starting http server", zap.String("address", cfg.HTTPServerAddr), zap.String("BaseURL", cfg.BaseURL))

	router := setupRouter(log, cfg, srv)

	server := &http.Server{
		Addr:    cfg.HTTPServerAddr,
		Handler: router,
	}

	go runServer(log, server, cfg)
	return server

}

// Приватные вспомогательные функции
func setupRouter(log *zap.Logger, cfg *config.Config, srv *service.Service) http.Handler {

	router := chi.NewRouter()
	// Настройка middleware и роутов
	router.Use(mwlogger.RequestLogging(log))
	router.Use(mwgzip.Gzip)

	router.Use(auth.JWTAutoIssue([]byte(cfg.JwtKey)))

	router.Post("/", shorturl.GetHandler(srv, cfg.BaseURL, log))
	router.Get("/{id}", redirect.GetHandler(srv, log))

	router.Post("/api/shorten", shortenapi.GetHandler(srv, cfg.BaseURL, log))

	router.Get("/ping", ping.GetHandler(srv, log))
	router.Post("/api/shorten/batch", batch.GetHandler(srv, cfg.BaseURL, log))
	router.Get("/api/user/urls", userurls.GetHandler(srv, cfg.BaseURL, log))
	router.Delete("/api/user/urls", deluserurls.GetHandler(srv, cfg.BaseURL, log))

	router.Group(func(router chi.Router) {
		router.Use(trustednet.CheckTrustedSubnet(cfg.TrustedSubnet))
		router.Get("/api/internal/stats", stats.GetHandler(srv, log))
	})

	return router
}

func runServer(log *zap.Logger, server *http.Server, cfg *config.Config) {

	if cfg.EnableHTTPS {
		// Запуск сервера с HTTPS
		go func() {
			if err := server.ListenAndServeTLS(cfg.SSLCertFile, cfg.SSLKeyFile); err != nil && err != http.ErrServerClosed {
				log.Error("failed to serve server", zap.Error(err))
			}
		}()
	} else {
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error("failed to serve server", zap.Error(err))
			}
		}()
	}

	log.Info("Server started", zap.String("address", cfg.HTTPServerAddr))

}
