package integration

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/handlers/batch"
	"github.com/ryabkov82/shortener/internal/app/handlers/deluserurls"
	"github.com/ryabkov82/shortener/internal/app/handlers/ping"
	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/shortenapi"
	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	"github.com/ryabkov82/shortener/internal/app/handlers/userurls"

	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
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

// TestMain является точкой входа для выполнения интеграционных тестов с PostgreSQL.
//
// Основные функции:
//   - Запускает контейнер PostgreSQL с использованием testcontainers
//   - Устанавливает переменную окружения TEST_DB_DSN с DSN строкой подключения
//   - Обеспечивает корректное завершение работы контейнера после тестов
//   - Запускает все тесты пакета
//
// Этапы работы:
//  1. Создает контейнер PostgreSQL с конфигурацией по умолчанию
//  2. Экспортирует DSN строку подключения в переменную TEST_DB_DSN
//  3. Запускает все тесты пакета
//  4. Останавливает контейнер после выполнения тестов
//
// Переменные окружения:
//   - TEST_DB_DSN: строка подключения к тестовой БД (доступна во время выполнения тестов)
//
// Пример использования в тестах:
//
//	func TestSomething(t *testing.T) {
//	    dsn := os.Getenv("TEST_DB_DSN")
//	    // использование dsn для подключения к тестовой БД
//	}
//
// Примечания:
//   - Контейнер автоматически останавливается после выполнения всех тестов
//   - Для корректной работы требуется Docker
//   - Использует конфигурацию по умолчанию из testconfig.DefaultPGConfig()
func TestMain(m *testing.M) {

	ctx := context.Background()
	// 1. Запуск контейнера

	container, dsn, err := testconfig.StartPGContainer(ctx, testconfig.DefaultPGConfig())
	if err != nil {
		panic(err)
	}

	// 2. Подготовка тестового окружения
	if err := logger.Initialize("debug"); err != nil {
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
