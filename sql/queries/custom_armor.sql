-- name: ShowAllCustomArmor :many
SELECT * FROM custom_armor
WHERE (tier = sqlc.arg(tier) OR sqlc.arg(tier) = '')
ORDER BY name ASC;

-- name: GetCustomArmorBySlug :one
SELECT * FROM custom_armor
WHERE slug = ?;

-- name: DeleteCustomArmor :exec
DELETE FROM custom_armor
WHERE slug = ?;

-- name: CreateCustomArmor :one
INSERT INTO custom_armor (
  slug, name, tier, description, threshold_major, threshold_severe, base_score, feature
)
VALUES (?,?,?,?,?,?,?,?)
RETURNING *;

-- name: UpdateCustomArmor :one
UPDATE custom_armor SET
  name = ?,
  tier = ?,
  description = ?,
  threshold_major = ?,
  threshold_severe = ?,
  base_score = ?,
  feature = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE slug = ?
RETURNING *;
