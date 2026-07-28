-- name: CreateCountdown :one
INSERT INTO countdowns (name, value, max, kind, campaign_id)
VALUES (?,?,?,?,?)
RETURNING *;

-- name: GetCountdown :one
SELECT * FROM countdowns
WHERE id = ?;

-- name: ListCountdowns :many
SELECT * FROM countdowns
ORDER BY created_at ASC;

-- name: ListCountdownsForCampaign :many
SELECT * FROM countdowns
WHERE campaign_id = ?
ORDER BY created_at ASC;

-- name: ListUnassignedCountdowns :many
SELECT * FROM countdowns
WHERE campaign_id IS NULL
ORDER BY created_at ASC;

-- name: AdjustCountdown :one
UPDATE countdowns SET
  value = max(0, min("max", value + sqlc.arg(delta))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateCountdown :one
UPDATE countdowns SET
  name = sqlc.arg(name),
  "max" = sqlc.arg(max),
  value = max(0, min(sqlc.arg(max), CAST(sqlc.arg(value) AS INTEGER))),
  kind = sqlc.arg(kind),
  campaign_id = sqlc.arg(campaign_id),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteCountdown :exec
DELETE FROM countdowns
WHERE id = ?;
