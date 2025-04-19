-- name: GetJob :one
SELECT sqlc.embed(v), sqlc.embed(j) FROM job j JOIN vm v ON j.vm_id = v.id WHERE j.id = $1;

-- name: GetJobs :many
SELECT * FROM job;

-- name: InsertVm :one
INSERT INTO vm (name, memory, cpu, disk, image, port) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: InsertJob :one
INSERT INTO job (vm_id, status, base_path) VALUES ($1, $2, $3) RETURNING *;
