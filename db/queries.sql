-- name: GetJob :one
SELECT * FROM job WHERE id = $1;

-- name: GetJobs :many
SELECT * FROM job;
