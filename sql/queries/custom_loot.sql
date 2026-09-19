-- name: ShowAllCustomLoot :many
SELECT * FROM custom_loot
WHERE kind = sqlc.arg(kind)
  AND (rarity = sqlc.arg(rarity) OR sqlc.arg(rarity) = '')
ORDER BY name ASC;

-- name: GetCustomLootBySlug :one
SELECT * FROM custom_loot
WHERE kind = ? AND slug = ?;

-- name: DeleteCustomLoot :exec
DELETE FROM custom_loot
WHERE kind = ? AND slug = ?;

-- name: CreateCustomLoot :one
INSERT INTO custom_loot (kind, slug, name, rarity, description)
VALUES (?,?,?,?,?)
RETURNING *;

-- name: UpdateCustomLoot :one
UPDATE custom_loot SET
  name = ?,
  rarity = ?,
  description = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE kind = ? AND slug = ?
RETURNING *;
