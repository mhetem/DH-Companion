-- +goose Up
ALTER TABLE countdowns ADD COLUMN beats TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE countdowns DROP COLUMN beats;
