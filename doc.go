// Package shortener представляет сервис сокращения URL-ссылок.
//
// Конфигурация осуществляется через:
//   - Флаги командной строки
//   - Переменные окружения
//
// Примеры запуска:
//
//	# С конфигурацией по умолчанию
//	./shortener
//
//	# С указанием порта
//	./shortener -a :8080
//
//	# С подключением к PostgreSQL
//	./shortener -d postgres://user:pass@localhost:5432/db
//
// Доступные флаги:
//
//	-a, --address    Адрес HTTP сервера (по умолчанию ":8080")
//	-b, --base-url   Базовый URL для коротких ссылок
//	-d, --database   DSN для подключения к PostgreSQL
//	-f, --file       Путь к файлу хранилища (для in-memory режима)
//	-l, --log-level  Уровень логирования (debug, info, warn, error)
package shortener
