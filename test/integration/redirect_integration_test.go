package integration

import (
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/test/testhandlers"
)

// TestRedirect_Postgres тестирует обработчик редиректов с использованием PostgreSQL в качестве хранилища.
//
// Проверяет:
//   - Корректность редиректов для URL, сохраненных в PostgreSQL
//   - Обработку несуществующих коротких ссылок
//   - Соответствие HTTP-статусов (307, 404)
//   - Работу с JWT-авторизацией при использовании PostgreSQL
//
// Особенности:
//   - Требует настроенного подключения к PostgreSQL через TEST_DB_DSN
//   - Автоматически пропускается, если переменная окружения не задана
//   - Переиспользует базовые тест-кейсы из testhandlers.TestRedirect
//   - Проверяет специфичное для PostgreSQL поведение (блокировки, транзакции)
//
// Пример запуска:
//
//	go test -v -run TestRedirect_Postgres
//
// Зависимости:
//   - Запущенный экземпляр PostgreSQL (обычно через testcontainers)
//   - Примененные миграции базы данных
//   - Пакет testhandlers с базовыми тестами обработчиков
func TestRedirect_Postgres(t *testing.T) {

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Skip("TEST_DB_DSN не установлен")
	}
	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		t.Fatal(err)
	}

	testhandlers.TestRedirect(t, pg)
}
