-- +goose Up
CREATE TABLE companions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id  INTEGER NOT NULL UNIQUE REFERENCES characters(id) ON DELETE CASCADE,
    name          TEXT NOT NULL DEFAULT '',
    evasion       INTEGER NOT NULL DEFAULT 10 CHECK (evasion BETWEEN 0 AND 30),
    damage_die    TEXT NOT NULL DEFAULT 'd6',
    attack_range  TEXT NOT NULL DEFAULT 'Melee',
    attack        TEXT NOT NULL DEFAULT '',
    stress_max    INTEGER NOT NULL DEFAULT 5 CHECK (stress_max BETWEEN 1 AND 12),
    stress_marked INTEGER NOT NULL DEFAULT 0 CHECK (stress_marked BETWEEN 0 AND stress_max),
    experiences   TEXT NOT NULL DEFAULT '[]',
    upgrades      TEXT NOT NULL DEFAULT '[]',
    notes         TEXT NOT NULL DEFAULT '',
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

-- +goose Down
DROP TABLE companions;
