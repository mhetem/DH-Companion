-- name: CreateParty :one
INSERT INTO parties (name, size, tier) 
VALUES (?,?,?) 
RETURNING *;

-- name: ListParties :many
SELECT * FROM parties 
ORDER BY name ASC;

-- name: GetParty :one
SELECT * FROM parties 
WHERE id = ?;

-- name: UpdateParty :one
UPDATE parties SET
  name = ?,
  size = ?,
  tier = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: DeleteParty :exec
DELETE FROM parties WHERE id = ?;
