// Package testconfig предоставляет утилиты для настройки тестового окружения с использованием testcontainers.
//
// Основные возможности:
//   - Запуск контейнера PostgreSQL для интеграционных тестов
//   - Автоматическая конфигурация подключения к БД
//   - Параллельно-безопасная инициализация (используется sync.Once)
//   - Логирование работы контейнера в реальном времени
//
// Пример использования:
//
//	ctx := context.Background()
//	cfg := DefaultPGConfig()
//	container, dsn, err := StartPGContainer(ctx, cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer container.Terminate(ctx)
//
//	db, err := sql.Open("postgres", dsn)
//	// ... работа с БД
//
// Пакет используется для интеграционных тестов, требующих изолированного экземпляра БД.
package testconfig

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pgContainer testcontainers.Container
	pgDSN       string
	pgOnce      sync.Once
)

// PGConfig содержит конфигурационные параметры для настройки контейнера PostgreSQL.
//
// Поля структуры:
//   - Image:    Docker-образ PostgreSQL (например, "postgres:13-alpine")
//   - User:     Имя пользователя для подключения к БД
//   - Password: Пароль пользователя БД
//   - DBName:   Название создаваемой базы данных
//   - Port:     Порт для подключения к PostgreSQL (формат "5432")
//
// Пример использования:
//
//	cfg := PGConfig{
//	    Image:    "postgres:13-alpine",
//	    User:     "testuser",
//	    Password: "testpass",
//	    DBName:   "testdb",
//	    Port:     "5432",
//	}
//
// Примечания:
//   - Все поля обязательные
//   - Для тестов рекомендуется использовать образы с alpine
//   - Порт должен соответствовать порту, используемому в выбранном образе
type PGConfig struct {
	Image    string
	User     string
	Password string
	DBName   string
	Port     string
}

// DefaultPGConfig возвращает конфигурацию PostgreSQL со значениями по умолчанию
// Используется образ postgres:13-alpine с пользователем test, паролем test и БД test на порту 5432
func DefaultPGConfig() PGConfig {
	return PGConfig{
		Image:    "postgres:13-alpine",
		User:     "test",
		Password: "test",
		DBName:   "test",
		Port:     "5432",
	}
}

// StartPGContainer запускает контейнер PostgreSQL с заданной конфигурацией.
// Использует sync.Once для гарантии однократного запуска контейнера.
// Возвращает:
//   - testcontainers.Container: интерфейс для управления контейнером
//   - string: DSN строку для подключения к БД
//   - error: ошибка, если возникла при запуске контейнера
//
// Контейнер будет автоматически остановлен при завершении работы приложения.
func StartPGContainer(ctx context.Context, cfg PGConfig) (testcontainers.Container, string, error) {
	var startErr error
	pgOnce.Do(func() {

		// Конвертируем строку в nat.Port
		pgPort := nat.Port(cfg.Port + "/tcp") // Например: "5432" -> "5432/tcp"
		req := testcontainers.ContainerRequest{
			Image:        cfg.Image,
			ExposedPorts: []string{string(pgPort)},
			Env: map[string]string{
				"POSTGRES_USER":     cfg.User,
				"POSTGRES_PASSWORD": cfg.Password,
				"POSTGRES_DB":       cfg.DBName,
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready"),
				wait.ForListeningPort(pgPort),
			).WithDeadline(1 * time.Minute),
		}

		logger := log.New(os.Stdout, "[POSTGRES] ", log.LstdFlags)

		pgContainer, startErr = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
			Logger:           logger,
		})
		if startErr != nil {
			return
		}

		// Получаем логи в реальном времени
		go func() {
			reader, err := pgContainer.Logs(ctx)
			if err != nil {
				logger.Println("Failed to get logs:", err)
				return
			}
			defer reader.Close()

			buf := make([]byte, 1024)
			for {
				n, err := reader.Read(buf)
				if err != nil {
					return
				}
				logger.Print(string(buf[:n]))
			}
		}()

		host, err := pgContainer.Host(ctx)
		if err != nil {
			startErr = err
			return
		}

		mappedPort, err := pgContainer.MappedPort(ctx, pgPort)
		if err != nil {
			startErr = err
			return
		}

		pgDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.User, cfg.Password, host, mappedPort.Port(), cfg.DBName)
	})

	return pgContainer, pgDSN, startErr
}

// GetTestPGDSN возвращает DSN строку для подключения к тестовой БД PostgreSQL.
// Перед использованием необходимо вызвать StartPGContainer для инициализации контейнера.
func GetTestPGDSN() string {
	return pgDSN
}
