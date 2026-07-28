-- name: CreateCombat :one
INSERT INTO combats (encounter_id, campaign_id, session_id)
VALUES (?,?,?)
RETURNING *;

-- name: GetCombat :one
SELECT * FROM combats
WHERE id = ?;

-- name: ListCombats :many
SELECT * FROM combats
ORDER BY created_at DESC;

-- name: GetActiveCombat :one
SELECT * FROM combats
WHERE active = 1
ORDER BY updated_at DESC
LIMIT 1;

-- name: ListCombatsForEncounter :many
SELECT * FROM combats
WHERE encounter_id = ?
ORDER BY created_at DESC;

-- name: ListCombatsForCampaign :many
SELECT * FROM combats
WHERE campaign_id = ?
ORDER BY created_at DESC;

-- name: ListCombatsForSession :many
SELECT * FROM combats
WHERE session_id = ?
ORDER BY created_at ASC;

-- name: SetCombatLinks :one
UPDATE combats SET
  campaign_id = sqlc.arg(campaign_id),
  session_id = sqlc.arg(session_id),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustFear :one
UPDATE combats SET
  fear = max(0, min(12, fear + sqlc.arg(delta))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetFear :one
UPDATE combats SET
  fear = max(0, min(12, CAST(sqlc.arg(fear) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetCombatActive :one
UPDATE combats SET
  active = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: DeactivateAllCombats :exec
UPDATE combats SET
  active = 0,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE active = 1;

-- name: DeleteCombat :exec
DELETE FROM combats
WHERE id = ?;
