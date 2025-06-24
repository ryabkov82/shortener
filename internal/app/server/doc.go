// Package server реализует основной HTTP-сервер для сервиса сокращения URL.
//
// Сервер предоставляет:
// - REST API для работы с короткими ссылками
// - Поддержку различных хранилищ (PostgreSQL, in-memory)
// - Middleware для аутентификации, логирования и сжатия
// - Graceful shutdown
//
// Основные компоненты:
//
//   - **Конфигурация**: Настройки через config.Config
//   - **Роутинг**: Реализован на базе chi.Router
//   - **Обработчики**:
//   - / - Создание короткой ссылки (POST)
//   - /{id} - Перенаправление (GET)
//   - /api/shorten - JSON API создания ссылки
//   - /api/shorten/batch - Пакетное создание
//   - /api/user/urls - Список ссылок пользователя
//   - /ping - Проверка доступности БД
//
// Пример запуска:
//
//	cfg := config.Load()
//	log := logger.Initialize(cfg.LogLevel)
//	server.StartServer(log, cfg)
//
// Особенности:
//   - Поддержка двух хранилищ (PostgreSQL и in-memory с персистентностью)
//   - JWT аутентификация
//   - GZIP сжатие
//   - Подробное логирование
//   - Профилирование через pprof
//   - Graceful shutdown при получении SIGINT/SIGTERM
package server
