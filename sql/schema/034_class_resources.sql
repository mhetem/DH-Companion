-- +goose Up
ALTER TABLE characters ADD COLUMN favor INTEGER NOT NULL DEFAULT 0 CHECK (favor BETWEEN 0 AND 6);

ALTER TABLE characters ADD COLUMN focus INTEGER NOT NULL DEFAULT 0 CHECK (focus BETWEEN 0 AND 6);

ALTER TABLE characters ADD COLUMN combo_die INTEGER NOT NULL DEFAULT 4
    CHECK (combo_die IN (4, 6, 8, 10, 12));

ALTER TABLE characters ADD COLUMN known_stances TEXT NOT NULL DEFAULT '[]';

ALTER TABLE characters ADD COLUMN active_stance TEXT NOT NULL DEFAULT '';

ALTER TABLE characters ADD COLUMN active_effects TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE characters DROP COLUMN active_effects;

ALTER TABLE characters DROP COLUMN active_stance;

ALTER TABLE characters DROP COLUMN known_stances;

ALTER TABLE characters DROP COLUMN combo_die;

ALTER TABLE characters DROP COLUMN focus;

ALTER TABLE characters DROP COLUMN favor;
