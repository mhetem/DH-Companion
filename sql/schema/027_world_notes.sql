-- +goose Up
CREATE TABLE world_notes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    parent_id   INTEGER REFERENCES world_notes(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL DEFAULT 'page',
    title       TEXT NOT NULL DEFAULT 'Untitled',
    world_date  TEXT NOT NULL DEFAULT '',
    body        TEXT NOT NULL DEFAULT '',
    position    INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
) STRICT;

CREATE INDEX idx_world_notes_campaign_position ON world_notes (campaign_id, position);

CREATE INDEX idx_world_notes_parent ON world_notes (parent_id);

-- +goose StatementBegin
CREATE TRIGGER world_notes_search_insert AFTER INSERT ON world_notes BEGIN
    INSERT INTO search (entity, entity_id, campaign_id, slug, title, body)
    VALUES ('world', new.id, new.campaign_id, '', new.title, new.body);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER world_notes_search_update AFTER UPDATE OF title, body ON world_notes BEGIN
    DELETE FROM search WHERE entity = 'world' AND entity_id = old.id;
    INSERT INTO search (entity, entity_id, campaign_id, slug, title, body)
    VALUES ('world', new.id, new.campaign_id, '', new.title, new.body);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER world_notes_search_delete AFTER DELETE ON world_notes BEGIN
    DELETE FROM search WHERE entity = 'world' AND entity_id = old.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER world_notes_search_delete;

DROP TRIGGER world_notes_search_update;

DROP TRIGGER world_notes_search_insert;

DROP TABLE world_notes;
