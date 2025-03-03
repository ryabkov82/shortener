package server

import (
	"log"
	"net/http"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	"github.com/ryabkov82/shortener/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

// StartServer запускает HTTP-сервер.
func StartServer() {

	storage := storage.NewInMemoryStorage()

	cfg := config.Load()

	router := chi.NewRouter()
	router.Post("/", shorturl.GetHandler(storage, cfg.BaseURL))
	router.Get("/{id}", redirect.GetHandler(storage))

	log.Printf("Server started at %s", cfg.HTTPServerAddr)

	log.Fatal(http.ListenAndServe(cfg.HTTPServerAddr, router))

}
