
-- name: ShowAllCustomAdversaries :many
SELECT * FROM custom_adversaries
WHERE (tier = sqlc.arg(tier) OR sqlc.arg(tier) = '')
  AND (type = sqlc.arg(type) OR sqlc.arg(type) = 'All')
ORDER BY name ASC;

-- name: GetCustomBySlug :one
SELECT * FROM custom_adversaries
WHERE slug = ?;

-- name: GetCustomByIDs :many
SELECT * FROM custom_adversaries
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteCustomAdversary :exec
DELETE FROM custom_adversaries
WHERE slug = ?;

-- name: CreateCustomAdversary :one
INSERT INTO custom_adversaries (
  slug, name, tier, type, description, horde_number, motives, experiences,
  difficulty, threshold_minor, threshold_major, hp, stress,
  standard_attack, features
)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
RETURNING *;

-- name: UpdateCustomAdversary :one
UPDATE custom_adversaries SET
  name = ?,
  tier = ?,
  type = ?,
  description = ?,
  horde_number = ?,
  motives = ?,
  experiences = ?,
  difficulty = ?,
  threshold_minor = ?,
  threshold_major = ?,
  hp = ?,
  stress = ?,
  standard_attack = ?,
  features = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;
