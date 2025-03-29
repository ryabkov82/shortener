package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(StoragePath string) (*PostgresStorage, error) {

	db, err := sql.Open("pgx", StoragePath)

	if err != nil {
		return nil, err
	}
	return &PostgresStorage{db}, nil

}

func (s *PostgresStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.db.PingContext(ctx)
	return err
}
