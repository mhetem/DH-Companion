-- +goose Up
CREATE VIRTUAL TABLE search USING fts5(
    entity UNINDEXED,
    entity_id UNINDEXED,
    campaign_id UNINDEXED,
    slug UNINDEXED,
    title,
    body,
    tokenize = 'porter unicode61'
);

-- +goose StatementBegin
CREATE TRIGGER notes_search_insert AFTER INSERT ON notes BEGIN
    INSERT INTO search (entity, entity_id, campaign_id, slug, title, body)
    VALUES ('note', new.id, new.campaign_id, '', new.title, new.body);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER notes_search_update AFTER UPDATE ON notes BEGIN
    DELETE FROM search WHERE entity = 'note' AND entity_id = old.id;
    INSERT INTO search (entity, entity_id, campaign_id, slug, title, body)
    VALUES ('note', new.id, new.campaign_id, '', new.title, new.body);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER notes_search_delete AFTER DELETE ON notes BEGIN
    DELETE FROM search WHERE entity = 'note' AND entity_id = old.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER notes_search_delete;

DROP TRIGGER notes_search_update;

DROP TRIGGER notes_search_insert;

DROP TABLE search;
