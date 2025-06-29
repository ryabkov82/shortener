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

type PGConfig struct {
	Image    string
	User     string
	Password string
	DBName   string
	Port     string
}

func DefaultPGConfig() PGConfig {
	return PGConfig{
		Image:    "postgres:13-alpine",
		User:     "test",
		Password: "test",
		DBName:   "test",
		Port:     "5432",
	}
}

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
				wait.ForSQL(
					pgPort,
					"postgres",
					func(host string, port nat.Port) string {
						return fmt.Sprintf("host=%s port=%s user=test password=test dbname=test sslmode=disable",
							host, port.Port())
					}).WithStartupTimeout(1*time.Minute).WithPollInterval(1*time.Second),
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

		mappedPort, err := pgContainer.MappedPort(ctx, pgPort)
		if err != nil {
			startErr = err
			return
		}

		pgDSN = fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable",
			cfg.User, cfg.Password, mappedPort.Port(), cfg.DBName)
	})

	return pgContainer, pgDSN, startErr
}

func GetTestPGDSN() string {
	return pgDSN
}
