-- +goose Up
CREATE TABLE campaigns (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    current_fear INTEGER NOT NULL DEFAULT 0 CHECK (current_fear BETWEEN 0 AND 12),
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

-- +goose Down
DROP TABLE campaigns;
