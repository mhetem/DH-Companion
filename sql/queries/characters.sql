-- name: CreateCharacter :one
INSERT INTO characters (
  name, pronouns, class_slug, subclass_slug, ancestry_slug, community_slug,
  agility, strength, finesse, instinct, presence, knowledge,
  hp_max, stress_max, hope, evasion, armor_score,
  threshold_major, threshold_severe, background, connections
)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
RETURNING *;

-- name: GetCharacter :one
SELECT * FROM characters
WHERE id = ?;

-- name: ListCharacters :many
SELECT * FROM characters
ORDER BY updated_at DESC;

-- name: UpdateCharacter :one
UPDATE characters SET
  name = sqlc.arg(name),
  pronouns = sqlc.arg(pronouns),
  class_slug = sqlc.arg(class_slug),
  subclass_slug = sqlc.arg(subclass_slug),
  ancestry_slug = sqlc.arg(ancestry_slug),
  community_slug = sqlc.arg(community_slug),
  agility = sqlc.arg(agility),
  strength = sqlc.arg(strength),
  finesse = sqlc.arg(finesse),
  instinct = sqlc.arg(instinct),
  presence = sqlc.arg(presence),
  knowledge = sqlc.arg(knowledge),
  hp_max = sqlc.arg(hp_max),
  hp_marked = min(hp_marked, sqlc.arg(hp_max)),
  stress_max = sqlc.arg(stress_max),
  stress_marked = min(stress_marked, sqlc.arg(stress_max)),
  evasion = sqlc.arg(evasion),
  armor_score = sqlc.arg(armor_score),
  armor_marked = min(armor_marked, sqlc.arg(armor_score)),
  threshold_major = sqlc.arg(threshold_major),
  threshold_severe = sqlc.arg(threshold_severe),
  background = sqlc.arg(background),
  connections = sqlc.arg(connections),
  notes = sqlc.arg(notes),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteCharacter :exec
DELETE FROM characters
WHERE id = ?;

-- name: AdjustCharacterHP :one
UPDATE characters SET
  hp_marked = max(0, min(hp_max, hp_marked + CAST(sqlc.arg(delta) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustCharacterStress :one
UPDATE characters SET
  stress_marked = max(0, min(stress_max, stress_marked + CAST(sqlc.arg(delta) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustCharacterHope :one
UPDATE characters SET
  hope = max(0, min(6, hope + CAST(sqlc.arg(delta) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustCharacterArmor :one
UPDATE characters SET
  armor_marked = max(0, min(armor_score, armor_marked + CAST(sqlc.arg(delta) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetCharacterVitals :one
UPDATE characters SET
  hp_marked = max(0, min(hp_max, CAST(sqlc.arg(hp_marked) AS INTEGER))),
  stress_marked = max(0, min(stress_max, CAST(sqlc.arg(stress_marked) AS INTEGER))),
  hope = max(0, min(6, CAST(sqlc.arg(hope) AS INTEGER))),
  armor_marked = max(0, min(armor_score, CAST(sqlc.arg(armor_marked) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetCharacterGold :one
UPDATE characters SET
  gold_handfuls = max(0, min(9, CAST(sqlc.arg(gold_handfuls) AS INTEGER))),
  gold_bags = max(0, min(9, CAST(sqlc.arg(gold_bags) AS INTEGER))),
  gold_chests = max(0, min(99, CAST(sqlc.arg(gold_chests) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetCharacterBeastform :one
UPDATE characters SET
  beastform_slug = sqlc.arg(beastform_slug),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateCharacterProgress :one
UPDATE characters SET
  level = sqlc.arg(level),
  proficiency = sqlc.arg(proficiency),
  subclass_mastery = sqlc.arg(subclass_mastery),
  multiclass_slug = sqlc.arg(multiclass_slug),
  multiclass_subclass_slug = sqlc.arg(multiclass_subclass_slug),
  agility = sqlc.arg(agility),
  strength = sqlc.arg(strength),
  finesse = sqlc.arg(finesse),
  instinct = sqlc.arg(instinct),
  presence = sqlc.arg(presence),
  knowledge = sqlc.arg(knowledge),
  marked_traits = sqlc.arg(marked_traits),
  hp_max = sqlc.arg(hp_max),
  stress_max = sqlc.arg(stress_max),
  evasion = sqlc.arg(evasion),
  threshold_major = sqlc.arg(threshold_major),
  threshold_severe = sqlc.arg(threshold_severe),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;
