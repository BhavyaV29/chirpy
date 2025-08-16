-- name: GetUserByEmail :one
SELECT id,created_at,updated_at,email,hashed_password FROM users WHERE Email=$1;
