-- +goose Down
BEGIN;

DROP INDEX IF EXISTS idx_short_urls_user_url;
ALTER TABLE short_urls DROP COLUMN IF EXISTS user_id;

COMMIT;