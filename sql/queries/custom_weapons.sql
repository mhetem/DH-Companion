-- name: ShowAllCustomWeapons :many
SELECT * FROM custom_weapons
WHERE (tier = sqlc.arg(tier) OR sqlc.arg(tier) = '')
  AND (type = sqlc.arg(type) OR sqlc.arg(type) = '')
  AND (category = sqlc.arg(category) OR sqlc.arg(category) = '')
ORDER BY name ASC;

-- name: GetCustomWeaponBySlug :one
SELECT * FROM custom_weapons
WHERE slug = ?;

-- name: DeleteCustomWeapon :exec
DELETE FROM custom_weapons
WHERE slug = ?;

-- name: CreateCustomWeapon :one
INSERT INTO custom_weapons (
  slug, name, tier, type, description, category, trait,
  weapon_range, damage, damage_type, burden, feature
)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
RETURNING *;

-- name: UpdateCustomWeapon :one
UPDATE custom_weapons SET
  name = ?,
  tier = ?,
  type = ?,
  description = ?,
  category = ?,
  trait = ?,
  weapon_range = ?,
  damage = ?,
  damage_type = ?,
  burden = ?,
  feature = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE slug = ?
RETURNING *;
