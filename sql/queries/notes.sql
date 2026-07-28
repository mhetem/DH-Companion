-- name: CreateNote :one
INSERT INTO notes (campaign_id, kind, title, body)
VALUES (?,?,?,?)
RETURNING *;

-- name: GetNote :one
SELECT * FROM notes
WHERE id = ?;

-- name: ListNotesForCampaign :many
SELECT * FROM notes
WHERE campaign_id = ?
ORDER BY kind ASC, title ASC;

-- name: ListNotesForCampaignByKind :many
SELECT * FROM notes
WHERE campaign_id = ? AND kind = ?
ORDER BY title ASC;

-- name: UpdateNote :one
UPDATE notes SET
  kind = ?,
  title = ?,
  body = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: DeleteNote :exec
DELETE FROM notes
WHERE id = ?;
