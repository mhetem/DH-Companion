-- +goose Up
CREATE TABLE notes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL DEFAULT 'npc'
                CHECK (kind IN ('npc', 'location', 'faction', 'lore', 'plot')),
    title       TEXT NOT NULL DEFAULT 'Untitled',
    body        TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_notes_campaign_kind ON notes (campaign_id, kind);

-- +goose Down
DROP TABLE notes;
