-- name: CreateUser :one
INSERT INTO users (
    id, email, password_hash, is_mfa_enabled, role, status,
    first_name, last_name, phone, address, city, state, zip, country,
    created_at, updated_at, last_login_at, deleted_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: UpdateUser :exec
UPDATE users SET email = $2, password_hash = $3, is_mfa_enabled = $4, role = $5, status = $6,
    first_name = $7, last_name = $8, phone = $9, address = $10, city = $11, state = $12, zip = $13, country = $14,
    updated_at = $15, last_login_at = $16
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL;
