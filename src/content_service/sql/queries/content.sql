-- name: AddContent :one
INSERT INTO content (id, created_at, modified_at, user_id, title, description, type, s3_key) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetContentById :one
SELECT * FROM content WHERE id=$1 FOR UPDATE NOWAIT;

-- name: GetUserContent :many
SELECT (id, created_at, title, description, type) FROM content WHERE user_id=$1 LIMIT $2 OFFSET $3;

-- name: GetContentList :many
SELECT (id, created_at, title, description, type) FROM content LIMIT $1 OFFSET $2;

-- name: UpdateContentDetails :one
UPDATE content SET title=$1, description=$2, type=$3, modified_at=$4
WHERE id=$5
RETURNING *;

-- name: UpdateS3Key :exec
UPDATE content SET s3_key=$1, modified_at=$2
WHERE id=$3;

-- name: DeleteContent :exec
DELETE FROM content WHERE id=$1 AND user_id=$2;