/*
Package config предоставляет загрузку и валидацию конфигурации приложения.

Поддерживает несколько источников конфигурации:
- Аргументы командной строки
- Переменные окружения
- JSON-файлы конфигурации
- Значения по умолчанию

Приоритет настроек (от высшего к низшему):
1. Аргументы командной строки
2. Переменные окружения
3. JSON-файл конфигурации (если указан)
4. Значения по умолчанию

Формат JSON-конфигурации:

	{
	    "server_address": "localhost:8080",
	    "base_url": "http://localhost",
	    "file_storage_path": "/path/to/file.db",
	    "database_dsn": "",
	    "enable_https": true,
	    "jwt_secret": "secret_key",
	    "pprof": {
	        "enabled": true,
	        "auth_user": "admin",
	        "auth_pass": "password"
	    }
	}

Путь к JSON-файлу конфигурации можно указать:
- Через флаг -c или --config
- Через переменную окружения CONFIG
*/
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config содержит все параметры конфигурации приложения.
type Config struct {
	HTTPServerAddr string      `json:"server_address"`      // Адрес HTTP-сервера в формате host:port
	GRPCServerAddr string      `json:"grpc_server_address"` // Адрес gRPC-сервера
	BaseURL        string      `json:"base_url"`            // Базовый URL для сокращённых ссылок
	LogLevel       string      `json:"log_level"`           // Уровень логирования (debug, info, warn, error)
	FileStorage    string      `json:"file_storage_path"`   // Путь к файлу хранилища
	DBConnect      string      `json:"database_dsn"`        // Строка подключения к БД
	JwtKey         string      `json:"jwt_secret"`          // Секретный ключ для JWT
	ConfigPProf    PProfConfig `json:"pprof"`               // Настройки pprof
	EnableHTTPS    bool        `json:"enable_https"`        // Включение HTTPS
	SSLCertFile    string      `json:"ssl_cert_file"`       // Путь к SSL сертификату
	SSLKeyFile     string      `json:"ssl_key_file"`        // Путь к SSL ключу
	TrustedSubnet  string      `json:"trusted_subnet"`      // Доверенная подсеть
}

// PProfConfig содержит настройки профилирования pprof.
type PProfConfig struct {
	AuthUser string `json:"auth_user"`
	AuthPass string `json:"auth_pass"`
	Endpoint string `json:"endpoint"`
	BindAddr string `json:"bind_addr"`
	Enabled  bool   `json:"enabled"`
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

// validateCertFiles проверяет существование файлов сертификатов.
// Возвращает:
//
//	error - ошибка или nil
func validateCertFiles(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return errors.New("SSL certificate file not found")
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return errors.New("SSL key file not found")
	}
	return nil
}

