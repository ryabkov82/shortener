// Пакет main предоставляет точку входа для сервиса сокращения URL.
//
// Основные функции:
//
//   - Инициализация конфигурации системы
//   - Настройка логгера
//   - Запуск HTTP сервера
package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/server"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {

	printBuildInfo()

	cfg := config.Load()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	// Запуск сервера с использованием конфигурации
	logger.Log.Info("Starting server", zap.String("address", cfg.HTTPServerAddr), zap.String("BaseURL", cfg.BaseURL))
	// log.Printf("Starting server on %s with base URL %s", cfg.HTTPServerAddr, cfg.BaseURL)
	server.StartServer(logger.Log, cfg)

}

func printBuildInfo() {
	// Set default value "N/A" if variables are empty
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)
}
