package integration

import (
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/test/testhandlers"
)

// TestDelUserUrls_Postgres тестирует обработчик удаления пользовательских URL с использованием PostgreSQL хранилища.
//
// Проверяет:
//   - Корректность пометки URL как удалённых в PostgreSQL
//   - Работу асинхронного удаления через фоновые процессы
//   - Изоляцию данных между разными пользователями
//   - Соответствие поведения спецификации API при использовании PostgreSQL
//
// Особенности:
//   - Требует доступ к PostgreSQL через переменную TEST_DB_DSN
//   - Автоматически пропускается, если TEST_DB_DSN не задана
//   - Использует общие тестовые сценарии из testhandlers.TestDelUserUrls
//   - Проверяет специфичное для PostgreSQL поведение транзакций
//
// Пример использования:
//
//	go test -v -run TestDelUserUrls_Postgres
//
// Зависимости:
//   - Запущенный экземпляр PostgreSQL (обычно в Docker-контейнере)
//   - Применённые миграции базы данных
//   - Пакет testhandlers с базовыми тестами обработчиков
func TestDelUserUrls_Postgres(t *testing.T) {

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Skip("TEST_DB_DSN не установлен")
	}
	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		t.Fatal(err)
	}

	testhandlers.TestDelUserUrls(t, pg)
}
