-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN IF NOT EXISTS icon_path TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS icon_path;
-- +goose StatementEnd
