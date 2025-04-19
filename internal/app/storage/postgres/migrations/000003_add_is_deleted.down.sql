-- +goose Down
BEGIN;

ALTER TABLE short_urls DROP COLUMN IF EXISTS is_deleted;

COMMIT;