// Package config предоставляет загрузку и валидацию конфигурации приложения.
//
// Поддерживает несколько источников конфигурации:
// - Аргументы командной строки
// - Переменные окружения
// - Значения по умолчанию
//
// Приоритет настроек:
// 1. Переменные окружения
// 2. Аргументы командной строки
// 3. Значения по умолчанию
package config

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config содержит все параметры конфигурации приложения.
type Config struct {
	HTTPServerAddr string      // Адрес HTTP-сервера в формате host:port
	BaseURL        string      // Базовый URL для сокращённых ссылок
	LogLevel       string      // Уровень логирования (debug, info, warn, error)
	FileStorage    string      // Путь к файлу хранилища
	DBConnect      string      // Строка подключения к БД
	JwtKey         string      // Секретный ключ для JWT
	ConfigPProf    PProfConfig // Настройки pprof
}

// PProfConfig содержит настройки профилирования pprof.
type PProfConfig struct {
	AuthUser string
	AuthPass string
	Endpoint string
	BindAddr string
	Enabled  bool
}

// validateHTTPServerAddr проверяет корректность адреса сервера.
//
// Формат адреса: host:port
// Где port должен быть числом от 1 до 65535
//
// Возвращает:
//
//	error - ошибка валидации или nil
func validateHTTPServerAddr(addr string) error {
	hp := strings.Split(addr, ":")
	if len(hp) != 2 {
		return errors.New("address must be in host:port format")
	}

	port, err := strconv.Atoi(hp[1])
	if err != nil || port < 1 || port > 65535 {
		return errors.New("port must be a number between 1 and 65535")
	}

	return nil
}

// validateBaseURL проверяет корректность базового URL.
//
// URL должен быть:
// - Абсолютным (содержать схему)
// - Валидным согласно net/url.Parse
//
// Возвращает:
//
//	error - ошибка валидации или nil
func validateBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	if u.Scheme == "" || u.Host == "" {
		return errors.New("URL must be absolute with scheme and host")
	}

	return nil
}

// Load загружает конфигурацию из разных источников.
//
// Порядок загрузки:
// 1. Устанавливает значения по умолчанию
// 2. Читает аргументы командной строки
// 3. Перезаписывает переменными окружения
//
// Возвращает:
//
//	*Config - загруженную конфигурацию
func Load() *Config {
	cfg := &Config{
		HTTPServerAddr: "localhost:8080",
		BaseURL:        "http://localhost:8080",
		LogLevel:       "info",
		FileStorage:    "storage.dat",
		JwtKey:         "your_strong_secret_here",
		ConfigPProf: PProfConfig{
			Enabled:  true,
			AuthUser: "admin",
			AuthPass: "admin",
			Endpoint: "/debug/pprof",
			BindAddr: ":6060",
		},
	}

	// Загрузка из аргументов командной строки
	flag.Func("a", "Server address in host:port format", func(flagValue string) error {
		if err := validateHTTPServerAddr(flagValue); err != nil {
			return err
		}
		cfg.HTTPServerAddr = flagValue
		return nil
	})

	flag.Func("b", "Base URL for shortened links (e.g. http://example.com)", func(flagValue string) error {
		if err := validateBaseURL(flagValue); err != nil {
			return err
		}
		cfg.BaseURL = strings.TrimSuffix(flagValue, "/")
		return nil
	})

	flag.StringVar(&cfg.LogLevel, "l", cfg.LogLevel, "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.FileStorage, "f", cfg.FileStorage, "Path to file storage")
	flag.StringVar(&cfg.DBConnect, "d", cfg.DBConnect, "Database connection string")
	flag.Parse()

	// Переопределение переменными окружения
	loadFromEnv(cfg)

	// Дополнительная обработка
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")

	return cfg
}

// loadFromEnv загружает значения из переменных окружения.
func loadFromEnv(cfg *Config) {
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		if err := validateHTTPServerAddr(envAddr); err != nil {
			log.Fatalf("invalid SERVER_ADDRESS: %v", err)
		}
		cfg.HTTPServerAddr = envAddr
	}

	if envURL := os.Getenv("BASE_URL"); envURL != "" {
		if err := validateBaseURL(envURL); err != nil {
			log.Fatalf("invalid BASE_URL: %v", err)
		}
		cfg.BaseURL = envURL
	}

	if envFile := os.Getenv("FILE_STORAGE_PATH"); envFile != "" {
		cfg.FileStorage = envFile
	}

	if envDB := os.Getenv("DATABASE_DSN"); envDB != "" {
		cfg.DBConnect = envDB
	}

	if envJWT := os.Getenv("JWT_SECRET"); envJWT != "" {
		if len(envJWT) < 32 {
			log.Fatal("JWT_SECRET must be at least 32 characters long")
		}
		cfg.JwtKey = envJWT
	}

	// Обработка pprof настроек
	if user := os.Getenv("PPROF_USER"); user != "" {
		cfg.ConfigPProf.AuthUser = user
	}
	if pass := os.Getenv("PPROF_PASS"); pass != "" {
		cfg.ConfigPProf.AuthPass = pass
	}
	if enabled := os.Getenv("PPROF_ENABLED"); enabled != "" {
		if v, err := strconv.ParseBool(enabled); err == nil {
			cfg.ConfigPProf.Enabled = v
		}
	}
}
