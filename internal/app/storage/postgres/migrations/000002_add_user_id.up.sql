-- +goose Up
BEGIN;

-- Добавляем колонку 
 ALTER TABLE short_urls ADD COLUMN IF NOT EXISTS  user_id UUID;

-- Индекс для ускорения запросов
CREATE UNIQUE INDEX IF NOT EXISTS idx_short_urls_user_url ON short_urls(user_id, original_url);

-- Удаляем не актуакльные индексы
DROP INDEX IF EXISTS idx_short_urls_original_url_unique;

COMMIT;