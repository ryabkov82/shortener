-- +goose Down
BEGIN;

DROP TABLE IF EXISTS short_urls;

COMMIT;