// Пакет main предоставляет точку входа для сервиса сокращения URL.
//
// Основные функции:
//
//   - Инициализация и конфигурация всех компонентов системы
//   - Запуск HTTP сервера
//   - Настройка логгера
//
// Конфигурация осуществляется через:
//   - Флаги командной строки
//   - Переменные окружения
//
// Примеры запуска:
//
//	# С конфигурацией по умолчанию
//	./shortener
//
//	# С указанием порта
//	./shortener -a :8080
//
//	# С подключением к PostgreSQL
//	./shortener -d postgres://user:pass@localhost:5432/db
//
// Доступные флаги:
//
//	-a, --address    Адрес HTTP сервера (по умолчанию ":8080")
//	-b, --base-url   Базовый URL для коротких ссылок
//	-d, --database   DSN для подключения к PostgreSQL
//	-f, --file       Путь к файлу хранилища (для in-memory режима)
//	-l, --log-level  Уровень логирования (debug, info, warn, error)
package main

import (
	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/server"
)

func main() {

	cfg := config.Load()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	// Запуск сервера с использованием конфигурации
	logger.Log.Info("Starting server", zap.String("address", cfg.HTTPServerAddr), zap.String("BaseURL", cfg.BaseURL))
	//log.Printf("Starting server on %s with base URL %s", cfg.HTTPServerAddr, cfg.BaseURL)
	server.StartServer(logger.Log, cfg)

}
