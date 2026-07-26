-- +goose Up
CREATE TABLE custom_environments (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    slug                  TEXT UNIQUE NOT NULL,
    name                  TEXT NOT NULL,
    tier                  TEXT NOT NULL,
    type                  TEXT NOT NULL,
    description           TEXT NOT NULL,
    impulses              TEXT NOT NULL DEFAULT '',
    -- Difficulty is TEXT, not INTEGER: SRD environments such as Ambushed print
    -- "Special (see 'Relative Strength')" instead of a number.
    difficulty            TEXT NOT NULL,
    potential_adversaries TEXT NOT NULL DEFAULT '[]',
    features              TEXT NOT NULL DEFAULT '[]',
    created_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_custom_environments_tier_type ON custom_environments (tier, type);

-- +goose Down
DROP TABLE custom_environments;
