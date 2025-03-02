package main

import (
	"net/http"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	"github.com/ryabkov82/shortener/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

func main() {

	storage := storage.New()

	cfg := config.Load()

	router := chi.NewRouter()
	router.Post("/", shorturl.GetHandler(storage, cfg.BaseURL))
	router.Get("/{id}", redirect.GetHandler(storage))
	http.ListenAndServe(cfg.HTTPServerAddr, router)
}
