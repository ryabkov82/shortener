package integration

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/handlers/http/batch"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/deluserurls"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/ping"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/shortenapi"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/shorturl"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/userurls"

	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/http/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/test/testconfig"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

var (
	client *resty.Client
	serv   *service.Service
	testPG service.Repository
)

// TestMain является точкой входа для интеграционных тестов и настраивает тестовое окружение.
//
// Функция выполняет:
//  1. Запуск контейнера PostgreSQL через testcontainers-go
//  2. Инициализацию зависимостей:
//     - Логгер с уровнем debug
//     - Хранилище PostgreSQL
//     - Сервисный слой
//     - HTTP сервер с роутингом и middleware
//  3. Запуск всех тестов
//  4. Корректную очистку ресурсов
//
// Настройка тестового сервера:
//   - Глобальные middleware:
//   - Логирование запросов
//   - Поддержка gzip
//   - Группа с авторизацией:
//   - JWT авторизация с автоматической выдачей
//   - Основные обработчики (сокращение URL, редиректы)
//   - Группа со строгой авторизацией:
//   - Строгая JWT проверка
//   - Обработчики работы с пользовательскими URL
//
// Особенности:
//   - Для каждого теста создается чистое окружение
//   - HTTP клиент сброшен (отключен cookie jar)
//   - Гарантируется остановка контейнера после тестов
//   - Логирование инициализируется с уровнем debug
//
// Пример переменных окружения:
//
//	TEST_DB_DSN="postgres://user:pass@localhost:5432/testdb?sslmode=disable"
//
// Примечания:
//   - Требует запущенного Docker демона
//   - Использует образ PostgreSQL из testconfig.DefaultPGConfig()
//   - Для изоляции тестов cookie jar отключен явно
func TestMain(m *testing.M) {

	ctx := context.Background()
	// 1. Запуск контейнера

	container, dsn, err := testconfig.StartPGContainer(ctx, testconfig.DefaultPGConfig())
	if err != nil {
		panic(err)
	}

	// 2. Подготовка тестового окружения
	if err = logger.Initialize("debug"); err != nil {
		panic(err)
	}

	testPG, err = postgres.NewPostgresStorage(dsn)
	if err != nil {
		panic(err)
	}

	serv = service.NewService(testPG)

	baseURL := "http://localhost:8080/"

	tc := testutils.NewTestClient(func(r chi.Router) {
		r.Use(mwlogger.RequestLogging(logger.Log))
		r.Use(mwgzip.Gzip)

		r.Group(func(r chi.Router) {
			r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

			r.Post("/", shorturl.GetHandler(serv, baseURL, logger.Log))
			r.Get("/{id}", redirect.GetHandler(serv, logger.Log))
			r.Post("/api/shorten", shortenapi.GetHandler(serv, baseURL, logger.Log))
			r.Get("/ping", ping.GetHandler(serv, logger.Log))
			r.Post("/api/shorten/batch", batch.GetHandler(serv, baseURL, logger.Log))
		})

		// Группа со строгой аутентификацией
		r.Group(func(r chi.Router) {
			r.Use(auth.StrictJWTAutoIssue(testutils.TestSecretKey))
			r.Get("/api/user/urls", userurls.GetHandler(serv, baseURL, logger.Log))
			r.Delete("/api/user/urls", deluserurls.GetHandler(serv, baseURL, logger.Log))
		})
	})

	client = tc.Client
	// Отключение cookie jar (не сохранять cookies)
	client.SetCookieJar(nil)

	// 3. Запуск всех тестов
	code := m.Run()

	// 4. Очистка
	if err := container.Terminate(ctx); err != nil {
		log.Printf("Failed to terminate container: %v", err)
	}
	tc.Close()

	os.Exit(code)
}
