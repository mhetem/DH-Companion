-- name: CreateWorldNote :one
INSERT INTO world_notes (campaign_id, parent_id, kind, title, world_date, body, position)
VALUES (?,?,?,?,?,?,?)
RETURNING *;

-- name: NextWorldNotePosition :one
SELECT coalesce(max(position), 0) + 1 AS next_position FROM world_notes
WHERE campaign_id = ?;

-- name: GetWorldNote :one
SELECT * FROM world_notes
WHERE id = ?;

-- name: ListWorldNotesForCampaign :many
SELECT * FROM world_notes
WHERE campaign_id = ?
ORDER BY position ASC, id ASC;

-- name: UpdateWorldNote :one
UPDATE world_notes SET
  kind = ?,
  title = ?,
  world_date = ?,
  body = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: MoveWorldNote :one
UPDATE world_notes SET
  parent_id = ?,
  position = ?
WHERE id = ?
RETURNING *;

-- name: SetWorldNotePosition :exec
UPDATE world_notes SET
  position = ?
WHERE id = ?;

-- name: DeleteWorldNote :exec
DELETE FROM world_notes
WHERE id = ?;
