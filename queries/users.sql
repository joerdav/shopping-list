-- name: CreateUser :exec
INSERT INTO users (id)
VALUES (?);

-- name: GetUser :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

