package main

import (
	"log"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/server"
)

func main() {

	cfg := config.Load()

	// Запуск сервера с использованием конфигурации
	log.Printf("Starting server on %s with base URL %s", cfg.HTTPServerAddr, cfg.BaseURL)
	server.StartServer(cfg)

}
