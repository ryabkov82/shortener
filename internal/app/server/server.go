package server

import (
	"log"
	"net/http"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

// StartServer запускает HTTP-сервер.
func StartServer(cfg *config.Config) {

	storage := storage.NewInMemoryStorage()

	service := service.NewService(storage)

	router := chi.NewRouter()
	router.Post("/", shorturl.GetHandler(service, cfg.BaseURL))
	router.Get("/{id}", redirect.GetHandler(service))

	log.Printf("Server started at %s", cfg.HTTPServerAddr)

	log.Fatal(http.ListenAndServe(cfg.HTTPServerAddr, router))

}
