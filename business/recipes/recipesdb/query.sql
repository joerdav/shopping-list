-- name: CreateRecipe :exec
INSERT INTO recipes (id, name, user_id)
VALUES (?, ?, ?);

-- name: GetRecipe :one
SELECT * FROM recipes
WHERE id = ? LIMIT 1;

-- name: ListRecipes :many
SELECT * FROM recipes
WHERE user_id = ?
ORDER BY name;

-- name: SetIngredient :exec
INSERT OR REPLACE INTO ingredients (item_id, recipe_id, quantity)
VALUES (?, ?, ?);

-- name: DeleteIngredient :exec
DELETE FROM ingredients
WHERE item_id = ? AND recipe_id = ?;

-- name: ListIngredientsByRecipe :many
SELECT * FROM ingredients
WHERE recipe_id = ?;
