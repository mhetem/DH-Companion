-- name: CreateCampaign :one
INSERT INTO campaigns (name, description)
VALUES (?,?)
RETURNING *;

-- name: GetCampaign :one
SELECT * FROM campaigns
WHERE id = ?;

-- name: ListCampaigns :many
SELECT * FROM campaigns
ORDER BY updated_at DESC;

-- name: UpdateCampaign :one
UPDATE campaigns SET
  name = ?,
  description = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: AdjustCampaignFear :one
UPDATE campaigns SET
  current_fear = max(0, min(12, current_fear + CAST(sqlc.arg(delta) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetCampaignFear :one
UPDATE campaigns SET
  current_fear = max(0, min(12, CAST(sqlc.arg(fear) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteCampaign :exec
DELETE FROM campaigns
WHERE id = ?;
