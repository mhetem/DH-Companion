-- name: CreateHandout :one
INSERT INTO handouts (campaign_id, kind, title, body, media, media_name)
VALUES (?,?,?,?,?,?)
RETURNING *;

-- name: GetHandout :one
SELECT * FROM handouts
WHERE id = ?;

-- name: ListHandoutsForCampaign :many
SELECT * FROM handouts
WHERE campaign_id = ?
ORDER BY kind ASC, title ASC, id ASC;

-- name: UpdateHandout :one
UPDATE handouts SET
  kind = ?,
  title = ?,
  body = ?,
  media = ?,
  media_name = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: SetHandoutView :exec
UPDATE handouts SET
  view_rotate = ?,
  view_zoom = ?,
  view_x = ?,
  view_y = ?
WHERE id = ?;

-- name: DeleteHandout :exec
DELETE FROM handouts
WHERE id = ?;

-- name: CountHandoutsUsingMedia :one
SELECT count(*) FROM handouts
WHERE media = ?;

-- name: ListHandoutMediaForCampaign :many
SELECT DISTINCT media FROM handouts
WHERE campaign_id = ? AND media <> '';
