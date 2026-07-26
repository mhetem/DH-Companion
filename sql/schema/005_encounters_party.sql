-- +goose Up
ALTER TABLE encounters
    ADD COLUMN party_id INTEGER REFERENCES parties(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE encounters DROP COLUMN party_id;
