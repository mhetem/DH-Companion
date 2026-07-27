-- name: CreateCountdown :one
INSERT INTO countdowns (name, value, max, kind)
VALUES (?,?,?,?)
RETURNING *;

-- name: GetCountdown :one
SELECT * FROM countdowns
WHERE id = ?;

-- name: ListCountdowns :many
SELECT * FROM countdowns
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
  value = max(0, min(sqlc.arg(max), sqlc.arg(value))),
  kind = sqlc.arg(kind),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteCountdown :exec
DELETE FROM countdowns
WHERE id = ?;
