-- +goose Up
CREATE TABLE combatants (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    combat_id      INTEGER NOT NULL REFERENCES combats(id) ON DELETE CASCADE,
    adversary_slug TEXT,
    display_name   TEXT NOT NULL,
    hp_max         INTEGER NOT NULL CHECK (hp_max >= 0),
    hp_marked      INTEGER NOT NULL DEFAULT 0 CHECK (hp_marked BETWEEN 0 AND hp_max),
    stress_max     INTEGER NOT NULL CHECK (stress_max >= 0),
    stress_marked  INTEGER NOT NULL DEFAULT 0 CHECK (stress_marked BETWEEN 0 AND stress_max),
    spotlight      INTEGER NOT NULL DEFAULT 0 CHECK (spotlight IN (0, 1)),
    created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_combatants_combat ON combatants (combat_id);

-- +goose Down
DROP TABLE combatants;
