-- +goose Up
CREATE TABLE custom_weapons (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    slug         TEXT UNIQUE NOT NULL,
    name         TEXT NOT NULL,
    tier         TEXT NOT NULL,
    type         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    category     TEXT NOT NULL,
    trait        TEXT NOT NULL DEFAULT '',
    -- RANGE is a keyword in SQLite's window-function grammar; the column is
    -- prefixed so nothing has to be quoted at every call site.
    weapon_range TEXT NOT NULL DEFAULT '',
    damage       TEXT NOT NULL DEFAULT '',
    damage_type  TEXT NOT NULL DEFAULT '',
    burden       TEXT NOT NULL DEFAULT '',
    -- A single feature object as JSON, or '' for a weapon with no feature. The
    -- SRD never prints more than one, so this isn't the features array the
    -- adversary and environment tables carry.
    feature      TEXT NOT NULL DEFAULT '',
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_custom_weapons_tier_type ON custom_weapons (tier, type, category);

CREATE TABLE custom_armor (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    slug             TEXT UNIQUE NOT NULL,
    name             TEXT NOT NULL,
    tier             TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    threshold_major  INTEGER NOT NULL DEFAULT 0,
    threshold_severe INTEGER NOT NULL DEFAULT 0,
    base_score       INTEGER NOT NULL DEFAULT 0,
    feature          TEXT NOT NULL DEFAULT '',
    created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_custom_armor_tier ON custom_armor (tier);

-- Items and consumables have the same shape and differ only in kind, so they
-- share one table the way cards.Loot serves both. Homebrew loot carries no roll
-- number: a roll is an index into a full 60-row table, which a handful of
-- homebrew entries is not, so the roller samples these by rarity instead.
CREATE TABLE custom_loot (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    kind        TEXT NOT NULL CHECK (kind IN ('item', 'consumable')),
    slug        TEXT NOT NULL,
    name        TEXT NOT NULL,
    rarity      TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE (kind, slug)
) STRICT;

CREATE INDEX idx_custom_loot_kind_rarity ON custom_loot (kind, rarity);

-- +goose Down
DROP TABLE custom_loot;
DROP TABLE custom_armor;
DROP TABLE custom_weapons;
