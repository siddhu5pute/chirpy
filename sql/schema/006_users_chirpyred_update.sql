-- +goose Up
ALTER TABLE users
ALTER COLUMN is_chirpy_red SET NOT NULL,
ALTER COLUMN is_chirpy_red SET DEFAULT false;

-- +goose Down
ALTER TABLE users
ALTER COLUMN is_chirpy_red DROP NOT NULL,
ALTER COLUMN is_chirpy_red DROP DEFAULT;