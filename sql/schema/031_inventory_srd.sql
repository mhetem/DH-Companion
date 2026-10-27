-- +goose Up
ALTER TABLE inventory_items ADD COLUMN srd_slug TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE inventory_items DROP COLUMN srd_slug;
