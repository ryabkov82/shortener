// Package postgres предоставляет реализацию хранилища URL на PostgreSQL.
//
// Пакет включает:
// - Подключение к PostgreSQL с настройкой пула соединений
// - Подготовленные SQL-запросы для повышения производительности
// - Поддержку транзакций для пакетных операций
// - Обработку миграций базы данных
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/storage"
)

// PostgresStorage реализует интерфейс хранилища для работы с PostgreSQL.
type PostgresStorage struct {
	db              *sql.DB
	getShortURLStmt *sql.Stmt
	getURLStmt      *sql.Stmt
	insertURLStmt   *sql.Stmt
}

// NewPostgresStorage создает новое подключение к PostgreSQL и инициализирует хранилище.
//
// Параметры:
//   - StoragePath: строка подключения к PostgreSQL
//
// Возвращает:
//   - *PostgresStorage: инициализированное хранилище
//   - error: ошибка при подключении или инициализации
func NewPostgresStorage(storagePath string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", storagePath)
	if err != nil {
		return nil, err
	}

	if err = applyMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Подготовка часто используемых запросов
	getShortURLStmt, err := db.Prepare(`SELECT short_code FROM short_urls WHERE original_url = $1 and user_id = $2`)
	if err != nil {
		return nil, err
	}

	getURLStmt, err := db.Prepare(`SELECT original_url, is_deleted FROM short_urls WHERE short_code = $1`)
	if err != nil {
		return nil, err
	}

	insertURLStmt, err := db.Prepare(`
	INSERT INTO short_urls (original_url, short_code, user_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id, original_url) DO UPDATE SET
		original_url = EXCLUDED.original_url
	RETURNING short_code, xmax;
	`)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db, getShortURLStmt, getURLStmt, insertURLStmt}, nil
}

// Ping проверяет соединение с базой данных.
//
// Параметры:
//
//	ctx - контекст выполнения
//
// Возвращает:
//
//	error - ошибка соединения
func (s *PostgresStorage) Ping(ctx context.Context) error {
	ctxTm, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.db.PingContext(ctxTm)
}

// GetShortKey возвращает сокращенный URL для оригинального.
//
// Параметры:
//
//	ctx - контекст выполнения
//	originalURL - оригинальный URL
//
// Возвращает:
//
//	models.URLMapping - соответствие URL
//	error - ошибка операции
func (s *PostgresStorage) GetShortKey(ctx context.Context, originalURL string) (models.URLMapping, error) {
	mapping := models.URLMapping{
		OriginalURL: originalURL,
	}

	userID := ctx.Value(jwtauth.UserIDContextKey)
	err := s.getShortURLStmt.QueryRowContext(ctx, originalURL, userID).Scan(&mapping.ShortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mapping, storage.ErrURLNotFound
		}
		return mapping, err
	}

	return mapping, nil
}

// GetRedirectURL возвращает оригинальный URL для сокращенного.
//
// Параметры:
//
//	ctx - контекст выполнения
//	shortKey - сокращенный ключ URL
//
// Возвращает:
//
//	models.URLMapping - соответствие URL
//	error - ошибка операции
func (s *PostgresStorage) GetRedirectURL(ctx context.Context, shortKey string) (models.URLMapping, error) {
	mapping := models.URLMapping{
		ShortURL: shortKey,
	}

	var deletedFlag bool
	err := s.getURLStmt.QueryRowContext(ctx, shortKey).Scan(&mapping.OriginalURL, &deletedFlag)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mapping, fmt.Errorf("%w", storage.ErrURLNotFound)
		}
		return mapping, fmt.Errorf("ошибка при поиске URL: %w", err)
	}

	if deletedFlag {
		return mapping, storage.ErrURLDeleted
	}
	return mapping, nil
}

