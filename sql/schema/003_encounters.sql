-- +goose Up
CREATE TABLE encounters (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    encounter_name     TEXT NOT NULL DEFAULT '',
    adversaries        TEXT NOT NULL DEFAULT '[]',
    custom_adversaries TEXT NOT NULL DEFAULT '[]',
    -- Slug of the attached environment, SRD or custom. Nullable: an encounter
    -- doesn't need one. Deliberately not a foreign key, since the slug may
    -- resolve against the embedded SRD rather than a row in this database.
    environment_slug   TEXT,
    created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

-- +goose Down
DROP TABLE encounters;
