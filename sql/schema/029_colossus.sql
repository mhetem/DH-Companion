-- +goose Up
ALTER TABLE custom_adversaries ADD COLUMN size TEXT NOT NULL DEFAULT '';

ALTER TABLE custom_adversaries ADD COLUMN segments TEXT NOT NULL DEFAULT '[]';

ALTER TABLE combatants
    ADD COLUMN colossus_id INTEGER REFERENCES combatants(id) ON DELETE CASCADE;

ALTER TABLE combatants ADD COLUMN segment TEXT NOT NULL DEFAULT '';

ALTER TABLE combatants ADD COLUMN broken INTEGER NOT NULL DEFAULT 0 CHECK (broken IN (0, 1));

CREATE INDEX idx_combatants_colossus ON combatants (colossus_id);

-- +goose Down
DROP INDEX idx_combatants_colossus;

ALTER TABLE combatants DROP COLUMN broken;

ALTER TABLE combatants DROP COLUMN segment;

ALTER TABLE combatants DROP COLUMN colossus_id;

ALTER TABLE custom_adversaries DROP COLUMN segments;

ALTER TABLE custom_adversaries DROP COLUMN size;
