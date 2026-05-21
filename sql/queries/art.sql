-- name: CreateArt :one
INSERT INTO art (id, name, description, image, tags, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;


-- name: GetArtByID :one
SELECT *
FROM art
WHERE id = $1;


-- name: GetArtByUser :many
SELECT * FROM art
WHERE user_id = $1
ORDER BY created_at DESC;


-- name: ListPendingArt :many
SELECT * FROM art
WHERE status = 'pending'
ORDER BY created_at DESC;
-- name: ListArt :many
SELECT * FROM art
WHERE status = 'approved'
ORDER BY created_at DESC;


-- name: ListArtByTag :many
SELECT * FROM art
WHERE status = 'approved'
  AND $1 = ANY(tags)
ORDER BY created_at DESC;


-- name: ListArtByTags :many
SELECT * FROM art
WHERE status = 'approved'
  AND tags && $1::text[]
ORDER BY created_at DESC;


-- name: UpdateArt :one
UPDATE art
SET
    name        = COALESCE($2, name),
    description = COALESCE($3, description),
    tags        = COALESCE($4, tags),
    updated_at  = NOW()
WHERE id = $1 AND user_id = $5
RETURNING *;

-- name: UpdateArtStatus :one
UPDATE art
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteArt :one
DELETE FROM art
WHERE id = $1 AND user_id = $2
RETURNING id;
