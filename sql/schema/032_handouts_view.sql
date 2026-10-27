-- +goose Up
ALTER TABLE handouts ADD COLUMN view_rotate INTEGER NOT NULL DEFAULT 0
    CHECK (view_rotate IN (0, 90, 180, 270));

ALTER TABLE handouts ADD COLUMN view_zoom REAL NOT NULL DEFAULT 1;

ALTER TABLE handouts ADD COLUMN view_x REAL NOT NULL DEFAULT 0;

ALTER TABLE handouts ADD COLUMN view_y REAL NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE handouts DROP COLUMN view_y;

ALTER TABLE handouts DROP COLUMN view_x;

ALTER TABLE handouts DROP COLUMN view_zoom;

ALTER TABLE handouts DROP COLUMN view_rotate;
