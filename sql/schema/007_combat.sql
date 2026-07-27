-- +goose Up
CREATE TABLE combats (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    encounter_id INTEGER REFERENCES encounters(id) ON DELETE SET NULL,
    fear         INTEGER NOT NULL DEFAULT 0 CHECK (fear BETWEEN 0 AND 12),
    active       INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_combats_encounter ON combats (encounter_id);
CREATE INDEX idx_combats_active ON combats (active) WHERE active = 1;

-- +goose Down
DROP TABLE combats;
