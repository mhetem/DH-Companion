-- name: IndexCard :exec
INSERT INTO search (entity, entity_id, campaign_id, slug, title, body)
VALUES (?,0,0,?,?,?);

-- name: ClearCardIndex :exec
DELETE FROM search
WHERE entity IN ('adversary', 'environment');
