-- +goose Up
CREATE TABLE handouts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL DEFAULT 'scene'
                CHECK (kind IN ('map', 'scene', 'text')),
    title       TEXT NOT NULL DEFAULT 'Untitled',
    body        TEXT NOT NULL DEFAULT '',
    media       TEXT NOT NULL DEFAULT '',
    media_name  TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_handouts_campaign_kind ON handouts (campaign_id, kind);

CREATE INDEX idx_handouts_media ON handouts (media);

-- +goose Down
DROP TABLE handouts;
