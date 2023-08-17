-- name: AddUser :execresult
INSERT INTO users (name)
VALUES (?);
-- name: UpdateUser :exec
UPDATE users
SET name = ?
WHERE id = ?;
-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = ?
LIMIT 1;
-- name: ListUsers :many
SELECT *
FROM users;
-- EOF
