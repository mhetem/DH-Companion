-- +goose Up
CREATE TABLE session_encounters (
    session_id   INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    encounter_id INTEGER NOT NULL REFERENCES encounters(id) ON DELETE CASCADE,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    PRIMARY KEY (session_id, encounter_id)
) STRICT;

CREATE INDEX idx_session_encounters_encounter ON session_encounters (encounter_id);

-- +goose Down
DROP TABLE session_encounters;
