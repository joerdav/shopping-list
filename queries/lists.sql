-- name: CreateList :exec
INSERT INTO lists (id, created_date, user_id)
VALUES (?, ?, ?);

-- name: UpdateList :one
UPDATE lists
set bought = ?
WHERE id = ?
RETURNING *;

-- name: GetList :one
SELECT * FROM lists
WHERE id = ?
LIMIT 1;

-- name: GetAllLists :many
SELECT * FROM lists
WHERE user_id = ?
ORDER BY created_date DESC;

-- name: DeleteList :exec
DELETE FROM lists
WHERE id = ?;

-- name: SetItem :exec
INSERT OR REPLACE INTO list_items (item_id, list_id, quantity)
VALUES (?, ?, ?);

-- name: GetItemsByList :many
SELECT * FROM list_items
WHERE list_id = ?;

-- name: DeleteItemsByList :exec
DELETE FROM list_items
WHERE list_id = ?;

-- name: DeleteItem :exec
DELETE FROM list_items
where item_id = ? AND list_id = ?;

-- name: SetRecipe :exec
INSERT OR REPLACE INTO list_recipes (recipe_id, list_id, quantity)
VALUES (?, ?, ?);

-- name: DeleteRecipe :exec
DELETE FROM list_recipes
where recipe_id = ? AND list_id = ?;

-- name: DeleteRecipesByList :exec
DELETE FROM list_recipes
WHERE list_id = ?;

-- name: GetRecipesByList :many
SELECT * FROM list_recipes
WHERE list_id = ?;
