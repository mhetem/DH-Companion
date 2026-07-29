-- name: CreateInventoryItem :one
INSERT INTO inventory_items (character_id, name, kind, qty, equipped, detail)
VALUES (?,?,?,?,?,?)
RETURNING *;

-- name: GetInventoryItem :one
SELECT * FROM inventory_items
WHERE id = ?;

-- name: ListInventoryItems :many
SELECT * FROM inventory_items
WHERE character_id = ?
ORDER BY equipped DESC, name ASC;

-- name: UpdateInventoryItem :one
UPDATE inventory_items SET
  name = sqlc.arg(name),
  kind = sqlc.arg(kind),
  qty = sqlc.arg(qty),
  equipped = sqlc.arg(equipped),
  detail = sqlc.arg(detail),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustInventoryQty :one
UPDATE inventory_items SET
  qty = max(0, min(999, qty + CAST(sqlc.arg(delta) AS INTEGER))),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetInventoryEquipped :one
UPDATE inventory_items SET
  equipped = sqlc.arg(equipped),
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UnequipInventoryKind :exec
UPDATE inventory_items SET
  equipped = 0,
  updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE character_id = ? AND kind = ? AND equipped = 1;

-- name: DeleteInventoryItem :exec
DELETE FROM inventory_items
WHERE id = ?;
