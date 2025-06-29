package integration

import (
	"testing"

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
	testhandlers.TestBatch(t, client)
}
