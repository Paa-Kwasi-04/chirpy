-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token,created_at,updated_at,user_id,expires_at)
VALUES
    (
        $1,
        $2,
        $3,
        $4,
        $5
    )
RETURNING *;    

-- name: GetUserFromRefreshToken :one
SELECT users.id,users.email,
refresh_tokens.token,refresh_tokens.expires_at,refresh_tokens.revoked_at
FROM refresh_tokens
INNER JOIN users ON refresh_tokens.user_id = users.id
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $1,updated_at = $1
WHERE token = $2;