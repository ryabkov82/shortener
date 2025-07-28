# Makefile для тестирования shortener

# Конфигурация
BENCH_DIR = ./benchmark
TEST_DIR = ./...
DOCKER_COMPOSE = docker-compose -f docker-compose.test.yml

# Определяем переменные для версии
VERSION := "1.0.0"
ifeq ($(OS),Windows_NT)
	BUILD_DATE := $(shell powershell -command "Get-Date -Format 'yyyy-MM-ddTHH:mm:sszzz'")
else
	BUILD_DATE := $(shell date +'%Y-%m-%dT%H:%M:%S%z')
endif

COMMIT_HASH := $(shell git rev-parse --short HEAD)

ifeq ($(OS),Windows_NT)
    SET_ENV = set TEST_DB_DSN=host=localhost port=5433 user=test password=test dbname=test sslmode=disable
else
    SET_ENV = TEST_DB_DSN="postgres://test:test@localhost:5433/test?sslmode=disable"
endif

.PHONY: build
build:
	@echo "Building shortener..."
	go build -v -ldflags "\
		-X 'main.buildVersion=${VERSION}' \
		-X 'main.buildDate=${BUILD_DATE}' \
		-X 'main.buildCommit=${COMMIT_HASH}'" \
		-o bin/shortener cmd/shortener/main.go
	@echo "Build complete"

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
	go test -bench=InMemory -benchmem -memprofile=profiles/optimized.pprof -benchtime=5s $(BENCH_DIR)

# Бенчмарки с Postgres
bench-pg: start-db
	$(SET_ENV) && go test -bench=Postgres -benchmem -memprofile=profiles/result.pprof -benchtime=5s $(BENCH_DIR)

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
generate:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/shortener.proto