-- name: ShowAllEncounters :many
SELECT * FROM encounters
ORDER BY updated_at DESC;

-- name: CreateEncounter :one
INSERT INTO encounters (
  encounter_name, adversaries, custom_adversaries, environment_slug, party_id
)
VALUES (?,?,?,?,?)
RETURNING *;

-- name: DeleteEncounter :exec
DELETE FROM encounters
WHERE id = ?;

-- name: UpdateEncounter :one
UPDATE encounters SET
  encounter_name = ?,
  adversaries = ?,
  custom_adversaries = ?,
  environment_slug = ?,
  party_id = ?,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?
RETURNING *;

-- name: GetEncounter :one
SELECT * FROM encounters
WHERE id = ?;
