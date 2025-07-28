package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	// --- Тест 1: Загрузка конфигурации по умолчанию ---
	t.Run("Default config", func(t *testing.T) {
		// Создаем новый FlagSet для изоляции теста
		flag.CommandLine = flag.NewFlagSet("test1", flag.PanicOnError)
		os.Args = []string{"cmd"}
		cfg := Load()
		if cfg.HTTPServerAddr != "localhost:8080" {
			t.Errorf("Expected default server address 'localhost:8080', got '%s'", cfg.HTTPServerAddr)
		}
	})

	// --- Тест 2: Загрузка из JSON-файла ---
	t.Run("JSON config", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test2", flag.PanicOnError)
		os.Args = []string{"cmd"}
		configPath := filepath.Join("testdata", "valid_config.json")
		t.Setenv("CONFIG", configPath)

		cfg := Load()
		if cfg.HTTPServerAddr != "testhost:9090" {
			t.Errorf("Expected JSON server address 'testhost:9090', got '%s'", cfg.HTTPServerAddr)
		}
	})

	// --- Тест 3: Переопределение переменными окружения ---
	t.Run("Environment override", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test3", flag.PanicOnError)
		os.Args = []string{"cmd"}
		t.Setenv("SERVER_ADDRESS", "envhost:8080")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("JWT_SECRET", "env_jwt_secret_12345678901234567890")

		cfg := Load()
		if cfg.HTTPServerAddr != "envhost:8080" {
			t.Errorf("Expected env server address 'envhost:8080', got '%s'", cfg.HTTPServerAddr)
		}
		if cfg.LogLevel != "debug" {
			t.Errorf("Expected env log level 'debug', got '%s'", cfg.LogLevel)
		}
	})

	// --- Тест 4: Переопределение флагами командной строки ---
	t.Run("Flag override", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test4", flag.PanicOnError)
		os.Args = []string{"cmd", "-a", "flaghost:7070", "-l", "error"}

		cfg := Load()
		if cfg.HTTPServerAddr != "flaghost:7070" {
			t.Errorf("Expected flag server address 'flaghost:7070', got '%s'", cfg.HTTPServerAddr)
		}
		if cfg.LogLevel != "error" {
			t.Errorf("Expected flag log level 'error', got '%s'", cfg.LogLevel)
		}
	})

	// --- Тест 5: Валидация server address ---
	t.Run("Server address validation", func(t *testing.T) {
		tests := []struct {
			name    string
			addr    string
			wantErr bool
		}{
			{"Valid", "host:8080", false},
			{"No port", "host", true},
			{"Invalid port", "host:999999", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateHTTPServerAddr(tt.addr)
				if (err != nil) != tt.wantErr {
					t.Errorf("validateHTTPServerAddr() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	// --- Тест 6: Валидация base URL ---
	t.Run("Base URL validation", func(t *testing.T) {
		tests := []struct {
			name    string
			url     string
			wantErr bool
		}{
			{"Valid HTTP", "http://host", false},
			{"Valid HTTPS", "https://host", false},
			{"No scheme", "host", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateBaseURL(tt.url)
				if (err != nil) != tt.wantErr {
					t.Errorf("validateBaseURL() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	// --- Тест 7: Валидация SSL сертификатов ---
	t.Run("SSL cert validation", func(t *testing.T) {
		// Создаем временные файлы для теста
		certFile := filepath.Join(t.TempDir(), "cert.pem")
		keyFile := filepath.Join(t.TempDir(), "key.pem")
		os.WriteFile(certFile, []byte("test"), 0644)
		os.WriteFile(keyFile, []byte("test"), 0644)

		tests := []struct {
			name      string
			certFile  string
			keyFile   string
			wantError bool
		}{
			{"Valid files", certFile, keyFile, false},
			{"Missing cert", "missing.pem", keyFile, true},
			{"Missing key", certFile, "missing.pem", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateCertFiles(tt.certFile, tt.keyFile)
				if (err != nil) != tt.wantError {
					t.Errorf("validateCertFiles() error = %v, wantError %v", err, tt.wantError)
				}
			})
		}
	})

	// --- Тест 8: Объединение конфигураций ---
	t.Run("Config merging", func(t *testing.T) {
		original := &Config{
			HTTPServerAddr: "original:8080",
			BaseURL:        "http://original",
			LogLevel:       "info",
		}

		new := &Config{
			HTTPServerAddr: "new:9090",
			LogLevel:       "debug",
		}

		mergeConfigs(original, new)

		if original.HTTPServerAddr != "new:9090" {
			t.Error("HTTPServerAddr not updated")
		}
		if original.BaseURL != "http://original" {
			t.Error("BaseURL should remain unchanged")
		}
		if original.LogLevel != "debug" {
			t.Error("LogLevel not updated")
		}
	})

	// --- Тест 9: Невалидный JSON ---
	t.Run("Invalid JSON", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test9", flag.PanicOnError)
		os.Args = []string{"cmd"}
		configPath := filepath.Join("testdata", "invalid_config.json")
		t.Setenv("CONFIG", configPath)

		// Не должен паниковать
		cfg := Load()
		if cfg.HTTPServerAddr != "localhost:8080" {
			t.Error("Should fall back to default values")
		}
	})

	// --- Тест 10: PPROF конфигурация ---
	t.Run("PPROF config", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test10", flag.PanicOnError)
		os.Args = []string{"cmd"}
		t.Setenv("PPROF_ENABLED", "false")
		t.Setenv("PPROF_USER", "testuser")
		t.Setenv("PPROF_PASS", "testpass")

		cfg := Load()
		if cfg.ConfigPProf.Enabled {
			t.Error("PPROF should be disabled")
		}
		if cfg.ConfigPProf.AuthUser != "testuser" {
			t.Error("PPROF user not updated")
		}
	})

	// --- Тест 11: Приоритеты конфигурации ---
	t.Run("Configuration priorities", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test11", flag.PanicOnError)
		// 1. JSON config (низший приоритет)
		configPath := filepath.Join("testdata", "valid_config.json")
		t.Setenv("CONFIG", configPath) // server_address = "testhost:9090" в JSON

		// 2. Флаги (средний приоритет)
		os.Args = []string{"cmd", "-a", "flaghost:7070"}

		// 3. Переменные окружения (высший приоритет)
		t.Setenv("SERVER_ADDRESS", "envhost:8080")

		cfg := Load()

		// Проверяем приоритеты (env > flags > JSON)
		if cfg.HTTPServerAddr != "envhost:8080" {
			t.Errorf("Expected 'envhost:8080' (env has highest priority), got '%s'", cfg.HTTPServerAddr)
		}
	})

	// --- Тест 12: Валидация gRPC server address ---
	t.Run("gRPC server address validation", func(t *testing.T) {

		cases := []struct {
			name    string
			addr    string
			isValid bool
		}{
			{"Valid 1", ":50051", true},
			{"Valid 2", "localhost:50051", true},
			{"Valid 3", "127.0.0.1:50051", true},
			{"Valid 4", "[::1]:50051", true},
			{"No addr", "", false},
			{"Invalid port", ":99999", false},
			{"Missing port", "localhost", false},
		}

		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				err := validateGRPCServerAddr(tt.addr)
				if tt.isValid && err != nil {
					t.Errorf("%q should be valid: %v", tt.addr, err)
				}
				if !tt.isValid && err == nil {
					t.Errorf("%q should be invalid", tt.addr)
				}
			})

		}
	})

}
