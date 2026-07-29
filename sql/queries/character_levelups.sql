-- name: CreateLevelUp :one
INSERT INTO character_levelups (character_id, level, tier, choices, summary)
VALUES (?,?,?,?,?)
RETURNING *;

-- name: ListLevelUps :many
SELECT * FROM character_levelups
WHERE character_id = ?
ORDER BY level ASC;

-- name: ListLevelUpsForTier :many
SELECT * FROM character_levelups
WHERE character_id = ? AND tier = ?
ORDER BY level ASC;
