-- name: AddCharacterDomainCard :one
INSERT INTO character_domain_cards (character_id, card_slug, location)
VALUES (?,?,?)
RETURNING *;

-- name: GetCharacterDomainCard :one
SELECT * FROM character_domain_cards
WHERE character_id = ? AND card_slug = ?;

-- name: ListCharacterDomainCards :many
SELECT * FROM character_domain_cards
WHERE character_id = ?
ORDER BY id ASC;

-- name: CountCharacterLoadout :one
SELECT count(*) FROM character_domain_cards
WHERE character_id = ? AND location = 'loadout';

-- name: SetDomainCardLocation :one
UPDATE character_domain_cards SET
  location = sqlc.arg(location),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE character_id = sqlc.arg(character_id) AND card_slug = sqlc.arg(card_slug)
RETURNING *;

-- name: RemoveCharacterDomainCard :exec
DELETE FROM character_domain_cards
WHERE character_id = ? AND card_slug = ?;
