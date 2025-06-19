-- name: GetSingleChirp :one
SELECT * FROM chirps WHERE ID=$1;
