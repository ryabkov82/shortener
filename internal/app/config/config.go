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

type Config struct {
	HTTPServerAddr string
	BaseURL        string
	LogLevel       string
	FileStorage    string
	DBConnect      string
	JwtKey         string
}

func validateHTTPServerAddr(addr string) error {

	hp := strings.Split(addr, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	_, err := strconv.Atoi(hp[1])

	return err
}

func validateBaseURL(baseURL string) error {

	_, err := url.Parse(baseURL)

	return err
}

func Load() *Config {

	cfg := new(Config)
	cfg.HTTPServerAddr = "localhost:8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.JwtKey = "your_strong_secret_here"

	flag.Func("a", "Server address host:port", func(flagValue string) error {

		err := validateHTTPServerAddr(flagValue)

		if err != nil {
			return err
		}

		cfg.HTTPServerAddr = flagValue
		return nil
	})

	flag.Func("b", "Base address shortened url", func(flagValue string) error {

		err := validateBaseURL(flagValue)

		if err != nil {
			return err
		}

		cfg.BaseURL = flagValue
		return nil
	})

	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")

	flag.StringVar(&cfg.FileStorage, "f", "storage.dat", "File storage path")

	flag.StringVar(&cfg.DBConnect, "d", "", "Database connect string")

	flag.Parse()

	if envHTTPServerAddr := os.Getenv("SERVER_ADDRESS"); envHTTPServerAddr != "" {

		err := validateHTTPServerAddr(envHTTPServerAddr)
		if err != nil {
			log.Fatalf("error validate SERVER_ADDRESS: %s", err)
		}

		cfg.HTTPServerAddr = envHTTPServerAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {

		err := validateBaseURL(envBaseURL)
		if err != nil {
			log.Fatalf("error validate BASE_URL: %s", err)
		}

		cfg.BaseURL = envBaseURL
	}

	if envFileStorage := os.Getenv("FILE_STORAGE_PATH"); envFileStorage != "" {
		cfg.FileStorage = envFileStorage
	}

	if envDBConnect := os.Getenv("DATABASE_DSN"); envDBConnect != "" {
		cfg.DBConnect = envDBConnect
	}

	if envJWTSECRET := os.Getenv("JWT_SECRET"); envJWTSECRET != "" {
		if len(envJWTSECRET) < 32 {
			log.Fatal("JWT_SECRET must be at least 32 characters long")
		}

		cfg.JwtKey = envJWTSECRET
	}

	// Убедимся, что BaseURL не заканчивается на "/"
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")

	return cfg

}
