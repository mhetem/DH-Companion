-- name: LinkSessionEncounter :exec
INSERT OR IGNORE INTO session_encounters (session_id, encounter_id)
VALUES (?,?);

-- name: UnlinkSessionEncounter :exec
DELETE FROM session_encounters
WHERE session_id = ? AND encounter_id = ?;

-- name: ListEncountersForSession :many
SELECT encounters.* FROM encounters
JOIN session_encounters ON session_encounters.encounter_id = encounters.id
WHERE session_encounters.session_id = ?
ORDER BY session_encounters.created_at ASC;

-- name: ListSessionsForEncounter :many
SELECT sessions.* FROM sessions
JOIN session_encounters ON session_encounters.session_id = sessions.id
WHERE session_encounters.encounter_id = ?
ORDER BY sessions.number DESC;
