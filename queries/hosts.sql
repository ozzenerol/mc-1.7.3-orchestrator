-- name: CreateHost :one
INSERT INTO hosts (name, ip)
VALUES ($1, $2)
RETURNING *;

-- name: GetHost :one
SELECT * FROM hosts
WHERE id = $1;

-- name: ListHosts :many
SELECT * FROM hosts
ORDER BY created_at;

-- name: UpdateHost :one
UPDATE hosts
SET name = $2, ip = $3
WHERE id = $1
RETURNING *;

-- name: DeleteHost :execrows
DELETE FROM hosts
WHERE id = $1;
