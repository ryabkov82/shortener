package integration

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/test/testconfig"
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

	// 2. Экспорт DSN для других тестов
	os.Setenv("TEST_DB_DSN", dsn)

	// 3. Запуск всех тестов
	code := m.Run()

	// 4. Очистка
	if err := container.Terminate(ctx); err != nil {
		log.Printf("Failed to terminate container: %v", err)
	}
	os.Unsetenv("TEST_DB_DSN")

	os.Exit(code)
}
