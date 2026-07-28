-- +goose Up
ALTER TABLE combats
    ADD COLUMN campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL;

CREATE INDEX idx_combats_campaign ON combats (campaign_id);

-- +goose Down
DROP INDEX idx_combats_campaign;

ALTER TABLE combats DROP COLUMN campaign_id;
