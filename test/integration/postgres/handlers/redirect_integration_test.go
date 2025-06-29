package integration

import (
	"testing"

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
	testhandlers.TestRedirect(t, testPG, client)
}
