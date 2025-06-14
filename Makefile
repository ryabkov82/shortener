# Makefile для тестирования shortener

# Конфигурация
BENCH_DIR = ./benchmark
TEST_DIR = ./...
DOCKER_COMPOSE = docker-compose -f docker-compose.test.yml

ifeq ($(OS),Windows_NT)
    SET_ENV = set TEST_DB_DSN=host=localhost port=5433 user=test password=test dbname=test sslmode=disable
else
    SET_ENV = TEST_DB_DSN="postgres://test:test@localhost:5433/test?sslmode=disable"
endif

# Цели по умолчанию
.PHONY: default
default: test

## Тестирование
.PHONY: test test-race test-cover

# Обычные тесты
test:
	go test -v $(TEST_DIR)

# Тесты с детектором гонок
test-race:
	go test -race -v $(TEST_DIR)

# Тесты с покрытием
test-cover:
	go test -coverprofile=coverage.out -covermode=atomic $(TEST_DIR)
	go tool cover -html=coverage.out -o coverage.html

## Бенчмарки
.PHONY: bench bench-pg bench-full

# Базовые бенчмарки (in-memory)
bench:
	go test -bench=InMemory -benchmem -memprofile=profiles/base_1.pprof -benchtime=5s $(BENCH_DIR)

# Бенчмарки с Postgres
bench-pg: start-db
	$(SET_ENV) && go test -bench=Postgres -benchmem -memprofile=profiles/base_pg.pprof -benchtime=5s $(BENCH_DIR)

# Полный набор бенчмарков
bench-full: bench bench-pg

## Вспомогательные команды
.PHONY: start-db stop-db clean

# Запуск тестовой БД
start-db:
	$(DOCKER_COMPOSE) up -d postgres
	timeout /t 5 /nobreak > NUL

# Остановка тестовой БД
stop-db:
	$(DOCKER_COMPOSE) down

# Очистка
clean:
	rm -f *.pprof coverage.*
	$(DOCKER_COMPOSE) down -v