// SaveURL сохраняет новое соответствие URL.
//
// Параметры:
//
//	ctx - контекст выполнения
//	mapping - соответствие URL для сохранения
//
// Возвращает:
//
//	error - ошибка операции
func (s *PostgresStorage) SaveURL(ctx context.Context, mapping *models.URLMapping) error {
	var xmax int64 // Системный столбец для определения конфликтов

	userID := ctx.Value(jwtauth.UserIDContextKey)
	err := s.insertURLStmt.QueryRowContext(ctx, mapping.OriginalURL, mapping.ShortURL, userID).Scan(&mapping.ShortURL, &xmax)

	if err != nil {
		return err
	}

	if xmax > 0 {
		err = storage.ErrURLExists
	}

	return err
}

// GetExistingURLs возвращает существующие сокращения для URL.
//
// Параметры:
//
//	ctx - контекст выполнения
//	originalURLs - список оригинальных URL
//
// Возвращает:
//
//	map[string]string - соответствия URL
//	error - ошибка операции
func (s *PostgresStorage) GetExistingURLs(ctx context.Context, originalURLs []string) (map[string]string, error) {
	existing := make(map[string]string)
	if len(originalURLs) == 0 {
		return existing, nil
	}

	query := "SELECT original_url, short_code FROM short_urls WHERE original_url = ANY($1) and user_id = $2"
	userID := ctx.Value(jwtauth.UserIDContextKey)

	rows, err := s.db.QueryContext(ctx, query, originalURLs, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var originalURL, shortURL string
		if err := rows.Scan(&originalURL, &shortURL); err != nil {
			return nil, err
		}
		existing[originalURL] = shortURL
	}

	return existing, rows.Err()
}

// SaveNewURLs сохраняет пакет новых URL.
//
// Параметры:
//
//	ctx - контекст выполнения
//	urls - список URL для сохранения
//
// Возвращает:
//
//	error - ошибка операции
func (s *PostgresStorage) SaveNewURLs(ctx context.Context, urls []models.URLMapping) error {
	if len(urls) == 0 {
		return nil
	}

	userID := ctx.Value(jwtauth.UserIDContextKey)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO short_urls (original_url, short_code, user_id) VALUES($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err = stmt.ExecContext(ctx, url.OriginalURL, url.ShortURL, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetUserUrls возвращает все URL пользователя.
//
// Параметры:
//
//	ctx - контекст выполнения
//	baseURL - базовый URL сервиса
//
// Возвращает:
//
//	[]models.URLMapping - список URL пользователя
//	error - ошибка операции
func (s *PostgresStorage) GetUserUrls(ctx context.Context, baseURL string) ([]models.URLMapping, error) {
	userID := ctx.Value(jwtauth.UserIDContextKey)
	query := "SELECT original_url, short_code FROM short_urls WHERE user_id = $1"

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userURLs []models.URLMapping
	for rows.Next() {
		var originalURL, shortURL string
		if err := rows.Scan(&originalURL, &shortURL); err != nil {
			return nil, err
		}
		userURLs = append(userURLs, models.URLMapping{
			OriginalURL: originalURL,
			ShortURL:    baseURL + "/" + shortURL,
		})
	}

	return userURLs, rows.Err()
}

// BatchMarkAsDeleted помечает URL пользователя как удаленные.
//
// Параметры:
//
//	userID - идентификатор пользователя
//	urls - список коротких URL для удаления
//
// Возвращает:
//
//	error - ошибка операции
func (s *PostgresStorage) BatchMarkAsDeleted(userID string, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	var params []interface{}
	query := "UPDATE short_urls SET is_deleted = true WHERE short_code IN ("

	for i, url := range urls {
		query += fmt.Sprintf("$%d,", i+1)
		params = append(params, url)
	}
	query = strings.TrimSuffix(query, ",") + ") AND user_id = $" + fmt.Sprintf("%d", len(urls)+1)
	params = append(params, userID)

	_, err := s.db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("error updating batch: %w", err)
	}

	return nil
}

// Close освобождает ресурсы
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
