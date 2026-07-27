-- name: ShowAllCustomEnvironments :many
SELECT * FROM custom_environments
WHERE (tier = sqlc.arg(tier) OR sqlc.arg(tier) = '')
  AND (type = sqlc.arg(type) OR sqlc.arg(type) = 'All')
ORDER BY name ASC;

-- name: GetCustomEnvironmentBySlug :one
SELECT * FROM custom_environments
WHERE slug = ?;

-- name: GetCustomEnvironmentsByIDs :many
SELECT * FROM custom_environments
WHERE id IN (sqlc.slice('ids'));

-- name: DeleteCustomEnvironment :exec
DELETE FROM custom_environments
WHERE slug = ?;

-- name: CreateCustomEnvironment :one
INSERT INTO custom_environments (
  slug, name, tier, type, description, impulses, difficulty,
  potential_adversaries, features
)
VALUES (?,?,?,?,?,?,?,?,?)
RETURNING *;

-- name: UpdateCustomEnvironment :one
UPDATE custom_environments SET
  name = ?,
  tier = ?,
  type = ?,
  description = ?,
  impulses = ?,
  difficulty = ?,
  potential_adversaries = ?,
  features = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;
