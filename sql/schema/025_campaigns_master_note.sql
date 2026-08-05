-- +goose Up
ALTER TABLE campaigns
    ADD COLUMN master_note TEXT NOT NULL DEFAULT '';

ALTER TABLE campaigns
    ADD COLUMN master_note_updated_at TEXT NOT NULL DEFAULT '';

-- +goose StatementBegin
CREATE TRIGGER campaigns_master_search_body AFTER UPDATE OF master_note ON campaigns BEGIN
    DELETE FROM search WHERE entity = 'master' AND entity_id = old.id;
    INSERT INTO search (entity, entity_id, campaign_id, slug, title, body)
    SELECT 'master', new.id, new.id, '', new.name, new.master_note
    WHERE trim(new.master_note) <> '';
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER campaigns_master_search_title AFTER UPDATE OF name ON campaigns BEGIN
    DELETE FROM search WHERE entity = 'master' AND entity_id = old.id;
    INSERT INTO search (entity, entity_id, campaign_id, slug, title, body)
    SELECT 'master', new.id, new.id, '', new.name, new.master_note
    WHERE trim(new.master_note) <> '';
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER campaigns_master_search_delete AFTER DELETE ON campaigns BEGIN
    DELETE FROM search WHERE entity = 'master' AND entity_id = old.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER campaigns_master_search_delete;

DROP TRIGGER campaigns_master_search_title;

DROP TRIGGER campaigns_master_search_body;

ALTER TABLE campaigns DROP COLUMN master_note_updated_at;

ALTER TABLE campaigns DROP COLUMN master_note;
