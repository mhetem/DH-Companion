-- +goose Up
CREATE TABLE countdowns (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    value      INTEGER NOT NULL DEFAULT 0 CHECK (value BETWEEN 0 AND max),
    max        INTEGER NOT NULL CHECK (max > 0),
    kind       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

-- +goose Down
DROP TABLE countdowns;
