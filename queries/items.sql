-- name: CreateItem :exec
INSERT INTO items (id, name, shop_id, user_id)
VALUES (?, ?, ?, ?);

-- name: GetItem :one
SELECT * FROM items
WHERE id = ? LIMIT 1;

-- name: ListItems :many
SELECT * FROM items
WHERE user_id = ?
ORDER BY name;

-- name: ListItemsByShop :many
SELECT * FROM items
WHERE shop_id = ?
ORDER BY name;

-- name: UpdateItem :one
UPDATE items
set name = ?,
shop_id = ?
WHERE id = ?
RETURNING *;
