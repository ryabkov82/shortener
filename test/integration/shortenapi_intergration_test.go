package integration

import (
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/test/testhandlers"
)

// TestShortenAPI_Postgres тестирует JSON API сокращения URL с использованием PostgreSQL хранилища.
//
// Проверяет:
//   - Корректность создания сокращённых ссылок в PostgreSQL
//   - Обработку дубликатов URL (статус 409 Conflict)
//   - Валидацию входных данных (статус 400 Bad Request)
//   - Формат JSON ответа (поле "result")
//   - Сохранение и чтение данных через PostgreSQL
//
// Особенности:
//   - Требует переменную окружения TEST_DB_DSN с параметрами подключения
//   - Автоматически пропускается при отсутствии TEST_DB_DSN
//   - Использует тестовые сценарии из testhandlers.TestShortenAPI
//   - Проверяет специфичное для PostgreSQL поведение:
//   - Конкурентные запросы
//   - Транзакционность операций
//   - Целостность данных
//
// Пример использования:
//
//	go test -v -run TestShortenAPI_Postgres
//
// Зависимости:
//   - Запущенный сервер PostgreSQL (рекомендуется версия 13+)
//   - Применённые миграции базы данных
//   - Пакет testhandlers с базовыми тестами API
func TestShortenAPI_Postgres(t *testing.T) {

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Skip("TEST_DB_DSN не установлен")
	}
	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		t.Fatal(err)
	}

	testhandlers.TestShortenAPI(t, pg)
}
