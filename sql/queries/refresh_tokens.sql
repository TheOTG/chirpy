-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4
);

-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 AND expires_at > NOW() AND revoked_at IS NULL
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = NOW(), revoked_at = NOW()
WHERE token = $1;

-- name: DeleteAllTokens :exec
DELETE FROM refresh_tokens;