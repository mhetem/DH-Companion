-- +goose Up
CREATE TABLE custom_adversaries (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    slug            TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    tier            TEXT NOT NULL,
    type            TEXT NOT NULL,
    description     TEXT NOT NULL,
    horde_number    TEXT NOT NULL DEFAULT '',
    motives         TEXT NOT NULL,
    experiences     TEXT NOT NULL,
    difficulty      TEXT NOT NULL,
    threshold_minor TEXT NOT NULL,
    threshold_major TEXT NOT NULL,
    hp              TEXT NOT NULL,
    stress          TEXT NOT NULL,
    standard_attack TEXT,
    features        TEXT,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_custom_adversaries_tier_type ON custom_adversaries (tier, type);

-- +goose Down
DROP TABLE custom_adversaries;
