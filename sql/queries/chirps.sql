-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ListChirps :many
SELECT * FROM chirps
WHERE user_id = sqlc.narg('user_id') OR sqlc.narg('user_id') IS NULL
ORDER BY created_at;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1 LIMIT 1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;

-- name: DeleteAllChirps :exec
DELETE FROM chirps;