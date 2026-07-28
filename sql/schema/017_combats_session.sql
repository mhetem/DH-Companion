-- +goose Up
ALTER TABLE combats
    ADD COLUMN session_id INTEGER REFERENCES sessions(id) ON DELETE SET NULL;

CREATE INDEX idx_combats_session ON combats (session_id);

-- +goose Down
DROP INDEX idx_combats_session;

ALTER TABLE combats DROP COLUMN session_id;
