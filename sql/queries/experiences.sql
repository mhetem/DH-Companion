-- name: CreateExperience :one
INSERT INTO experiences (character_id, name, modifier)
VALUES (?,?,?)
RETURNING *;

-- name: GetExperience :one
SELECT * FROM experiences
WHERE id = ?;

-- name: ListExperiences :many
SELECT * FROM experiences
WHERE character_id = ?
ORDER BY id ASC;

-- name: UpdateExperience :one
UPDATE experiences SET
  name = sqlc.arg(name),
  modifier = sqlc.arg(modifier),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustExperienceModifier :one
UPDATE experiences SET
  modifier = max(0, min(9, modifier + CAST(sqlc.arg(delta) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteExperience :exec
DELETE FROM experiences
WHERE id = ?;
