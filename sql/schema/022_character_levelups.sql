-- +goose Up
CREATE TABLE character_levelups (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    level        INTEGER NOT NULL CHECK (level BETWEEN 2 AND 10),
    tier         INTEGER NOT NULL CHECK (tier BETWEEN 2 AND 4),
    choices      TEXT NOT NULL DEFAULT '[]',
    summary      TEXT NOT NULL DEFAULT '',
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE (character_id, level)
) STRICT;

CREATE INDEX idx_character_levelups_character ON character_levelups (character_id, tier);

-- +goose Down
DROP TABLE character_levelups;
