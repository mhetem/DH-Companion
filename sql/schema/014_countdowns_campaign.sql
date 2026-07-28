-- +goose Up
ALTER TABLE countdowns
    ADD COLUMN campaign_id INTEGER REFERENCES campaigns(id) ON DELETE CASCADE;

CREATE INDEX idx_countdowns_campaign ON countdowns (campaign_id);

-- +goose Down
DROP INDEX idx_countdowns_campaign;

ALTER TABLE countdowns DROP COLUMN campaign_id;