func validateGRPCServerAddr(addr string) error {
	if addr == "" {
		return errors.New("gRPC server address cannot be empty")
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	// Проверка порта
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	return nil
}

// Load загружает конфигурацию из разных источников.
//
// Порядок загрузки:
// 1. Устанавливает значения по умолчанию
// 2. Читает JSON-конфиг (если указан)
// 3. Читает аргументы командной строки
// 4. Перезаписывает переменными окружения
//
// Возвращает:
// *Config - загруженную конфигурацию
func Load() *Config {
	cfg := &Config{
		HTTPServerAddr: "localhost:8080",
		GRPCServerAddr: "localhost:50051",
		BaseURL:        "http://localhost:8080",
		LogLevel:       "info",
		FileStorage:    "storage.dat",
		JwtKey:         "your_strong_secret_here",
		EnableHTTPS:    false,
		SSLCertFile:    "cert.pem",
		SSLKeyFile:     "key.pem",
		ConfigPProf: PProfConfig{
			Enabled:  true,
			AuthUser: "admin",
			AuthPass: "admin",
			Endpoint: "/debug/pprof",
			BindAddr: ":6060",
		},
	}

	// Загрузка из JSON-файла если указан
	configFile := getConfigFilePath()
	if configFile != "" {
		fileCfg, err := loadFromJSON(configFile)
		if err != nil {
			log.Printf("Ошибка загрузки JSON-конфига: %v", err)
		} else {
			// Объединяем конфиги, сохраняя значения по умолчанию для незаполненных полей
			mergeConfigs(cfg, fileCfg)
		}
	}

	// Загрузка из аргументов командной строки
	loadFromFlags(cfg)

	// Переопределение переменными окружения
	loadFromEnv(cfg)

	// Дополнительная обработка
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")

	// Валидация SSL файлов если HTTPS включен
	if cfg.EnableHTTPS {
		if err := validateCertFiles(cfg.SSLCertFile, cfg.SSLKeyFile); err != nil {
			log.Fatalf("HTTPS configuration error: %v", err)
		}
	}

	return cfg
}

// getConfigFilePath возвращает путь к файлу конфигурации из флагов или переменных окружения
func getConfigFilePath() string {

	for i, arg := range os.Args[1:] {
		if arg == "-c" || arg == "--config" {
			if i+1 < len(os.Args) {
				return os.Args[i+2]
			}
		}
		if strings.HasPrefix(arg, "-c=") {
			return strings.TrimPrefix(arg, "-c=")
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}
	return ""
}

// loadFromJSON загружает конфигурацию из JSON-файла
func loadFromJSON(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// mergeConfigs объединяет две конфигурации, сохраняя оригинальные значения для пустых полей
func mergeConfigs(original, new *Config) {
	if new.HTTPServerAddr != "" {
		original.HTTPServerAddr = new.HTTPServerAddr
	}
	if new.BaseURL != "" {
		original.BaseURL = new.BaseURL
	}
	if new.LogLevel != "" {
		original.LogLevel = new.LogLevel
	}
	if new.FileStorage != "" {
		original.FileStorage = new.FileStorage
	}
	if new.DBConnect != "" {
		original.DBConnect = new.DBConnect
	}
	if new.JwtKey != "" {
		original.JwtKey = new.JwtKey
	}
	if new.EnableHTTPS {
		original.EnableHTTPS = new.EnableHTTPS
	}
	if new.SSLCertFile != "" {
		original.SSLCertFile = new.SSLCertFile
	}
	if new.SSLKeyFile != "" {
		original.SSLKeyFile = new.SSLKeyFile
	}
	if new.TrustedSubnet != "" {
		original.TrustedSubnet = new.TrustedSubnet
	}
	if new.GRPCServerAddr != "" {
		original.GRPCServerAddr = new.GRPCServerAddr
	}

	// Объединение PProfConfig
	if new.ConfigPProf.AuthUser != "" {
		original.ConfigPProf.AuthUser = new.ConfigPProf.AuthUser
	}
	if new.ConfigPProf.AuthPass != "" {
		original.ConfigPProf.AuthPass = new.ConfigPProf.AuthPass
	}
	if new.ConfigPProf.Endpoint != "" {
		original.ConfigPProf.Endpoint = new.ConfigPProf.Endpoint
	}
	if new.ConfigPProf.BindAddr != "" {
		original.ConfigPProf.BindAddr = new.ConfigPProf.BindAddr
	}
	if new.ConfigPProf.Enabled {
		original.ConfigPProf.Enabled = new.ConfigPProf.Enabled
	}
}

// loadFromFlags загружает значения из флагов командной строки
func loadFromFlags(cfg *Config) {
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
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS server")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet in CIDR notation")

	flag.Func("ga", "gRPC server address in host:port format", func(flagValue string) error {
		if err := validateGRPCServerAddr(flagValue); err != nil {
			return err
		}
		cfg.GRPCServerAddr = flagValue
		return nil
	})

	flag.Parse()
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

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	// Обработка HTTPS настроек
	if envEnableHTTPS := os.Getenv("SSL_ENABLE"); envEnableHTTPS != "" {
		if v, err := strconv.ParseBool(envEnableHTTPS); err == nil {
			cfg.EnableHTTPS = v
		}
	}

	if envCert := os.Getenv("SSL_CERT_FILE"); envCert != "" {
		cfg.SSLCertFile = envCert
	}

	if envKey := os.Getenv("SSL_KEY_FILE"); envKey != "" {
		cfg.SSLKeyFile = envKey
	}

	if envSubnet := os.Getenv("TRUSTED_SUBNET"); envSubnet != "" {
		cfg.TrustedSubnet = envSubnet
	}

	if envGRPCAddr := os.Getenv("GRPC_SERVER_ADDRESS"); envGRPCAddr != "" {
		if err := validateGRPCServerAddr(envGRPCAddr); err != nil {
			log.Fatalf("invalid GRPC_SERVER_ADDRESS: %v", err)
		}
		cfg.GRPCServerAddr = envGRPCAddr
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
