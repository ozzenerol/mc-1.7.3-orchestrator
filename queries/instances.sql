-- name: CreateInstance :one
INSERT INTO instances 
(host_id, port, max_players, dedicated_ram_mb, log_path)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetInstance :one
SELECT * FROM instances
WHERE id = $1;

-- name: ListInstances :many
SELECT * FROM instances
ORDER BY created_at DESC;

-- name: UpdateInstance :one
UPDATE instances
SET 
host_id             = $2,
port                = $3,
max_players         = $4,
dedicated_ram_mb    = $5,
log_path            = $6
WHERE id            = $1
RETURNING *;

-- name: DeleteInstance :execrows
DELETE FROM instances
WHERE id = $1;
