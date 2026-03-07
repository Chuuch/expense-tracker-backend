-- name: CreateUser :one
INSERT INTO users (
    id, email, password_hash, is_mfa_enabled, role, status,
    username, phone, address, city, state, zip, country,
    verification_code, verification_code_expires_at,
    created_at, updated_at, last_login_at, deleted_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12, $13,
    $14, $15, $16, $17, $18, $19
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: UpdateUser :exec
UPDATE users SET email = $2, password_hash = $3, is_mfa_enabled = $4, role = $5, status = $6,
    username = $7, phone = $8, address = $9, city = $10, state = $11, zip = $12, country = $13,
    verification_code = $14, verification_code_expires_at = $15,
    updated_at = $16, last_login_at = $17
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL;
