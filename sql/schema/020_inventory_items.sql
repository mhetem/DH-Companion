-- +goose Up
CREATE TABLE inventory_items (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    kind         TEXT NOT NULL DEFAULT 'item'
                 CHECK (kind IN ('primary-weapon', 'secondary-weapon', 'armor', 'consumable', 'item')),
    qty          INTEGER NOT NULL DEFAULT 1 CHECK (qty BETWEEN 0 AND 999),
    equipped     INTEGER NOT NULL DEFAULT 0 CHECK (equipped IN (0, 1)),
    detail       TEXT NOT NULL DEFAULT '',
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_inventory_items_character ON inventory_items (character_id, kind);

-- +goose Down
DROP TABLE inventory_items;
