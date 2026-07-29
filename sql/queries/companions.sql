-- name: CreateCompanion :one
INSERT INTO companions (
  character_id, name, evasion, damage_die, attack_range, attack,
  stress_max, experiences, upgrades, notes
)
VALUES (?,?,?,?,?,?,?,?,?,?)
RETURNING *;

-- name: GetCompanion :one
SELECT * FROM companions
WHERE character_id = ?;

-- name: UpdateCompanion :one
UPDATE companions SET
  name = sqlc.arg(name),
  evasion = sqlc.arg(evasion),
  damage_die = sqlc.arg(damage_die),
  attack_range = sqlc.arg(attack_range),
  attack = sqlc.arg(attack),
  stress_max = sqlc.arg(stress_max),
  stress_marked = min(stress_marked, sqlc.arg(stress_max)),
  experiences = sqlc.arg(experiences),
  upgrades = sqlc.arg(upgrades),
  notes = sqlc.arg(notes),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE character_id = sqlc.arg(character_id)
RETURNING *;

-- name: AdjustCompanionStress :one
UPDATE companions SET
  stress_marked = max(0, min(stress_max, stress_marked + CAST(sqlc.arg(delta) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE character_id = sqlc.arg(character_id)
RETURNING *;

-- name: SetCompanionStress :one
UPDATE companions SET
  stress_marked = max(0, min(stress_max, CAST(sqlc.arg(stress_marked) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE character_id = sqlc.arg(character_id)
RETURNING *;

-- name: DeleteCompanion :exec
DELETE FROM companions
WHERE character_id = ?;
