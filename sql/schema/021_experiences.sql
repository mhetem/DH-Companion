-- +goose Up
CREATE TABLE experiences (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    modifier     INTEGER NOT NULL DEFAULT 2 CHECK (modifier BETWEEN 0 AND 9),
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_experiences_character ON experiences (character_id);

-- +goose Down
DROP TABLE experiences;
