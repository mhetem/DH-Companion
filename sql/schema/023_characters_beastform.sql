-- +goose Up
ALTER TABLE characters
    ADD COLUMN beastform_slug TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE characters DROP COLUMN beastform_slug;
