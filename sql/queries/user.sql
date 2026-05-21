-- name: CreateUser :one
INSERT INTO users (
    name,
    email,
    password,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    COALESCE($10::jsonb, '{}'::jsonb)
)
RETURNING 
    id,
    name,
    email,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links,
    created_at,
    updated_at;

-- name: GetUser :one
SELECT 
    id,
    name,
    email,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links,
    created_at,
    updated_at
FROM users
WHERE id = $1;


-- name: GetUserByEmail :one
SELECT 
    id,
    name,
    email,
    password,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links,
    created_at,
    updated_at
FROM users
WHERE email = $1;


-- name: PatchUserProfile :one
UPDATE users
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    email = COALESCE(sqlc.narg('email')::text, email),
    batch = COALESCE(sqlc.narg('batch')::text, batch),
    description = COALESCE(sqlc.narg('description')::text, description),
    social_links = COALESCE(sqlc.narg('social_links')::jsonb, social_links),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING
    id,
    name,
    email,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links,
    created_at,
    updated_at;


-- name: PatchUserAdmin :one
UPDATE users
SET
    status = COALESCE($2, status),
    role = COALESCE($3, role),
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    name,
    email,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links,
    created_at,
    updated_at;


-- name: PatchUserPassword :one
UPDATE users
SET
    password = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    updated_at;


-- name: PatchUserImages :one
UPDATE users
SET
    image = COALESCE($2, image),
    banner_image = COALESCE($3, banner_image),
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    image,
    banner_image,
    description,
    social_links,
    updated_at;
