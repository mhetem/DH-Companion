-- name: CreateSession :one
INSERT INTO sessions (campaign_id, number, title, date, recap)
VALUES (?,?,?,?,?)
RETURNING *;

-- name: NextSessionNumber :one
SELECT coalesce(max(number), 0) + 1 AS next_number FROM sessions
WHERE campaign_id = ?;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = ?;

-- name: ListSessionsForCampaign :many
SELECT * FROM sessions
WHERE campaign_id = ?
ORDER BY number DESC;

-- name: UpdateSession :one
UPDATE sessions SET
  number = ?,
  title = ?,
  date = ?,
  recap = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = ?;
