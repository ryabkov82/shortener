package integration

import (
	"testing"

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
	testhandlers.TestDelUserUrls(t, serv, client)
}
