package main

import (
	"net/http"

	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	"github.com/ryabkov82/shortener/internal/app/storage"
)

func main() {

	storage := storage.New()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", shorturl.GetHandler(storage))
	mux.HandleFunc("GET /{id}", redirect.GetHandler(storage))
	http.ListenAndServe("localhost:8080", mux)
}
