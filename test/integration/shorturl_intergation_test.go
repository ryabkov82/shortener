package integration

import (
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/test/testhandlers"
)

// TestGetHandler_Postgres тестирует обработчик сокращения URL с использованием PostgreSQL в качестве хранилища.
//
// Проверяет:
//   - Корректность создания коротких ссылок в PostgreSQL
//   - Возвращаемые HTTP-статусы (201 Created, 400 Bad Request, 409 Conflict)
//   - Сохранение и чтение данных из PostgreSQL
//   - Формат возвращаемого ответа
//   - Обработку дубликатов URL
//
// Особенности:
//   - Требует переменную окружения TEST_DB_DSN с параметрами подключения к PostgreSQL
//   - Автоматически пропускается, если TEST_DB_DSN не задана
//   - Использует базовые тест-кейсы из testhandlers.TestShortenURL
//   - Проверяет специфичное для PostgreSQL поведение:
//   - Работу транзакций
//   - Блокировки при конкурентном доступе
//   - Целостность данных
//
// Пример использования:
//
//	go test -v -run TestGetHandler_Postgres
//
// Зависимости:
//   - Запущенный сервер PostgreSQL (рекомендуется версия 12+)
//   - Применённые миграции базы данных
//   - Пакет testhandlers с базовыми тестами
func TestGetHandler_Postgres(t *testing.T) {

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Skip("TEST_DB_DSN не установлен")
	}
	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		t.Fatal(err)
	}

	testhandlers.TestShortenURL(t, pg)
}
