package integration

import (
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/test/testhandlers"
)

// TestUserUrls_Postgres тестирует обработчик получения списка URL пользователя с использованием PostgreSQL.
//
// Проверяет:
//   - Корректность получения списка URL из PostgreSQL
//   - Работу авторизации через JWT при использовании PostgreSQL
//   - Возвращаемые статусы:
//   - 200 OK при успешном запросе
//   - 204 No Content при отсутствии URL
//   - 401 Unauthorized без авторизации
//   - Формат JSON-ответа
//   - Изоляцию данных между пользователями
//
// Особенности:
//   - Требует переменную окружения TEST_DB_DSN
//   - Автоматически пропускается при отсутствии подключения к БД
//   - Использует тестовые сценарии из testhandlers.TestUserUrls
//   - Проверяет специфичное для PostgreSQL поведение:
//   - Работу с индексами
//   - Эффективность запросов
//   - Транзакционность
//
// Пример использования:
//
//	go test -v -run TestUserUrls_Postgres
//
// Зависимости:
//   - Запущенный сервер PostgreSQL
//   - Применённые миграции БД
//   - Пакет testhandlers с базовыми тестами
func TestUserUrls_Postgres(t *testing.T) {

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Skip("TEST_DB_DSN не установлен")
	}
	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		t.Fatal(err)
	}

	testhandlers.TestUserUrls(t, pg)
}
