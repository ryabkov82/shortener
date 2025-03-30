package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

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

	err = initDB(db)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	getShortURLStmt, err := db.Prepare(`SELECT short_code FROM short_urls WHERE original_url = $1`)
	if err != nil {
		return nil, err
	}

	getURLStmt, err := db.Prepare(`SELECT original_url	FROM short_urls WHERE short_code = $1`)

	if err != nil {
		return nil, err
	}

	insertURLStmt, err := db.Prepare(`INSERT INTO short_urls (original_url, short_code) VALUES ($1, $2)`)

	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db, getShortURLStmt, getURLStmt, insertURLStmt}, nil

}

func initDB(db *sql.DB) error {

	// Создание таблицы, если она не существует
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS short_urls (
		id SERIAL PRIMARY KEY,
		original_url TEXT NOT NULL,
		short_code VARCHAR(20) NOT NULL UNIQUE
	);
	
	CREATE INDEX IF NOT EXISTS idx_short_urls_short_code ON short_urls(short_code);
    CREATE UNIQUE INDEX IF NOT EXISTS idx_short_urls_original_url_unique ON short_urls(original_url);
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

func (s *PostgresStorage) Ping() error {

	if s.db == nil {
		// База данных не инициализирована
		return errors.New("database is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.db.PingContext(ctx)
	return err
}

func (s *PostgresStorage) GetShortKey(originalURL string) (models.URLMapping, error) {

	mapping := models.URLMapping{
		OriginalURL: originalURL,
	}

	err := s.getShortURLStmt.QueryRow(originalURL).Scan(&mapping.ShortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mapping, storage.ErrURLNotFound
		}
		return mapping, fmt.Errorf("ошибка при поиске URL: %v", err)
	}

	return mapping, nil
}

func (s *PostgresStorage) GetRedirectURL(shortKey string) (models.URLMapping, error) {

	mapping := models.URLMapping{
		ShortURL: shortKey,
	}

	err := s.getURLStmt.QueryRow(shortKey).Scan(&mapping.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mapping, storage.ErrURLNotFound
		}
		return mapping, fmt.Errorf("ошибка при поиске URL: %v", err)
	}

	return mapping, nil

}

func (s *PostgresStorage) SaveURL(mapping models.URLMapping) error {

	_, err := s.insertURLStmt.Exec(mapping.OriginalURL, mapping.ShortURL)
	return err

}
