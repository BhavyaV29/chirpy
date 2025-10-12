-- name: UpgradeChirpyRed :one
UPDATE users
SET is_chirpy_red=TRUE
WHERE id=$1
RETURNING is_chirpy_red;