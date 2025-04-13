package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/storage"
)

type PostgresStorage struct {
	db              *sql.DB
	getShortURLStmt *sql.Stmt
	getURLStmt      *sql.Stmt
	insertURLStmt   *sql.Stmt
}

func NewPostgresStorage(StoragePath string) (*PostgresStorage, error) {

	db, err := sql.Open("pgx", StoragePath)

	if err != nil {
		return nil, err
	}

	//err = initDB(db)
	// Применяем миграции из текущего пакета
	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	getShortURLStmt, err := db.Prepare(`SELECT short_code FROM short_urls WHERE original_url = $1 and user_id = $2`)
	if err != nil {
		return nil, err
	}

	getURLStmt, err := db.Prepare(`SELECT original_url	FROM short_urls WHERE short_code = $1 and user_id = $2`)

	if err != nil {
		return nil, err
	}

	insertURLStmt, err := db.Prepare(`
	INSERT INTO short_urls (original_url, short_code, user_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id, original_url) DO UPDATE SET
		original_url = EXCLUDED.original_url -- Фейковое обновление
	RETURNING short_code, xmax;
	`)

	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db, getShortURLStmt, getURLStmt, insertURLStmt}, nil

}

func (s *PostgresStorage) Ping(ctx context.Context) error {

	// устанавливаем таймаут 5 секунд
	ctxTm, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := s.db.PingContext(ctxTm)
	return err
}

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

func (s *PostgresStorage) GetRedirectURL(ctx context.Context, shortKey string) (models.URLMapping, error) {

	mapping := models.URLMapping{
		ShortURL: shortKey,
	}

	userID := ctx.Value(jwtauth.UserIDContextKey)

	err := s.getURLStmt.QueryRowContext(ctx, shortKey, userID).Scan(&mapping.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mapping, fmt.Errorf("%w", storage.ErrURLNotFound)
		}
		return mapping, fmt.Errorf("ошибка при поиске URL: %w", err)
	}

	return mapping, nil

}

func (s *PostgresStorage) SaveURL(ctx context.Context, mapping *models.URLMapping) error {

	var xmax int64 // Системный столбец, показывающий был ли конфликт

	userID := ctx.Value(jwtauth.UserIDContextKey)

	err := s.insertURLStmt.QueryRowContext(ctx, mapping.OriginalURL, mapping.ShortURL, userID).Scan(&mapping.ShortURL, &xmax)

	if err != nil {
		// сюда попадем в том числе, если был конфликт по полю short_code
		return err
	}
	// Если xmax > 0, значит запись с original_url уже существовала (был конфликт)
	if xmax > 0 {
		err = storage.ErrURLExists
	}

	return err

}

func (s *PostgresStorage) GetExistingURLs(ctx context.Context, originalURLs []string) (map[string]string, error) {

	existing := make(map[string]string)

	if len(originalURLs) == 0 {
		return existing, nil
	}

	// Создаем запрос с параметрами для всех URL
	query := "SELECT original_url, short_code FROM short_urls WHERE original_url = ANY($1) and user_id = $2"

	userID := ctx.Value(jwtauth.UserIDContextKey)

	// Просто передаем слайс - pgx/stdlib автоматически конвертирует
	rows, err := s.db.QueryContext(ctx, query, originalURLs, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var originalURL string
		var shortURL string
		if err := rows.Scan(&originalURL, &shortURL); err != nil {
			return nil, err
		}
		existing[originalURL] = shortURL
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *PostgresStorage) SaveNewURLs(ctx context.Context, urls []models.URLMapping) error {
	if len(urls) == 0 {
		return nil
	}

	userID := ctx.Value(jwtauth.UserIDContextKey)

	// Начинаем транзакцию
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Подготавливаем statement для пакетной вставки
	stmt, err := tx.Prepare("INSERT INTO short_urls (original_url, short_code, user_id) VALUES($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Выполняем вставку для каждого URL
	for _, url := range urls {
		_, err = stmt.ExecContext(ctx, url.OriginalURL, url.ShortURL, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetUserUrls(ctx context.Context, baseURL string) ([]models.URLMapping, error) {

	userID := ctx.Value(jwtauth.UserIDContextKey)

	// Создаем запрос поиска всех сокращенных пользователем url
	query := "SELECT original_url, short_code FROM short_urls WHERE user_id = $1"

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userURLs []models.URLMapping

	for rows.Next() {
		var originalURL string
		var shortURL string
		if err := rows.Scan(&originalURL, &shortURL); err != nil {
			return nil, err
		}
		userURLs = append(userURLs, models.URLMapping{
			OriginalURL: originalURL,
			ShortURL:    baseURL + "/" + shortURL,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return userURLs, nil
}
