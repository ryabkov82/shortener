-- +goose Up
BEGIN;

-- Добавляем колонку 
 ALTER TABLE short_urls ADD COLUMN IF NOT EXISTS  user_id UUID;

-- Убираем ограничение уникальности колонки short_code
DO $$
DECLARE
    constraint_name text;
BEGIN
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'short_urls'::regclass
    AND contype = 'u'
    AND conkey @> ARRAY(
        SELECT attnum 
        FROM pg_attribute 
        WHERE attrelid = 'short_urls'::regclass 
        AND attname = 'short_code'
    );

    IF constraint_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT %I', 'short_urls', constraint_name);
    END IF;
END $$;

-- Индекс для ускорения запросов
CREATE UNIQUE INDEX IF NOT EXISTS idx_short_urls_user_url ON short_urls(user_id, original_url);

-- Добавляем ограничение уникальности в разрезе user_id/original_url
CREATE UNIQUE INDEX IF NOT EXISTS idx_short_urls_user_code ON short_urls(user_id, short_code);

-- Удаляем не актуакльные индексы
DROP INDEX IF EXISTS idx_short_urls_short_code;
DROP INDEX IF EXISTS idx_short_urls_original_url_unique;

COMMIT;