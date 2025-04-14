-- +goose Up
BEGIN;

-- Создаем таблицу short_urls в текущем состоянии
CREATE TABLE IF NOT EXISTS short_urls (
    id SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_code VARCHAR(20) NOT NULL UNIQUE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_short_urls_short_code_unique ON short_urls(short_code);
--CREATE UNIQUE INDEX IF NOT EXISTS idx_short_urls_original_url_unique ON short_urls(original_url);

COMMIT;