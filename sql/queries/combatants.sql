-- name: CreateCombatant :one
INSERT INTO combatants (
  combat_id, adversary_slug, display_name, hp_max, stress_max
)
VALUES (?,?,?,?,?)
RETURNING *;

-- name: GetCombatant :one
SELECT * FROM combatants
WHERE id = ?;

-- name: ListCombatants :many
SELECT * FROM combatants
WHERE combat_id = ?
ORDER BY id ASC;

-- name: AdjustCombatantHP :one
UPDATE combatants SET
  hp_marked = max(0, min(hp_max, hp_marked + sqlc.arg(delta))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustCombatantStress :one
UPDATE combatants SET
  stress_marked = max(0, min(stress_max, stress_marked + sqlc.arg(delta))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetCombatantVitals :one
UPDATE combatants SET
  hp_marked = max(0, min(hp_max, CAST(sqlc.arg(hp_marked) AS INTEGER))),
  stress_marked = max(0, min(stress_max, CAST(sqlc.arg(stress_marked) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetCombatantSpotlight :one
UPDATE combatants SET
  spotlight = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: ClearCombatantSpotlights :exec
UPDATE combatants SET
  spotlight = 0,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE combat_id = ? AND spotlight = 1;

-- name: UpdateCombatant :one
UPDATE combatants SET
  display_name = sqlc.arg(display_name),
  hp_max = sqlc.arg(hp_max),
  stress_max = sqlc.arg(stress_max),
  hp_marked = min(hp_marked, sqlc.arg(hp_max)),
  stress_marked = min(stress_marked, sqlc.arg(stress_max)),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteCombatant :exec
DELETE FROM combatants
WHERE id = ?;

-- name: DeleteCombatantsForCombat :exec
DELETE FROM combatants
WHERE combat_id = ?;
