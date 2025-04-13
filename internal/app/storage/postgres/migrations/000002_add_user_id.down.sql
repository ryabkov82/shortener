-- +goose Down
BEGIN;

DROP INDEX IF EXISTS idx_short_urls_user_url;
DROP INDEX IF EXISTS idx_short_urls_user_code;
ALTER TABLE short_urls DROP COLUMN IF EXISTS user_id;

COMMIT;