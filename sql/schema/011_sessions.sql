-- +goose Up
CREATE TABLE sessions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    number      INTEGER NOT NULL CHECK (number > 0),
    title       TEXT NOT NULL DEFAULT 'Untitled',
    date        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    recap       TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE (campaign_id, number)
) STRICT;

CREATE INDEX idx_sessions_campaign ON sessions (campaign_id);

-- +goose Down
DROP TABLE sessions;
