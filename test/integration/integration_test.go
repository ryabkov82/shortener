package integration

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/test/testconfig"
)

func TestMain(m *testing.M) {

	ctx := context.Background()
	// 1. Запуск контейнера

	container, dsn, err := testconfig.StartPGContainer(ctx, testconfig.DefaultPGConfig())
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}()

	// 2. Экспорт DSN для других тестов
	os.Setenv("TEST_DB_DSN", dsn)
	defer os.Unsetenv("TEST_DB_DSN")

	// 3. Запуск всех тестов
	code := m.Run()
	os.Exit(code)
}
