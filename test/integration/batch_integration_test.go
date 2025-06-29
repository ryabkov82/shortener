package integration

import (
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/test/testhandlers"
)

// TestBatch_Postgres тестирует обработчик пакетного создания сокращённых URL с использованием PostgreSQL хранилища.
//
// Проверяет:
//   - Корректность работы обработчика с PostgreSQL в качестве бэкенда
//   - Сохранение и извлечение пакета URL из БД
//   - Целостность данных при пакетных операциях
//
// Особенности:
//   - Использует реальную PostgreSQL БД из тестового окружения
//   - Требует установленной переменной окружения TEST_DB_DSN
//   - Пропускает тест если TEST_DB_DSN не установлен
//   - Переиспользует общие тестовые сценарии из testhandlers.TestBatch
//
// Пример переменной окружения:
//
//	TEST_DB_DSN="postgres://user:pass@localhost:5432/db?sslmode=disable"
//
// Зависимости:
//   - Запущенный контейнер PostgreSQL (обычно настраивается в TestMain)
//   - Пакет testhandlers с базовыми тестами
//
// Использование:
//
//	go test -v -run TestBatch_Postgres
func TestBatch_Postgres(t *testing.T) {

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Skip("TEST_DB_DSN не установлен")
	}
	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		t.Fatal(err)
	}

	testhandlers.TestBatch(t, pg)
}
