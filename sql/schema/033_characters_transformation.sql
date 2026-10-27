-- +goose Up
ALTER TABLE characters ADD COLUMN transformation_slug TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE characters DROP COLUMN transformation_slug;
