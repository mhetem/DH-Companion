-- +goose Up
CREATE TABLE character_domain_cards (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    card_slug    TEXT NOT NULL,
    location     TEXT NOT NULL DEFAULT 'loadout' CHECK (location IN ('loadout', 'vault')),
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE (character_id, card_slug)
) STRICT;

CREATE INDEX idx_character_domain_cards_character ON character_domain_cards (character_id, location);

-- +goose Down
DROP TABLE character_domain_cards;
