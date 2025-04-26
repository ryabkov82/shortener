-- +goose Up
BEGIN;

-- Добавляем колонку 
 ALTER TABLE short_urls ADD COLUMN IF NOT EXISTS  is_deleted BOOLEAN DEFAULT FALSE;

COMMIT;