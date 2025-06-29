package testconfig

import (
	"context"
	"fmt"
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

		pgContainer, startErr = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if startErr != nil {
			return
		}

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
