-- name: GetShop :one
SELECT * FROM shops
WHERE id = ? LIMIT 1;

-- name: ListShops :many
SELECT * FROM shops
WHERE user_id = ?
ORDER BY name;

-- name: CreateShop :exec
INSERT INTO shops (id, name, user_id)
VALUES (?, ?, ?);

-- name: UpdateShop :exec
UPDATE shops
SET name = ?
WHERE id = ?;

-- name: DeleteShop :exec
DELETE FROM shops
WHERE id = ?;
