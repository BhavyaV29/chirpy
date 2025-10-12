-- name: NewPasswordEmail :one
UPDATE users
SET email=$1,hashed_password=$2 
WHERE id=$3
RETURNING id,email,created_at,updated_at,is_chirpy_red;